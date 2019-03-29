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
	GetChains(chain *model.Chain) []*model.Chain
	CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error)
	ParseAllChainEntries(chain *model.Chain) error
	ParseNewChainEntries(chain *model.Chain) error

	GetEntry(entry *model.Entry, user *model.User) *model.Entry
	CreateEntry(entry *model.Entry, user *model.User) (*model.Entry, error)
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
		resp.Chain = localChain.Base64Encode()
		return resp
	}

	log.Debug("Chain " + chain.ChainID + " not found into local DB")
	log.Debug("Search for chain on the blockchain")

	if chain.Exists() {
		chain = chain.Base64Encode()
		log.Debug("Chain " + chain.ChainID + " found on the blockchain")

		log.Debug("Getting chain status from the blockchain")
		chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()
		resp.Chain = chain

		log.Debug("Creating chain into local DB")
		err := c.store.CreateChain(chain)
		if err != nil {
			log.Error(err)
		}

		log.Debug("Start fetching chain entries into local DB")
		go c.ParseAllChainEntries(chain)

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

func (c *ServiceContext) GetChains(chain *model.Chain) []*model.Chain {

	return c.store.GetChains(chain)

}

func (c *ServiceContext) CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error) {

	chain = chain.Base64Decode()

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
	err = c.store.CreateChain(chain.Base64Encode())
	if err != nil {
		log.Error(err)
	}

	log.Debug("Creating entry into local DB")
	err = c.store.CreateEntry(chain.ConvertToEntryModel().Base64Encode())
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

	resp.Chain = chain.Base64Encode()

	return resp, nil

}

func (c *ServiceContext) GetEntry(entry *model.Entry, user *model.User) *model.Entry {

	log.Debug("Search for entry into local DB")

	// search for chain.ChainID into local DB
	localentry := c.store.GetEntry(entry)

	if localentry != nil {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")
		// localentry already base64 encoded
		return localentry
	}

	log.Debug("Entry " + entry.EntryHash + " not found into local DB")
	log.Debug("Search for entry on the blockchain")

	resp, err := entry.FillModelFromFactom()

	if err == nil {
		log.Debug("Entry " + entry.EntryHash + " found on Factom")
		resp.Status = model.EntryCompleted

		// search for chain.ChainID into local DB
		localChain := c.store.GetChain(resp.GetChain())

		if localChain == nil {
			log.Debug("Chain " + resp.ChainID + " not found into local DB")
			log.Debug("Creating chain into local DB")

			chain := resp.GetChain()
			chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()

			err = c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

			log.Debug("Start fetching chain entries into local DB")
			go c.ParseAllChainEntries(resp.GetChain())
		}

		log.Debug("Force binding chain ", resp.ChainID, " to user ", user.Name)
		err = c.store.BindChainToUser(resp.GetChain(), user)
		if err != nil {
			log.Error(err)
		}

		log.Debug("Creating entry into local DB")
		err = c.store.CreateEntry(resp.Base64Encode())
		if err != nil {
			log.Error(err)
		}

		return resp.Base64Encode()
	}

	return nil

}

func (c *ServiceContext) CreateEntry(entry *model.Entry, user *model.User) (*model.Entry, error) {

	entry = entry.Base64Decode()

	log.Debug("Checking if entry fits into 10KB")
	_, err := entry.Fit10KB()
	if err != nil {
		log.Error(err)
		return nil, fmt.Errorf(err.Error())
	}

	entry.EntryHash = entry.Hash()

	localChain := c.store.GetChain(entry.GetChain())
	if localChain == nil {

		log.Debug("Chain " + entry.ChainID + " not found into local DB")
		log.Debug("Checking if chain exists on Factom")

		if !entry.GetChain().Exists() {
			log.Error("Chain " + entry.ChainID + " not found on Factom")
			return nil, fmt.Errorf("Chain " + entry.ChainID + " not found")
		}

		log.Debug("Creating chain into local DB")

		chain := entry.GetChain()
		chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()
		err = c.store.CreateChain(chain)
		if err != nil {
			log.Error(err)
		}

		log.Debug("Start fetching chain entries into local DB")
		go c.ParseAllChainEntries(chain)

	}

	localEntry := c.store.GetEntry(&model.Entry{EntryHash: entry.EntryHash})

	if localEntry == nil {
		log.Debug("Entry " + entry.EntryHash + " not found into local DB")
		log.Debug("Creating entry into local DB")

		// new entry status queue
		entry.Status = model.EntryQueue

		err = c.store.CreateEntry(entry.Base64Encode())
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf(err.Error())
		}
	} else {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")
		// use entry status from local db
		entry.Status = localEntry.Status
	}

	log.Debug("Adding 'create-entry' into queue")

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", entry.ChainID, " to user ", user.Name)
	err = c.store.BindChainToUser(entry.GetChain(), user)
	if err != nil {
		log.Error(err)
	}

	return entry, nil
}

