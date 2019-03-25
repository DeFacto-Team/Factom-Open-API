package service

import (
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	"github.com/FactomProject/factom"
	//factom "github.com/ilzheev/factom"
	//	"github.com/FactomProject/factomd/wsapi"
	log "github.com/sirupsen/logrus"
)

type ChainService interface {
	GetChain(chain *model.Chain) (*model.Chain, error)
	CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error)

	getChainFromLocalDB(chain *model.Chain) *model.Chain
	parseBackChainEntries(chain *model.Chain) error
	parseBackChainEntriesFromEBlock(eblock string) error
}

func NewChainService(store store.Store, wallet wallet.Wallet) ChainService {
	return &ChainServiceContext{store: store, wallet: wallet}
}

type ChainServiceContext struct {
	store  store.Store
	wallet wallet.Wallet
}

func (c *ChainServiceContext) GetChain(chain *model.Chain) (*model.Chain, error) {
	return nil, fmt.Errorf("TEST")
	//return c.store.GetChain(nil, chain)
}

func (c *ChainServiceContext) CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error) {

	log.Debug("Checking if first entry of chain fits into 10KB")
	_, err := chain.ConvertToEntryModel().Fit10KB()
	if err != nil {
		return nil, err
	}
	chain.ChainID = chain.ID()
	// default chain status & chain sync status for new chains
	chain.Status = model.ChainQueue
	//	chain.Synced = false

	resp := &model.ChainWithLinks{}

	// calculate entryhash of the first entry
	resp.Links = append(resp.Links, "/chains/"+chain.ChainID+"/entries/"+chain.FirstEntryHash())

	// search for chain.ChainID into local DB
	localChain := c.getChainFromLocalDB(&model.Chain{ChainID: chain.ChainID})

	flagCheckExistenseOnFactom := false

	// if chain doesn't exists in local DB
	if localChain == nil {
		log.Debug("Chain ", chain.ChainID, " not found in local DB, creating it in DB")
		err = c.store.CreateChain(chain)
		if err != nil {
			log.Error(err)
		}
		err = c.store.CreateEntry(chain.ConvertToEntryModel())
		if err != nil {
			log.Error(err)
		}
		flagCheckExistenseOnFactom = true
	} else {
		log.Debug("Chain ", chain.ChainID, " found in local DB")
		if localChain.Status != model.ChainCompleted {
			log.Debug("Chain is not marked as completed into local DB")
			flagCheckExistenseOnFactom = true
		} else {
			log.Debug("Chain marked as completed into local DB")
		}
		chain.Status = localChain.Status
		chain.Synced = localChain.Synced
	}

	if flagCheckExistenseOnFactom == true {
		log.Debug("Checking chain's existence on Factom")
		if chain.Exists() == false {
			log.Debug("Chain does not exist on Factom, adding to queue")
		} else {
			log.Debug("Chain exists on Factom, updating it's status")
			chain.Status, _ = chain.GetStatusFromFactom()
			if chain.Status == model.ChainCompleted && localChain == nil {
				log.Debug("Chain processed on Factom and just added to local DB, start parsing it")
				go c.parseBackChainEntries(chain)
			}
			err = c.store.UpdateChain(chain)
			if err != nil {
				log.Error(err)
			}
		}
	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err = c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	/*
		resp.Links = append(resp.Links,
			"/chains/"+chain.ChainID+"/entries",
			"/chains/"+chain.ChainID+"/entries/first",
			"/chains/"+chain.ChainID+"/entries/last")
	*/

	resp.Chain = chain

	return resp, nil

}

/*
func (c *ChainServiceContext) CreateChain(chain *model.Chain) (*string, error) {

	tx, err := c.store.Begin()
	if err != nil {
		return nil, err
	}

	// Insert chain into "chains"
	resChain, err := c.store.CreateChain(tx, chain)
	if err != nil {
		c.store.Rollback(tx)
		return nil, err
	}
	if err = c.store.Commit(tx); err != nil {
		return nil, err
	}

	return resChain, nil
}
*/

func (c *ChainServiceContext) parseBackChainEntries(chain *model.Chain) error {

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

func (c *ChainServiceContext) parseBackChainEntriesFromEBlock(eblock string) error {

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

func (c *ChainServiceContext) getChainFromLocalDB(chain *model.Chain) *model.Chain {
	return c.store.GetChain(chain)
}
