package service

import (
	"encoding/json"
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	"github.com/FactomProject/factom"
	"github.com/jinzhu/copier"
	"time"
	//factom "github.com/ilzheev/factom"
	log "github.com/sirupsen/logrus"
)

type Service interface {
	CreateUser(user *model.User) error
	GetUserByAccessToken(token string) *model.User

	GetChain(chain *model.Chain, user *model.User) *model.ChainWithLinks
	GetChains(chain *model.Chain) []*model.Chain
	GetUserChains(chain *model.Chain, user *model.User) []*model.Chain
	SetChainSentToPool(chain *model.Chain) error
	ResetChainParsing(chain *model.Chain) error
	ResetChainsParsingAtAPIStart() error
	CreateChain(chain *model.Chain, user *model.User) (*model.ChainWithLinks, error)
	GetChainEntries(chain *model.Chain, user *model.User) ([]*model.Entry, error)

	GetEntry(entry *model.Entry, user *model.User) *model.Entry
	CreateEntry(entry *model.Entry, user *model.User) (*model.Entry, error)

	GetQueue(queue *model.Queue) []*model.Queue
	GetQueueToProcess() []*model.Queue
	GetQueueToClear() []*model.Queue
	ProcessQueue(queue *model.Queue) error
	ClearQueue(queue *model.Queue) error

	ParseAllChainEntries(chain *model.Chain, workerID int) error
	ParseNewChainEntries(chain *model.Chain) error
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
		// localChain already base64 encoded
		resp.Chain = localChain
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

func (c *ServiceContext) GetUserChains(chain *model.Chain, user *model.User) []*model.Chain {

	return c.store.GetUserChains(chain, user)

}

func (c *ServiceContext) SetChainSentToPool(chain *model.Chain) error {

	t := true
	return c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, SentToPool: &t})

}

func (c *ServiceContext) ResetChainParsing(chain *model.Chain) error {

	t := false
	return c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, WorkerID: -1, SentToPool: &t})

}

