package main

import (
	"encoding/json"
	"flag"
	"time"

	"github.com/DeFacto-Team/Factom-Open-API/api"
	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/service"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"

	"github.com/FactomProject/factom"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

const (
	MinutesInBlock = 10
)

func main() {

	var err error

	// Get config flag if exists
	configFile := flag.String("c", "config/config.yaml", "Path to config file")
	flag.Parse()

	// Load config
	var conf *config.Config
	if conf, err = config.NewConfig(*configFile); err != nil {
		log.Fatal(err)
	}

	// Setup logger
	log.SetLevel(log.Level(conf.LogLevel))
	log.Info("Starting service with configuration: ", conf.ConfigFile)

	// Create store
	store, err := store.NewStore(conf)
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
		log.Error(err)
	}

	// Create services
	s := service.NewService(store, wallet)
	log.Info("Services created successfully")

	go fetchChainUpdates(s)
	//go fetchUnsyncedChains(s)
	//go processQueue(s)
	//go clearQueue(s)

	// Start API
	api := api.NewApi(conf, s)
	log.WithField("address", api.GetApiInfo().Address).
		WithField("mw", api.GetApiInfo().MW).
		Info("Starting api")
	log.Fatal(api.Start())

}

func fetchChainUpdates(s service.Service) {

	var currentMinute int    // current minute
	var currentMinuteEnd int // current minute after parsing ended
	var currentDBlock int    // current dblock
	var latestDBlock int     // latest fetched dblock
	var sleepFor int         // sleep timer

	for {

		log.Info("Updates parser: Iteration started")

		// get current minute & dblock from Factom
		currentMinute, currentDBlock = getMinuteAndHeight()
		log.Info("Updates parser: currentMinute=", currentMinute, ", currentDBlock=", currentDBlock)

		// if current dblock <= latest fetched dblock, then elections should occur and need to sleep 1 minute before next try
		// on the first iteration latestDblock = 0, so this code won't run & new updates will be fetched when API started
		for currentDBlock <= latestDBlock {
			log.Info("Updates parser: Sleeping for 1 minute / currentDBlock=", currentDBlock, "latestDBlock=", latestDBlock)
			time.Sleep(1 * time.Minute)
			currentMinute, currentDBlock = getMinuteAndHeight()
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
		currentMinuteEnd, _ = getMinuteAndHeight()
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

		log.Info("Updates parser: Sleeping for ", sleepFor, " minutes")
		time.Sleep(time.Duration(sleepFor) * time.Minute)
	}
}

func fetchUnsyncedChains(s service.Service) {
	t := false
	chains := s.GetChains(&model.Chain{Synced: &t})
	for _, c := range chains {
		err := s.ParseAllChainEntries(c)
		if err != nil {
			log.Error(err)
		}
	}
}

func processQueue(s service.Service) {
	for {
		log.Info("Processing queue")
		queue := s.GetQueueToProcess()
		for _, q := range queue {
			err := s.ProcessQueue(q)
			if err != nil {
				log.Error(err)
			}
		}
		time.Sleep(5 * time.Second)
	}
}

func clearQueue(s service.Service) {
	for {
		log.Info("Clearing queue")
		queue := s.GetQueueToClear()
		for _, q := range queue {
			s.ClearQueue(q)
		}
		time.Sleep(60 * time.Second)
	}
}

func getMinuteAndHeight() (int, int) {

	var currentMinute float64
	var dBlockHeight float64
	var i interface{}

	request := factom.NewJSON2Request("current-minute", 0, nil)

	resp, err := factom.SendFactomdRequest(request)
	if err != nil {
		log.Error(err)
	}

	if err = json.Unmarshal(resp.JSONResult(), &i); err != nil {
		log.Error(err)
	}

	m, _ := i.(map[string]interface{})
	currentMinute = m["minute"].(float64)
	dBlockHeight = m["directoryblockheight"].(float64)

	return int(currentMinute), int(dBlockHeight)

}
