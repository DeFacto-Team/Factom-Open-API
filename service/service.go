package service

import (
	"encoding/json"
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	"github.com/DeFacto-Team/Factom-Open-API/wallet"
	"github.com/FactomProject/factom"
	"github.com/jinzhu/copier"
	log "github.com/sirupsen/logrus"
	"time"
)

// Service is an interface with all core functions
type Service interface {
	GetUser(user *model.User) *model.User
	GetUsers(user *model.User) []*model.User
	CreateUser(user *model.User) (*model.User, error)
	CheckUser(token string) *model.User
	UpdateUser(user *model.User) error
	DeleteUser(user *model.User) error

	GetChain(chain *model.Chain, user *model.User) (*model.Chain, error)
	GetChains(chain *model.Chain) []*model.Chain
	GetUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int)
	SearchUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int)
	SetChainSentToPool(chain *model.Chain) error
	ResetChainParsing(chain *model.Chain) error
	ResetChainsParsingAtAPIStart() error
	CreateChain(chain *model.Chain, user *model.User) (*model.Chain, error)
	GetChainEntries(entry *model.Entry, user *model.User, start int, limit int, sort string, force bool) ([]*model.Entry, int, error)
	SearchChainEntries(entry *model.Entry, user *model.User, start int, limit int, sort string, force bool) ([]*model.Entry, int, error)
	GetChainFirstOrLastEntry(entry *model.Entry, sort string, user *model.User) (*model.Entry, error)

	GetEntry(entry *model.Entry, user *model.User) (*model.Entry, error)
	CreateEntry(entry *model.Entry, user *model.User) (*model.Entry, error)

	GetQueue(queue *model.Queue) []*model.Queue
	GetQueueToProcess() []*model.Queue
	GetQueueToClear() []*model.Queue
	ProcessQueue(queue *model.Queue) error
	ClearQueue(queue *model.Queue) error

	ParseAllChainEntries(chain *model.Chain, workerID int) error
	ParseNewChainEntries(chain *model.Chain) error
}

// NewService initializes service with store & wallet as ServiceContext
func NewService(store store.Store, wallet wallet.Wallet) Service {
	return &Context{store: store, wallet: wallet}
}

// Context keeps store & wallet instances
type Context struct {
	store  store.Store
	wallet wallet.Wallet
}

// GetUser is generic function to get user from db
func (c *Context) GetUser(user *model.User) *model.User {

	return c.store.GetUser(user)

}

// GetUsers is generic function to get items from users db
func (c *Context) GetUsers(user *model.User) []*model.User {

	return c.store.GetUsers(user)

}

// CreateUser is generic function to create user into DB
func (c *Context) CreateUser(user *model.User) (*model.User, error) {

	resp, err := c.store.CreateUser(user)
	if err != nil {
		return nil, err
	}

	return resp, nil
}

// CheckUser returns only enabled users by their access token
func (c *Context) CheckUser(token string) *model.User {
	return c.store.GetUser(&model.User{AccessToken: token, Status: 1})
}

// UpdateUser is generic function to update user into DB
func (c *Context) UpdateUser(user *model.User) error {

	err := c.store.UpdateUser(user)
	if err != nil {
		return err
	}

	return nil

}

// UpdateUser is generic function to delete user from DB
func (c *Context) DeleteUser(user *model.User) error {

	err := c.store.DeleteUser(user)
	if err != nil {
		return err
	}

	return nil
}

// GetChain is high-level function, that run by api.GetChain()
func (c *Context) GetChain(chain *model.Chain, user *model.User) (*model.Chain, error) {

	log.Debug("Search for chain into local DB")

	// search for chain.ChainID into local DB
	localChain := c.store.GetChain(chain)

	if localChain != nil {
		log.Debug("Chain " + chain.ChainID + " found into local DB")

		log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
		err := c.store.BindChainToUser(chain, user)
		if err != nil {
			log.Error(err)
		}

		// localChain already base64 encoded
		return localChain, nil
	}

	log.Debug("Chain " + chain.ChainID + " not found into local DB")
	log.Debug("Search for chain on the blockchain")

	if chain.Exists() {
		chain = chain.Base64Encode()
		log.Debug("Chain " + chain.ChainID + " found on the blockchain")

		log.Debug("Getting chain status from the blockchain")
		chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()

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

		return chain, nil
	}

	return nil, fmt.Errorf("Chain %s does not exist", chain.ChainID)
}

// GetChains is generic function to get items from chains db
func (c *Context) GetChains(chain *model.Chain) []*model.Chain {

	return c.store.GetChains(chain)

}

// GetUserChains is high-level function, that run by api.GetChains()
func (c *Context) GetUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int) {

	return c.store.GetUserChains(chain, user, start, limit, sort)

}

