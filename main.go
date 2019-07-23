package main

import (
	"encoding/json"
	"flag"
	"os/user"
	"time"

	"github.com/DeFacto-Team/Factom-Open-API/api"
	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/pool"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"

	"github.com/FactomProject/factom"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	// number of minutes in Factom dblock
	MinutesInBlock = 10
	// number of background workers to fetch data from chains
	WorkersCount = 4
)

// @title Factom Open API
// @version 1.0.0
// @description Open-source REST API for the Factom blockchain.

// @contact.name De Facto
// @contact.url https://de-facto.pro
// @contact.email team@de-facto.pro

// @license.name MIT
// @license.url https://github.com/DeFacto-Team/Factom-Open-API/blob/master/LICENSE

// @host localhost:8081
// @BasePath /v1

// @securityDefinitions.apikey ApiKeyAuth
// @in header
// @name Authorization
func main() {

	var err error

	usr, err := user.Current()
	if err != nil {
		log.Error(err)
	}

	configFile := usr.HomeDir + "/.foa/config.yaml"

	flag.StringVar(&configFile, "c", configFile, "config.yaml path")
	flag.Parse()

	startAPI(configFile)

}

func startAPI(configFile string) {

	for {

		var err error
		var conf *config.Config

		if conf, err = config.NewConfig(configFile); err != nil {
			log.Fatal(err)
		}

		// Setup logger
		log.SetLevel(log.Level(conf.API.LogLevel))
		log.Info("Starting Factom Open API")

		// Create store
		store, err := store.NewStore(conf, true)
		if err != nil {
			log.Error("Database connection FAILED")
			log.Fatal(err)
		}
		defer store.Close()
		log.Info("Store created successfully")

		// Create factom
		if conf.Factom.URL != "" {
			factom.SetFactomdServer(conf.Factom.URL)
		}
		if conf.Factom.User != "" && conf.Factom.Password != "" {
			factom.SetFactomdRpcConfig(conf.Factom.User, conf.Factom.Password)
		}

		// Check factomd availability
		heights, err := factom.GetHeights()
		if err != nil {
			log.Warn("FAILED connection to factomd node: ", conf.Factom.URL)
		} else {
			log.Info("Using factomd node: ", conf.Factom.URL,
				" (DBlock=", heights.DirectoryBlockHeight, "/", heights.LeaderHeight,
				", EntryBlock=", heights.EntryHeight, "/", heights.EntryBlockHeight, ")")
			if heights.EntryBlockHeight-heights.EntryHeight > 1 {
				log.Warn("Factomd node is not fully synced! API will not be able to write data on the blockchain or read actual data from the blockchain!")
			}
		}

		// initialize wallet
		wallet, err := wallet.NewWallet(conf)
		if err != nil {
			log.Warn(err)
			log.Warn("You need to setup Es address in order to use API")
		}

		// Create services
		s := service.NewService(store, wallet)
		log.Info("Services created successfully")

		// Initialize pool for history fetching chains
		collector := pool.StartDispatcher(WorkersCount)

		// Initialize single-thread background workers
		die := make(chan bool)
		go pingDB(store, die)
		go fetchUnsyncedChains(s, collector, die)
		go fetchChainUpdates(s, die)
		go processQueue(s, die)
		go clearQueue(s, die)

		// Init REST API
		api := api.NewAPI(conf, s, configFile)

		// Start REST API
		log.WithField("port", api.GetAPIInfo().Port).
			WithField("version", api.GetAPIInfo().Version).
			WithField("middleware", api.GetAPIInfo().MW).
			Info("Starting REST API")
		err = api.Start()

		if err != nil {
			log.Warn("Restarting Factom Open API")
			close(die)
		}

	}

}

func fetchUnsyncedChains(s service.Service, collector pool.Collector, die chan bool) {

	log.Info("Reseting all unsynced local chains to put it into pool")
	err := s.ResetChainsParsingAtAPIStart()
	if err != nil {
		log.Error(err)
	}

	for {
		select {
		default:
			log.Info("Fetching unsynced chains: iteration started")
			t := false
			chains := s.GetChains(&model.Chain{Synced: &t, WorkerID: -1, SentToPool: &t})
			for _, c := range chains {
				s.SetChainSentToPool(c)
				collector.Work <- pool.Work{ID: c.ChainID, Job: c, Service: s}
			}
			time.Sleep(5 * time.Second)
		case <-die:
			return
		}
	}

}

