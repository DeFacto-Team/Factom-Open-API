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
	CreateChain(chain *model.Chain) (*string, error)
	BindChainToUser(chain *model.Chain, user *model.User) error
	ParseAllChainEntries(chain *model.Chain) error
}

func NewChainService(store store.Store, wallet wallet.Wallet) ChainService {
	return &ChainServiceContext{store: store, wallet: wallet}
}

type ChainServiceContext struct {
	store  store.Store
	wallet wallet.Wallet
}

func (c *ChainServiceContext) GetChain(chain *model.Chain) (*model.Chain, error) {
	return c.store.GetChain(nil, chain)
}

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

func (c *ChainServiceContext) BindChainToUser(chain *model.Chain, user *model.User) error {

	tx, err := c.store.Begin()
	if err != nil {
		return err
	}

	// Insert chain into "users_chains"
	_, err = c.store.CreateRelationUserChain(tx, user, chain)
	if err != nil {
		c.store.Rollback(tx)
		return err
	}
	if err = c.store.Commit(tx); err != nil {
		return err
	}

	return nil

}

func (c *ChainServiceContext) ParseAllChainEntries(chain *model.Chain) error {

	log.Debug("Parsing entries of chain " + chain.ChainID)

	head, err := factom.GetChainHeadAndStatus(chain.ChainID)

	if err != nil {
		return err
	}

	if head.ChainHead == "" && head.ChainInProcessList {
		return fmt.Errorf("Chain not yet included in a Directory Block")
	}

	log.Debug("Latest EntryBlock: " + head.ChainHead)

	c.ParseChainEntriesFromEBlock(head.ChainHead)

	return nil

}

func (c *ChainServiceContext) ParseChainEntriesFromEBlock(eblock string) error {

	for ebhash := eblock; ebhash != factom.ZeroHash; {

		log.Debug("Fetching EntryBlock " + ebhash)

		eb, err := factom.GetEBlock(ebhash)
		if err != nil {
			return err
		}
		s, err := factom.GetAllEBlockEntries(ebhash)
		if err != nil {
			return err
		}

		for _, fe := range s {
			entry := model.NewEntryFromFactomModel(fe)
			log.Debug("Fetching Entry " + entry.EntryHash)

			tx, err := c.store.Begin()
			if err != nil {
				return err
			}

			_, err = c.store.CreateEntry(tx, entry)
			if err != nil {
				log.Debug(err)
				c.store.Rollback(tx)
				return err
			}

			if err = c.store.Commit(tx); err != nil {
				log.Debug(err)
				return err
			}

		}

		ebhash = eb.Header.PrevKeyMR

	}

	return nil

}
