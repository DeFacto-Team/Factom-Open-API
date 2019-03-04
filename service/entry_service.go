package service

import (
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	//	"github.com/FactomProject/factom"
	//factom "github.com/ilzheev/factom"
	//	"github.com/FactomProject/factomd/wsapi"
	log "github.com/sirupsen/logrus"
)

type EntryService interface {
	//GetEntry(id int) (*model.Entry, error)
	CreateEntry(category *model.Entry) (*int, error)
}

func NewEntryService(store store.Store, wallet wallet.Wallet) EntryService {
	return &EntryServiceContext{store: store, wallet: wallet}
}

type EntryServiceContext struct {
	store  store.Store
	wallet wallet.Wallet
}

/*
func (esc *EntryServiceContext) GetEntry(id int) (*model.Entry, error) {
	return esc.store.GetEntry(nil, id)
}
*/

func (c *EntryServiceContext) CreateEntry(entry *model.Entry) (*int, error) {

	tx, err := c.store.Begin()
	if err != nil {
		return nil, err
	}

	fe := entry.ConvertToFactomModel()

	_, err = c.wallet.CommitRevealEntry(fe)
	if err != nil {
		log.Error(err)
	}

	// Insert entry into "entries"
	resEntry, err := c.store.CreateEntry(tx, entry)
	if err != nil {
		c.store.Rollback(tx)
		return nil, err
	}
	if err = c.store.Commit(tx); err != nil {
		return nil, err
	}

	return resEntry, nil
}
