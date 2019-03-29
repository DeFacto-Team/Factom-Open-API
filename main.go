package main

import (
	"flag"

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
	go fetchUnsyncedChains(s)

	// Start API
	api := api.NewApi(conf, s)
	log.WithField("address", api.GetApiInfo().Address).
		WithField("mw", api.GetApiInfo().MW).
		Info("Starting api")
	log.Fatal(api.Start())

}

func fetchChainUpdates(s service.Service) {
	chains := s.GetChains(&model.Chain{Status: model.ChainCompleted})
	for _, c := range chains {
		s.ParseNewChainEntries(c)
	}
}

func fetchUnsyncedChains(s service.Service) {
	t := false
	chains := s.GetChains(&model.Chain{Synced: &t})
	for _, c := range chains {
		s.ParseAllChainEntries(c)
	}
}