func (c *ServiceContext) ResetChainsParsingAtAPIStart() error {

	t := false
	return c.store.UpdateChainsWhere("synced IS FALSE", &model.Chain{WorkerID: -1, SentToPool: &t})

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
	resp.Links = append(resp.Links, "/entries/"+chain.FirstEntryHash())

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

	err = c.addToQueue(chain.ConvertToQueueParams(), model.QueueActionChain, user)
	if err != nil {
		log.Error(err)
	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err = c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	resp.Chain = chain.Base64Encode()

	return resp, nil

}

func (c *ServiceContext) GetChainEntries(chain *model.Chain, user *model.User) ([]*model.Entry, error) {

	resp := &model.ChainWithLinks{}

	log.Debug("Search for chain into local DB")

	// search for chain.ChainID into local DB
	localChain := c.store.GetChain(chain)

	if localChain == nil {

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

		} else {
			log.Debug("Chain " + chain.ChainID + " not found on the blockchain")
			return nil, fmt.Errorf("Chain " + chain.ChainID + " not found")
		}

	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err := c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	return c.store.GetChainEntries(chain), nil

}

func (c *ServiceContext) GetEntry(entry *model.Entry, user *model.User) *model.Entry {

	log.Debug("Search for entry into local DB")

	// search for chain.ChainID into local DB
	localentry := c.store.GetEntry(entry)

	if localentry != nil {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")

		log.Debug("Force binding chain ", localentry.ChainID, " to user ", user.Name)
		err := c.store.BindChainToUser(localentry.GetChain(), user)
		if err != nil {
			log.Error(err)
		}
		// localentry already base64 encoded
		return localentry
	}

	log.Debug("Entry " + entry.EntryHash + " not found into local DB")
	log.Debug("Search for entry on the blockchain")

	resp, err := entry.FillModelFromFactom()

	if err == nil {
		log.Debug("Entry " + entry.EntryHash + " found on Factom")
		resp.Status = resp.GetStatusFromFactom()

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

		}

		log.Debug("Creating entry into local DB")
		err = c.store.CreateEntry(resp.Base64Encode())
		if err != nil {
			log.Error(err)
		}

		log.Debug("Force binding chain ", resp.ChainID, " to user ", user.Name)
		err = c.store.BindChainToUser(resp.GetChain(), user)
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

	err = c.addToQueue(entry.ConvertToQueueParams(), model.QueueActionEntry, user)
	if err != nil {
		log.Error(err)
	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", entry.ChainID, " to user ", user.Name)
	err = c.store.BindChainToUser(entry.GetChain(), user)
	if err != nil {
		log.Error(err)
	}

	return entry.Base64Encode(), nil
}

func (c *ServiceContext) addToQueue(params *model.QueueParams, action string, user *model.User) error {

	log.Debug("Adding to queue: " + action)

	queue := &model.Queue{}
	queue.Params, _ = json.Marshal(params)
	queue.Action = action
	queue.UserID = user.ID

	localQueue := c.store.GetQueueItem(queue)

	if localQueue == nil {
		err := c.store.CreateQueue(queue)
		if err != nil {
			return err
		}
	}

	return nil

}

func (c *ServiceContext) GetQueue(queue *model.Queue) []*model.Queue {

	return c.store.GetQueue(queue)

}

func (c *ServiceContext) GetQueueToProcess() []*model.Queue {

	return c.store.GetQueueWhere("processed_at IS NULL AND (next_try_at IS NULL OR next_try_at<NOW())")

}

func (c *ServiceContext) GetQueueToClear() []*model.Queue {

	return c.store.GetQueueWhere("result IS NOT NULL AND processed_at IS NOT NULL AND processed_at < NOW() - INTERVAL '1 hour'")

}

func (c *ServiceContext) ProcessQueue(queue *model.Queue) error {

	params := &model.QueueParams{}
	err := json.Unmarshal(queue.Params, &params)
	if err != nil {
		return err
	}

	debugMessage := fmt.Sprintf("Queue processing: ID=%d, action=%s, try=%d", queue.ID, queue.Action, queue.TryCount)

	var processingIsSuccess bool
	var resp string

	switch queue.Action {
	case model.QueueActionChain:
		log.Debug(debugMessage)
		chain := &model.Chain{}
		copier.Copy(chain, params)
		resp, err = c.wallet.CommitRevealChain(chain.ConvertToFactomModel())
		if err != nil {
			processingIsSuccess = false
		} else {
			processingIsSuccess = true
			err = c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, Status: model.ChainProcessing})
			if err != nil {
				return err
			}
			err = c.store.UpdateEntry(&model.Entry{EntryHash: resp, Status: model.EntryProcessing})
			if err != nil {
				return err
			}
		}
	case model.QueueActionEntry:
		log.Debug(debugMessage)
		entry := &model.Entry{}
		copier.Copy(entry, params)
		resp, err = c.wallet.CommitRevealEntry(entry.ConvertToFactomModel())
		if err != nil {
			processingIsSuccess = false
		} else {
			processingIsSuccess = true
			err = c.store.UpdateEntry(&model.Entry{EntryHash: resp, Status: model.EntryProcessing})
			if err != nil {
				return err
			}
		}
	default:
		err := fmt.Errorf("Queue processing: action=%s not implemented")
		return err
	}

	if processingIsSuccess == true {
		log.Info("Queue processing: create " + queue.Action + " success " + resp)
		queue.Result = resp
		processedAt := time.Now()
		queue.ProcessedAt = &processedAt
	} else {
		log.Error("Queue processing: create " + queue.Action + " FAILED")
		queue.TryCount++
		queue.Error = err.Error()
		nextTryAt := time.Now().Add(time.Minute)
		queue.NextTryAt = &nextTryAt
	}

	err = c.store.UpdateQueue(queue)

	if err != nil {
		return err
	}

	return nil

}