func (c *ServiceContext) ParseNewChainEntries(chain *model.Chain) error {

	var parse_from string
	var parse_to string

	log.Debug("Checking chain " + chain.ChainID + " for updates")

	status, chainhead := chain.GetStatusFromFactom()

	// if chain has not processed on Factom, don't touch it
	if status != model.ChainCompleted {
		return fmt.Errorf("Chain has not processed on Factom yet")
	}

	// parse new entries if new blocks appeared
	if chain.LatestEntryBlock != chainhead {
		log.Debug("Chain " + chain.ChainID + " updated, parsing new entries")
		parse_from = chainhead
		parse_to = chain.LatestEntryBlock
		err := c.parseEntryBlocks(parse_from, parse_to, false)
		if err == nil {
			err = c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, LatestEntryBlock: chainhead})
			if err != nil {
				return err
			}
		}
	} else {
		log.Debug("No new entries found")
	}

	return nil

}

func (c *ServiceContext) ParseAllChainEntries(chain *model.Chain) error {

	var parseFrom string
	var parseTo string

	log.Debug("Checking chain " + chain.ChainID)

	status, chainhead := chain.GetStatusFromFactom()

	// if chain has not processed on Factom, don't touch it
	if status != model.ChainCompleted {
		return fmt.Errorf("Chain has not processed on Factom yet")
	}

	if chain.Synced == true {
		return fmt.Errorf("Chain already parsed")
	}

	// by default, parse from chainhead
	parseFrom = chainhead

	log.Debug("Chain " + chain.ChainID + " not synced, parsing all entries")
	parseTo = factom.ZeroHash

	// if some entryblocks already parsed, start from the latest parsed
	if chain.EarliestEntryBlock != "" {
		log.Debug("Start parsing from EntryBlock " + chain.EarliestEntryBlock)
		parseFrom = chain.EarliestEntryBlock
	}

	c.parseEntryBlocks(parseFrom, parseTo, true)

	return nil

}

func (c *ServiceContext) parseEntryBlocks(parseFrom string, parseTo string, updateEarliestEntryBlock bool) error {

	for ebhash := parseFrom; ebhash != parseTo; {
		var err error
		ebhash, err = c.parseEntryBlock(ebhash, updateEarliestEntryBlock)
		if err != nil {
			return fmt.Errorf(err.Error())
		}
	}

	return nil

}

func (c *ServiceContext) parseEntryBlock(ebhash string, updateEarliestEntryBlock bool) (string, error) {

	log.Debug("Fetching EntryBlock " + ebhash)

	eb, err := factom.GetEBlock(ebhash)
	if err != nil {
		return "", err
	}
	entryblock := model.NewEBlockFromFactomModel(ebhash, eb)
	err = c.store.CreateEBlock(entryblock)
	if err != nil {
		return "", err
	}
	s, err := factom.GetAllEBlockEntries(ebhash)
	if err != nil {
		return "", err
	}

	var entry *model.Entry

	for _, fe := range s {
		entry = model.NewEntryFromFactomModel(fe)
		log.Debug("Fetching Entry " + entry.EntryHash)
		entry.Status = model.EntryCompleted
		entry.EntryBlock = ebhash
		err = c.store.CreateEntry(entry.Base64Encode())
		if err != nil {
			log.Error(err)
		}
	}

	if updateEarliestEntryBlock == true {
		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, EarliestEntryBlock: ebhash})
		if err != nil {
			return "", err
		}
	}

	// if we parsed earliest block, set synced=true & update extIDs
	if eb.Header.PrevKeyMR == factom.ZeroHash {
		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, Synced: true, ExtIDs: model.NewEntryFromFactomModel(s[0]).Base64Encode().ExtIDs})
		if err != nil {
			return "", err
		}
	}

	return eb.Header.PrevKeyMR, nil

}

func (c *ServiceContext) getChainFromLocalDB(chain *model.Chain) *model.Chain {
	return c.store.GetChain(chain)
}
