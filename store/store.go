package store

import (
	"fmt"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"

	"github.com/jinzhu/gorm"
	//log "github.com/sirupsen/logrus"
)

type Store interface {
	Close() error

	CreateUser(user *model.User) error
	GetUser(user *model.User) *model.User
	GetUsers(user *model.User) []*model.User
	UpdateUser(user *model.User) error
	DeleteUser(user *model.User) error
	DisableUserUsageLimit(chain *model.User) error

	GetChain(chain *model.Chain) *model.Chain
	GetChains(chain *model.Chain) []*model.Chain
	GetUserChains(chain *model.Chain, user *model.User) []*model.Chain
	SearchUserChains(chain *model.Chain, user *model.User) []*model.Chain
	GetChainEntries(chain *model.Chain) []*model.Entry
	SearchChainEntries(chain *model.Chain, entry *model.Entry) []*model.Entry
	CreateChain(chain *model.Chain) error
	UpdateChain(chain *model.Chain) error
	UpdateChainsWhere(sql string, chain *model.Chain) error
	BindChainToUser(chain *model.Chain, user *model.User) error

	GetEntry(entry *model.Entry) *model.Entry
	CreateEntry(entry *model.Entry) error
	UpdateEntry(entry *model.Entry) error
	CreateEBlock(eblock *model.EBlock) error
	BindEntryToEBlock(entry *model.Entry, eblock *model.EBlock) error

	GetQueue(queue *model.Queue) []*model.Queue
	GetQueueWhere(sql string) []*model.Queue
	GetQueueItem(queue *model.Queue) *model.Queue
	CreateQueue(queue *model.Queue) error
	UpdateQueue(queue *model.Queue) error
	DeleteQueue(queue *model.Queue) error
}

// Контекст стореджа
type StoreContext struct {
	db *gorm.DB
}

// Create new store
func NewStore(conf *config.Config) (Store, error) {

	storeConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Store.Host, conf.Store.Port, conf.Store.User, conf.Store.Password, conf.Store.Dbname,
	)

	db, err := gorm.Open("postgres", storeConfig)
	if err != nil {
		return nil, err
	}

	if conf.Api.Logging && conf.LogLevel >= 6 {
		db.LogMode(true)
	}

	return &StoreContext{db}, nil

}

// Close store
func (c *StoreContext) Close() error {

	return c.db.Close()

}

func (c *StoreContext) CreateUser(user *model.User) error {

	if c.db.Create(&user).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("Creating user failed")

}

func (c *StoreContext) GetUser(user *model.User) *model.User {

	res := &model.User{}
	if c.db.First(&res, user).RecordNotFound() {
		return nil
	}
	return res

}

func (c *StoreContext) GetUsers(user *model.User) []*model.User {

	res := []*model.User{}
	c.db.Where(user).Find(&res)
	return res

}

func (c *StoreContext) UpdateUser(user *model.User) error {

	if c.db.Model(&user).Updates(user).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating user failed")

}

func (c *StoreContext) DeleteUser(user *model.User) error {

	if c.db.Delete(&user).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Deletion user failed")

}

func (c *StoreContext) DisableUserUsageLimit(user *model.User) error {

	if c.db.Model(user).Update("usage_limit", 0).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("DB: Updating user limit failed")

}

func (c *StoreContext) GetChain(chain *model.Chain) *model.Chain {

	res := &model.Chain{}
	if c.db.First(&res, chain).RecordNotFound() {
		return nil
	}
	return res

}

func (c *StoreContext) GetChains(chain *model.Chain) []*model.Chain {

	res := []*model.Chain{}
	c.db.Where(chain).Find(&res)
	return res

}

func (c *StoreContext) GetUserChains(chain *model.Chain, user *model.User) []*model.Chain {

	res := []*model.Chain{}
	c.db.Where(chain).Model(user).Related(&res, "Chains")
	return res

}

func (c *StoreContext) SearchUserChains(chain *model.Chain, user *model.User) []*model.Chain {

	res := []*model.Chain{}
	c.db.Where("ext_ids @> ?", chain.ExtIDs).Model(&user).Related(&res, "Chains")
	return res

}

func (c *StoreContext) CreateChain(chain *model.Chain) error {

	if err := c.db.FirstOrCreate(&chain).Error; err != nil {
		return err
	}
	return nil

}

func (c *StoreContext) GetEntry(entry *model.Entry) *model.Entry {

	res := &model.Entry{}
	if c.db.First(&res, entry).RecordNotFound() {
		return nil
	}
	return res

}

func (c *StoreContext) SearchChainEntries(chain *model.Chain, entry *model.Entry) []*model.Entry {

	res := []*model.Entry{}
	c.db.Where("ext_ids @> ?", entry.ExtIDs).Model(chain).Related(&res, "Entries")
	return res

}

func (c *StoreContext) CreateEntry(entry *model.Entry) error {

	if err := c.db.Assign(model.Entry{Status: entry.Status}).FirstOrCreate(&entry).Error; err != nil {
		return err
	}
	return nil

}

func (c *StoreContext) UpdateEntry(entry *model.Entry) error {

	if c.db.Model(&entry).Updates(entry).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating entry failed")

}

func (c *StoreContext) CreateEBlock(eblock *model.EBlock) error {

	if err := c.db.FirstOrCreate(&eblock).Error; err != nil {
		return err
	}
	return nil

}

func (c *StoreContext) UpdateChain(chain *model.Chain) error {

	if c.db.Model(&chain).Updates(chain).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating chain failed")

}

func (c *StoreContext) UpdateChainsWhere(sql string, chain *model.Chain) error {

	c.db.Model(model.Chain{}).Where(sql).Updates(chain)

	return nil

}

func (c *StoreContext) BindChainToUser(chain *model.Chain, user *model.User) error {

	c.db.Model(user).Association("Chains").Append(chain)

	return nil

}

func (c *StoreContext) BindEntryToEBlock(entry *model.Entry, eblock *model.EBlock) error {

	c.db.Model(eblock).Association("Entries").Append(entry)

	return nil

}

func (c *StoreContext) GetChainEntries(chain *model.Chain) []*model.Entry {

	res := []*model.Entry{}

	c.db.Model(chain).Related(&res, "Entries")

	return res

}

func (c *StoreContext) GetQueue(queue *model.Queue) []*model.Queue {

	res := []*model.Queue{}
	c.db.Where(queue).Find(&res)
	return res

}

func (c *StoreContext) GetQueueWhere(sql string) []*model.Queue {

	res := []*model.Queue{}
	c.db.Where(sql).Find(&res)
	return res

}

func (c *StoreContext) GetQueueItem(queue *model.Queue) *model.Queue {

	res := &model.Queue{}
	if c.db.First(&res, queue).RecordNotFound() {
		return nil
	}
	return res

}

func (c *StoreContext) CreateQueue(queue *model.Queue) error {

	if c.db.Create(&queue).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("Creating queue failed")

}

func (c *StoreContext) UpdateQueue(queue *model.Queue) error {

	if c.db.Model(&queue).Updates(queue).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating queue failed")

}

func (c *StoreContext) DeleteQueue(queue *model.Queue) error {

	if c.db.Delete(&queue).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Deletion queue failed")

}