func (c *ServiceContext) ClearQueue(queue *model.Queue) error {

	debugMessage := fmt.Sprintf("Queue clearing: ID=%d", queue.ID)
	log.Debug(debugMessage)

	params := &model.QueueParams{}
	err := json.Unmarshal(queue.Params, &params)
	if err != nil {
		return err
	}

	entry := &model.Entry{EntryHash: queue.Result}

	log.Debug("Queue clearing: Checking entry " + entry.EntryHash + " status")
	entry.Status = entry.GetStatusFromFactom()

	log.Debug("Queue clearing: Entry status=" + entry.Status)

	if entry.Status == model.EntryCompleted {
		log.Debug("Queue clearing: Soft delete row from queue table")
		err := c.store.DeleteQueue(queue)
		if err != nil {
			log.Error(err)
			return err
		}
	} else {
		log.Debug("Queue clearing: Force processing this task again")
		err := c.ProcessQueue(queue)
		if err != nil {
			log.Error(err)
		}
		processedAt := time.Now()
		err = c.store.UpdateQueue(&model.Queue{ID: queue.ID, ProcessedAt: &processedAt})
		if err != nil {
			log.Error(err)
			return err
		}
	}

	return nil

}

func (c *ServiceContext) ParseNewChainEntries(chain *model.Chain) error {

	var parse_from string
	var parse_to string

	log.Debug("Updates parser: Checking chain " + chain.ChainID)

	status, chainhead := chain.GetStatusFromFactom()

	// if chain has not processed on Factom, don't touch it
	if status != model.ChainCompleted {
		return fmt.Errorf("Updates parser: Chain has not processed on Factom yet")
	}

	// parse new entries if new blocks appeared
	if chain.LatestEntryBlock != chainhead {
		log.Debug("Updates parser: Chain " + chain.ChainID + " updated, parsing new entries")
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
		log.Debug("Updates parser: No new entries found")
	}

	return nil

}

func (c *ServiceContext) ParseAllChainEntries(chain *model.Chain, workerID int) error {

	var parseFrom string
	var parseTo string

	log.Debug("History parse: Checking chain " + chain.ChainID)

	status, chainhead := chain.GetStatusFromFactom()

	// if chain has not processed on Factom, don't touch it
	if status != model.ChainCompleted {
		return fmt.Errorf("History parse: Chain has not processed on Factom yet")
	}

	t := true

	if chain.Synced == &t {
		return fmt.Errorf("History parse: Chain already parsed")
	}

	// by default, parse from chainhead
	parseFrom = chainhead

	log.Debug("History parse: Chain " + chain.ChainID + " not synced, parsing all entries")
	parseTo = factom.ZeroHash

	// if some entryblocks already parsed, start from the latest parsed
	if chain.EarliestEntryBlock != "" {
		log.Debug("History parse: Start parsing from EntryBlock " + chain.EarliestEntryBlock)
		parseFrom = chain.EarliestEntryBlock
	}

	c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, LatestEntryBlock: chainhead, Status: status, WorkerID: workerID})

	err := c.parseEntryBlocks(parseFrom, parseTo, true)
	if err != nil {
		return err
	}

	return nil

}

func (c *ServiceContext) parseEntryBlocks(parseFrom string, parseTo string, updateEarliestEntryBlock bool) error {

	for ebhash := parseFrom; ebhash != parseTo; {
		var err error
		ebhash, err = c.parseEntryBlock(ebhash, updateEarliestEntryBlock)
		if err != nil {
			return err
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
		err = c.store.CreateEntry(entry.Base64Encode())
		if err != nil {
			log.Error(err)
		}
		err = c.store.BindEntryToEBlock(entry, entryblock)
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
		t := true
		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, Synced: &t, ExtIDs: model.NewEntryFromFactomModel(s[0]).Base64Encode().ExtIDs, WorkerID: -2})
		if err != nil {
			return "", err
		}
	}

	return eb.Header.PrevKeyMR, nil

}