func fetchChainUpdates(s service.Service, die chan bool) {

	var currentMinute int    // current minute
	var currentMinuteEnd int // current minute after parsing ended
	var currentDBlock int    // current dblock
	var latestDBlock int     // latest fetched dblock
	var sleepFor int         // sleep timer
	var err error

	for {

		select {
		default:
			log.Info("Updates parser: Iteration started")

			// get current minute & dblock from Factom
			currentMinute, currentDBlock, err = getMinuteAndHeight()
			if err != nil {
				continue
			}
			log.Info("Updates parser: currentMinute=", currentMinute, ", currentDBlock=", currentDBlock)

			// if current dblock <= latest fetched dblock, then elections should occur and need to sleep 1 minute before next try
			// on the first iteration latestDblock = 0, so this code won't run & new updates will be fetched when API started
			for currentDBlock <= latestDBlock {
				log.Info("Updates parser: Sleeping for 1 minute / currentDBlock=", currentDBlock, ", latestDBlock=", latestDBlock)
				time.Sleep(1 * time.Minute)
				currentMinute, currentDBlock, err = getMinuteAndHeight()
				log.Info("Updates parser: currentMinute=", currentMinute, ", currentDBlock=", currentDBlock)
			}

			// if we are here, then latestDBlock > currentDBlock (i.e. new dblock appeared)
			// parsing chains updates
			chains := s.GetChains(&model.Chain{Status: model.ChainCompleted})
			for _, c := range chains {
				err := s.ParseNewChainEntries(c)
				if err != nil {
					log.Error(err)
				}
			}

			// updating latest parsed dblock
			latestDBlock = currentDBlock

			// parsing may spend time, so check current minute
			currentMinuteEnd, _, err = getMinuteAndHeight()
			log.Debug("Updates parser: currentMinute=", currentMinuteEnd)

			// if current minute was {8|9} and becomes {0|1|2|3…}, i.e. new block appeared during the parsing
			// then no sleep in the end
			if currentMinuteEnd < currentMinute {

				sleepFor = 0

			} else {
				// else calculate sleep minutes before next block

				// workaround: if minute == 0, then sleep for 1 minute instead of 11
				if currentMinute == 0 {
					currentMinute = 10
				}
				// + 1 needed for sleeping at least 1 minute
				sleepFor = MinutesInBlock - currentMinute + 1

			}

			log.Info("Updates parser: Sleeping for ", sleepFor, " minute(s)")
			time.Sleep(time.Duration(sleepFor) * time.Minute)
		case <-die:
			return
		}

	}
}

// Get all tasks from queue where processed_at == NULL
func processQueue(s service.Service, die chan bool) {
	for {
		select {
		default:
			log.Info("Processing queue: iteration started")
			queue := s.GetQueueToProcess()
			for _, q := range queue {
				err := s.ProcessQueue(q)
				if err != nil {
					log.Error(err)
				}
			}
			time.Sleep(5 * time.Second)
		case <-die:
			return
		}
	}
}

func clearQueue(s service.Service, die chan bool) {
	for {
		select {
		default:
			log.Info("Clearing queue: iteration started")
			queue := s.GetQueueToClear()
			for _, q := range queue {
				s.ClearQueue(q)
			}
			time.Sleep(60 * time.Second)
		case <-die:
			return
		}
	}
}

func getMinuteAndHeight() (int, int, error) {

	var currentMinute float64
	var dBlockHeight float64
	var i interface{}

	request := factom.NewJSON2Request("current-minute", 0, nil)

	resp, err := factom.SendFactomdRequest(request)
	if err != nil {
		log.Error(err)
		return 0, 0, nil
	}

	if err = json.Unmarshal(resp.JSONResult(), &i); err != nil {
		log.Error(err)
		return 0, 0, nil
	}

	m, _ := i.(map[string]interface{})
	currentMinute = m["minute"].(float64)
	dBlockHeight = m["directoryblockheight"].(float64)

	return int(currentMinute), int(dBlockHeight), nil

}

func pingDB(s store.Store, die chan bool) {
	for {
		select {
		default:
			if err := s.Ping(); err != nil {
				log.Error("Database connection FAILED")
				log.Fatal(err)
			}
			time.Sleep(5 * time.Second)
		case <-die:
			return
		}
	}
}