// SearchUserChains is high-level function, that run by api.SearchChains()
func (c *Context) SearchUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int) {

	return c.store.SearchUserChains(chain, user, start, limit, sort)

}

// SetChainSentToPool marks chain as sent into history fetching pool for not sending it into pool more than once
func (c *Context) SetChainSentToPool(chain *model.Chain) error {

	t := true
	return c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, SentToPool: &t})

}

// ResetChainParsing resets WorkerID & SentToPool params of the chain. A fallback function, that runs in case of error while chain syncing.
func (c *Context) ResetChainParsing(chain *model.Chain) error {

	t := false
	return c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, WorkerID: -1, SentToPool: &t})

}

// ResetChainsParsingAtAPIStart resets WorkerID & SentToPool params of unsynced chains on API start, to let them finish syncing
func (c *Context) ResetChainsParsingAtAPIStart() error {

	t := false
	return c.store.UpdateChainsWhere("synced IS FALSE", &model.Chain{WorkerID: -1, SentToPool: &t})

}

// CreateChain is high-level function, that run by api.CreateChain()
func (c *Context) CreateChain(chain *model.Chain, user *model.User) (*model.Chain, error) {

	chain = chain.Base64Decode()

	log.Debug("Checking if first entry of chain fits into 10KB")
	_, err := chain.ConvertToEntryModel().Fit10KB()
	if err != nil {
		return nil, err
	}
	chain.ChainID = chain.ID()

	// default chain status for new chains
	chain.Status = model.ChainQueue

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

	// new chain & entry into local DB will be created with FactomTime=NOW()
	timeNow := time.Now().UTC().Round(time.Second)
	chain.FactomTime = &timeNow

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

	return chain.Base64Encode(), nil

}

// GetChainEntries is high-level function, that run by api.GetChainEntries()
func (c *Context) GetChainEntries(entry *model.Entry, user *model.User, start int, limit int, sort string, force bool) ([]*model.Entry, int, error) {

	flagJustCreated := false

	log.Debug("Search for chain into local DB")

	// search for chain.ChainID into local DB
	chain := entry.GetChain()
	localChain := c.store.GetChain(chain)

	if localChain == nil {

		log.Debug("Chain " + chain.ChainID + " not found into local DB")
		log.Debug("Search for chain on the blockchain")

		if chain.Exists() {
			chain = chain.Base64Encode()
			log.Debug("Chain " + chain.ChainID + " found on the blockchain")

			log.Debug("Getting chain status from the blockchain")
			chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()

			log.Debug("Creating chain into local DB")
			err := c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

			flagJustCreated = true

		} else {
			log.Debug("Chain " + chain.ChainID + " not found on the blockchain")
			return nil, 0, fmt.Errorf("Chain " + chain.ChainID + " not found")
		}

	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err := c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	// if force=true not passed
	if !force {
		// check if chain just created or not fully synced yet
		if flagJustCreated == true || (localChain.Status == model.ChainCompleted && !(*localChain.Synced)) {
			return nil, 0, nil
		}
	}

	result, total := c.store.GetChainEntries(entry.GetChain(), entry, start, limit, sort)

	return result, total, nil

}

// SearchChainEntries is high-level function, that run by api.SearchChainEntries()
func (c *Context) SearchChainEntries(entry *model.Entry, user *model.User, start int, limit int, sort string, force bool) ([]*model.Entry, int, error) {

	flagJustCreated := false

	log.Debug("Search for chain into local DB")

	chain := entry.GetChain()

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

			log.Debug("Creating chain into local DB")
			err := c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

			flagJustCreated = true

		} else {
			log.Debug("Chain " + chain.ChainID + " not found on the blockchain")
			return nil, 0, fmt.Errorf("Chain " + chain.ChainID + " not found")
		}

	}

	// If we are here, so no errors occured and we force bind chain to API user
	log.Debug("Force binding chain ", chain.ChainID, " to user ", user.Name)
	err := c.store.BindChainToUser(chain, user)
	if err != nil {
		log.Error(err)
	}

	// if force=true not passed
	if !force {
		// check if chain just created or not fully synced yet
		if flagJustCreated == true || (localChain.Status == model.ChainCompleted && !(*localChain.Synced)) {
			return nil, 0, nil
		}
	}

	result, total := c.store.SearchChainEntries(entry.GetChain(), entry, start, limit, sort)

	return result, total, nil

}

