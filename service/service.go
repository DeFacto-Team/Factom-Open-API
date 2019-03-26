package service

import (
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	"github.com/FactomProject/factom"
	//factom "github.com/ilzheev/factom"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	CreateUser(user *model.User) error
	GetUserByAccessToken(token string) *model.User

	GetChain(chain *model.Chain, user *model.User) *model.ChainWithLinks
	CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error)

	GetEntry(entry *model.Entry) *model.Entry
	CreateEntry(entry *model.Entry) (*model.Entry, error)

	parseBackChainEntries(chain *model.Chain) error
	parseBackChainEntriesFromEBlock(eblock string) error
}

func NewService(store store.Store, wallet wallet.Wallet) Service {
	return &ServiceContext{store: store, wallet: wallet}
}

type ServiceContext struct {
	store  store.Store
	wallet wallet.Wallet
}

func (c *ServiceContext) CreateUser(user *model.User) error {

	err := c.store.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (c *ServiceContext) GetUserByAccessToken(token string) *model.User {
	return c.store.GetUserByAccessToken(token)
}

func (c *ServiceContext) GetChain(chain *model.Chain, user *model.User) *model.ChainWithLinks {

	resp := &model.ChainWithLinks{}

	log.Debug("Search for chain into local DB")

	// search for chain.ChainID into local DB
	localChain := c.store.GetChain(chain)

	if localChain != nil {
		log.Debug("Chain " + chain.ChainID + " found into local DB")
		resp.Chain = localChain
		return resp
	}

	log.Debug("Chain " + chain.ChainID + " not found into local DB")
	log.Debug("Search for chain on the blockchain")

	if chain.Exists() {
		log.Debug("Chain " + chain.ChainID + " found on the blockchain")

		log.Debug("Getting chain status from the blockchain")
		chain.Status, _ = chain.GetStatusFromFactom()
		resp.Chain = chain

		log.Debug("Creating chain into local DB")
		err := c.store.CreateChain(chain)
		if err != nil {
			log.Error(err)
		}

		log.Debug("Start fetching chain entries into local DB")
		go c.parseBackChainEntries(chain)

		// If we are here, so no errors occured and we force bind chain to API user
		log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
		err = c.store.BindChainToUser(chain, user)
		if err != nil {
			log.Error(err)
		}

		return resp
	}

	return nil
}

func (c *ServiceContext) CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error) {

	log.Debug("Checking if first entry of chain fits into 10KB")
	_, err := chain.ConvertToEntryModel().Fit10KB()
	if err != nil {
		return nil, err
	}
	chain.ChainID = chain.ID()

	// default chain status for new chains
	chain.Status = model.ChainQueue

	resp := &model.ChainWithLinks{}

	// calculate entryhash of the first entry
	resp.Links = append(resp.Links, "/chains/"+chain.ChainID+"/entries/"+chain.FirstEntryHash())

	// check if chain exists on Factom
	if chain.Exists() == true {
		log.Error("Chain " + chain.ChainID + " already exists on Factom")
		return nil, fmt.Errorf("Chain " + chain.ChainID + " exists")
	}

	// search for chain.ChainID into local DB
	localChain := c.store.GetChain(&model.Chain{ChainID: chain.ChainID})

	if localChain != nil {
		log.Error("Chain " + chain.ChainID + " already into local DB")
		return nil, fmt.Errorf("Chain " + chain.ChainID + " exists")
	}

	log.Debug("Chain ", chain.ChainID, " not found both on Factom & into local DB")

	log.Debug("Creating chain into local DB")
	err = c.store.CreateChain(chain)
	if err != nil {
		log.Error(err)
	}

	log.Debug("Creating entry into local DB")
	err = c.store.CreateEntry(chain.ConvertToEntryModel())
	if err != nil {
		log.Error(err)
	}

	log.Debug("Adding 'create-chain' into queue")

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err = c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	resp.Chain = chain

	return resp, nil

}

