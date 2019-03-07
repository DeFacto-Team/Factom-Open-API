package service

import (
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	//	"github.com/FactomProject/factom"
	//factom "github.com/ilzheev/factom"
	//	"github.com/FactomProject/factomd/wsapi"
	//log "github.com/sirupsen/logrus"
)

type ChainService interface {
	GetChain(chain *model.Chain) (*model.Chain, error)
	CreateChain(chain *model.Chain) (*string, error)
	BindChainToUser(chain *model.Chain, user *model.User) error
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