// GetChainFirstOrLastEntry is high-level function, that run by api.GetChainFirstEntry() && api.GetChainLastEntry()
func (c *Context) GetChainFirstOrLastEntry(entry *model.Entry, sort string, user *model.User) (*model.Entry, error) {

	flagJustCreated := false

	log.Debug("Search for chain into local DB")

	// search for chain.ChainID into local DB
	chain := entry.GetChain()
	localChain := c.store.GetChain(chain)

	if localChain == nil {

		log.Debug("Chain " + chain.ChainID + " not found into local DB")
		log.Debug("Search for chain on the blockchain")

		if chain.Exists() {
			chain = chain.Base64Encode()
			log.Debug("Chain " + chain.ChainID + " found on the blockchain")

			log.Debug("Getting chain status from the blockchain")
			chain.Status, chain.LatestEntryBlock = chain.GetStatusFromFactom()

			log.Debug("Creating chain into local DB")
			err := c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

			flagJustCreated = true

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

	// check if chain just created or not fully synced yet
	if flagJustCreated == true {
		return nil, nil
	}

	// check if chain not fully synced yet
	if localChain.Status == model.ChainCompleted && !(*localChain.Synced) {
		switch sort {
		case "asc":
			return nil, nil
		case "desc":
			// while requesting last entry, it's enough to have at least one entry block parsed to return a result
			if localChain.EarliestEntryBlock == "" {
				return nil, nil
			}
		}
	}

	result := c.store.GetEntry(entry, sort)

	return result, nil

}

// GetEntry is high-level function, that run by api.GetEntry()
func (c *Context) GetEntry(entry *model.Entry, user *model.User) (*model.Entry, error) {

	log.Debug("Search for entry into local DB")

	// search for chain.ChainID into local DB
	localentry := c.store.GetEntry(entry, "")

	if localentry != nil {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")

		log.Debug("Force binding chain ", localentry.ChainID, " to user ", user.Name)
		err := c.store.BindChainToUser(localentry.GetChain(), user)
		if err != nil {
			log.Error(err)
		}
		// localentry already base64 encoded
		return localentry, nil
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

			// here we add existing Factom chain into local DB with factomTime = null
			err = c.store.CreateChain(chain)
			if err != nil {
				log.Error(err)
			}

		}

		log.Debug("Creating entry into local DB")

		// get entry timestamp from Factom ONLY IF ENTRY STATUS IS COMPLETED
		if resp.Status == model.EntryCompleted {
			factomTime, err := resp.GetTimeFromFactom()
			if err != nil {
				log.Error(err)
			} else {
				t := time.Unix(factomTime, 0).UTC()
				resp.FactomTime = &t
			}
		}

		err = c.store.CreateEntry(resp.Base64Encode())
		if err != nil {
			log.Error(err)
		}

		log.Debug("Force binding chain ", resp.ChainID, " to user ", user.Name)
		err = c.store.BindChainToUser(resp.GetChain(), user)
		if err != nil {
			log.Error(err)
		}

		return resp.Base64Encode(), nil
	}

	return nil, fmt.Errorf("Entry %s does not exist", entry.EntryHash)

}

// CreateEntry is high-level function, that run by api.CreateEntry()
func (c *Context) CreateEntry(entry *model.Entry, user *model.User) (*model.Entry, error) {

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

	localEntry := c.store.GetEntry(&model.Entry{EntryHash: entry.EntryHash}, "")

	if localEntry == nil {
		log.Debug("Entry " + entry.EntryHash + " not found into local DB")
		log.Debug("Creating entry into local DB")

		// new entry status queue, factomTime NOW()
		entry.Status = model.EntryQueue
		timeNow := time.Now().UTC().Round(time.Second)
		entry.FactomTime = &timeNow

		err = c.store.CreateEntry(entry.Base64Encode())
		if err != nil {
			log.Error(err)
			return nil, fmt.Errorf(err.Error())
		}
	} else {
		log.Debug("Entry " + entry.EntryHash + " found into local DB")
		// use entry status from local db
		entry.Status = localEntry.Status
		entry.FactomTime = localEntry.FactomTime
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

// addToQueue checks if task already exists into queue db and if not, then adds the task into queue db
func (c *Context) addToQueue(params *model.QueueParams, action string, user *model.User) error {

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

// GetQueue is generic function to get items from queue db
func (c *Context) GetQueue(queue *model.Queue) []*model.Queue {

	return c.store.GetQueue(queue)

}

// GetQueueToProcess gets unprocessed and failed (while previous processing) tasks from queue
func (c *Context) GetQueueToProcess() []*model.Queue {

	return c.store.GetQueueWhere("processed_at IS NULL AND (next_try_at IS NULL OR next_try_at<NOW())")

}

// GetQueueToClear gets tasks from queue, that was successfully processed more than 1 hour ago
func (c *Context) GetQueueToClear() []*model.Queue {

	return c.store.GetQueueWhere("result IS NOT NULL AND processed_at IS NOT NULL AND processed_at < NOW() - INTERVAL '1 hour'")

}

// ProcessQueue processes write task from queue: makes factomd commit+reveal request and update queue item according to response (success or error)
func (c *Context) ProcessQueue(queue *model.Queue) error {

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
		err := fmt.Errorf("Queue processing: action=%s not implemented", queue.Action)
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

// ClearQueue gets entry status from Factom, checks if it's 'completed' and then deletes task.
// Otherwise it runs ProcessQueue() for force processing.
func (c *Context) ClearQueue(queue *model.Queue) error {

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

// ParseNewChainEntries fetches new entries of chain, that appeared on Factom inside all new entry blocks till chain.LatestEntryBlock
func (c *Context) ParseNewChainEntries(chain *model.Chain) error {

	var parseFrom string
	var parseTo string

	log.Debug("Updates parser: Checking chain " + chain.ChainID)

	status, chainhead := chain.GetStatusFromFactom()

	// if chain has not processed on Factom, don't touch it
	if status != model.ChainCompleted {
		return fmt.Errorf("Updates parser: Chain has not processed on Factom yet")
	}

	// parse new entries if new blocks appeared
	if chain.LatestEntryBlock != chainhead {
		log.Debug("Updates parser: Chain " + chain.ChainID + " updated, parsing new entries")
		parseFrom = chainhead
		parseTo = chain.LatestEntryBlock
		err := c.parseEntryBlocks(parseFrom, parseTo, false)
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

// ParseAllChainEntries fetches all entries of chain from Factom.
// By default the parsing starts from ChainHead.
// If chain is partially fetched (i.e. reference field chain.EarliestEntryBlock != ""), the parsing starts from EarliestEntryBlock
func (c *Context) ParseAllChainEntries(chain *model.Chain, workerID int) error {

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

	// set chain LatestEntryBlock & assign worker ID
	c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, LatestEntryBlock: chainhead, WorkerID: workerID})

	// parsing chain entryblocks & entries recursively
	err := c.parseEntryBlocks(parseFrom, parseTo, true)
	if err != nil {
		return err
	}

	// if no errors (chain synced completely), then update chain status to completed
	// chain.Synced was already updated to true into parseEntryBlocks() → parseEntryBlock()
	// existing chains are 'completed' initially
	// this needed only for recently created chains with status "processing",
	// so basically the single entryblock parsed in the previous case
	// (except the case when API / factomd was offline for a long time, but there is no issue)
	c.store.UpdateChain(&model.Chain{ChainID: chain.ChainID, Status: status})

	return nil

}

// Parses entries from all entryblocks between parseFrom & parseTo.
// updateEarliestEntryBlock is a flag, that reflect should chain.EarliestEntryBlock be updated during the parsing.
// While history fetching (from X block till the first block), chain.EarliestEntryBlock is being updated.
// While updates fetching (from ChainHead till latest parsed block), chain.EarliestEntryBlock is NOT being updated.
func (c *Context) parseEntryBlocks(parseFrom string, parseTo string, updateEarliestEntryBlock bool) error {

	for ebhash := parseFrom; ebhash != parseTo; {
		var err error
		ebhash, err = c.parseEntryBlock(ebhash, updateEarliestEntryBlock)
		if err != nil {
			return err
		}
	}

	return nil

}

// Parses all entries from the entryblock and returns keymr of previous entryblock
func (c *Context) parseEntryBlock(ebhash string, updateEarliestEntryBlock bool) (string, error) {

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

	var entry *model.Entry
	var fistEntryOfEntryBlock *model.Entry

	for i, listItem := range eb.EntryList {
		fe, err := factom.GetEntry(listItem.EntryHash)
		if err != nil {
			return "", err
		}
		entry = model.NewEntryFromFactomModel(fe)
		log.Debug("Fetching Entry " + entry.EntryHash)
		entry.Status = model.EntryCompleted
		t := time.Unix(listItem.Timestamp, 0).UTC()
		entry.FactomTime = &t
		err = c.store.CreateEntry(entry.Base64Encode())
		if err != nil {
			log.Error(err)
			return "", err
		}
		err = c.store.BindEntryToEBlock(entry, entryblock)
		if err != nil {
			log.Error(err)
			return "", err
		}
		if i == 0 {
			fistEntryOfEntryBlock = entry
		}
	}

	if updateEarliestEntryBlock == true {
		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, EarliestEntryBlock: ebhash})
		if err != nil {
			return "", err
		}
	}

	// if we parsed the first entry block, set synced=true & update extIDs & set FactomTime to time of the block
	if eb.Header.PrevKeyMR == factom.ZeroHash {
		t := true
		factomTime := time.Unix(eb.Header.Timestamp, 0).UTC()
		// s[0] — first entry of the entry block
		err = c.store.UpdateChain(&model.Chain{ChainID: eb.Header.ChainID, Synced: &t, ExtIDs: fistEntryOfEntryBlock.Base64Encode().ExtIDs, FactomTime: &factomTime, WorkerID: -2})
		if err != nil {
			return "", err
		}
	}

	return eb.Header.PrevKeyMR, nil

}