func (c *ServiceContext) GetEntry(entry *model.Entry) *model.Entry {

	log.Debug("Search for entry into local DB")

	// search for chain.ChainID into local DB
	localentry := c.store.GetEntry(entry)

	if localentry != nil {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")
		return localentry
	}

	log.Debug("Entry " + entry.EntryHash + " not found into local DB")
	log.Debug("Search for entry on the blockchain")

	resp, err := entry.FillModelFromFactom()
	resp.Status = model.EntryCompleted

	if err != nil {
		log.Error(err)
	}

	log.Debug("Creating entry into local DB")
	err = c.store.CreateEntry(resp)
	if err != nil {
		log.Error(err)
	}

	/*
		if chain.Exists() {
			log.Debug("Chain " + chain.ChainID + " found on the blockchain")

			log.Debug("Getting chain status from the blockchain")
			chain.Status, _ = chain.GetStatusFromFactom()
			resp.Chain = chain

			log.Debug("Creating chain into local DB")
			err := c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

			log.Debug("Start fetching chain entries into local DB")
			go c.parseBackChainEntries(chain)

			// If we are here, so no errors occured and we force bind chain to API user
			log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
			err = c.store.BindChainToUser(chain, user)
			if err != nil {
				log.Error(err)
			}

			return resp
		}
	*/

	return resp
}

func (c *ServiceContext) CreateEntry(entry *model.Entry) (*model.Entry, error) {

	log.Debug("Checking if entry fits into 10KB")
	_, err := entry.Fit10KB()
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(err.Error())
	}

	entry.EntryHash = entry.Hash()

	// default entry status for new entries
	entry.Status = model.EntryQueue

	err = c.store.CreateEntry(entry)
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(err.Error())
	}

	return entry, nil
}

func (c *ServiceContext) parseBackChainEntries(chain *model.Chain) error {

	log.Debug("Parsing entries of chain " + chain.ChainID)

	var parse_from string

	if chain.EarliestEntryBlock != "" {
		parse_from = chain.EarliestEntryBlock
	} else {
		head, err := factom.GetChainHeadAndStatus(chain.ChainID)

		if err != nil {
			return err
		}

		if head.ChainHead == "" && head.ChainInProcessList {
			return fmt.Errorf("Chain not yet included in a Directory Block")
		}

		parse_from = head.ChainHead
	}

	log.Debug("Parsing back from eblock: " + parse_from)

	err := c.parseBackChainEntriesFromEBlock(parse_from)
	if err != nil {
		return err
	}

	return nil

}

func (c *ServiceContext) parseBackChainEntriesFromEBlock(eblock string) error {

	for ebhash := eblock; ebhash != factom.ZeroHash; {

		log.Debug("Fetching EntryBlock " + ebhash)

		eb, err := factom.GetEBlock(ebhash)
		if err != nil {
			return err
		}
		entryblock := model.NewEBlockFromFactomModel(ebhash, eb)
		err = c.store.CreateEBlock(entryblock)
		if err != nil {
			return err
		}
		s, err := factom.GetAllEBlockEntries(ebhash)
		if err != nil {
			return err
		}

		var entry *model.Entry

		for _, fe := range s {
			entry = model.NewEntryFromFactomModel(fe)
			log.Debug("Fetching Entry " + entry.EntryHash)
			entry.Status = model.EntryCompleted
			entry.EntryBlock = ebhash
			err = c.store.CreateEntry(entry)
			if err != nil {
				log.Error(err)
			}
		}

		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, EarliestEntryBlock: ebhash})
		if err != nil {
			return err
		}
		ebhash = eb.Header.PrevKeyMR

		// set synced=true on last iteration
		if ebhash == factom.ZeroHash {
			err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, Synced: true})
		}

	}

	return nil

}

func (c *ServiceContext) getChainFromLocalDB(chain *model.Chain) *model.Chain {
	return c.store.GetChain(chain)
}
