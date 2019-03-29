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
	GetUserByAccessToken(token string) *model.User

	GetChain(chain *model.Chain) *model.Chain
	GetChains(chain *model.Chain) []*model.Chain
	GetChainEntries(chain *model.Chain) []*model.Entry
	CreateChain(chain *model.Chain) error
	UpdateChain(chain *model.Chain) error
	BindChainToUser(chain *model.Chain, user *model.User) error

	GetEntry(entry *model.Entry) *model.Entry
	CreateEntry(entry *model.Entry) error
	CreateEBlock(eblock *model.EBlock) error
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

	if conf.Api.Logging && conf.LogLevel == 5 {
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

func (c *StoreContext) GetUserByAccessToken(token string) *model.User {

	user := &model.User{AccessToken: token}

	if c.db.First(&user, user).RecordNotFound() {
		return nil
	}
	return user

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

func (c *StoreContext) CreateEntry(entry *model.Entry) error {

	if err := c.db.Assign(model.Entry{EntryBlock: entry.EntryBlock, Status: entry.Status}).FirstOrCreate(&entry).Error; err != nil {
		return err
	}
	return nil

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

func (c *StoreContext) BindChainToUser(chain *model.Chain, user *model.User) error {

	c.db.Model(user).Association("Chains").Append(chain)

	if c.db.Model(user).Related(&chain, "Chains").RecordNotFound() {
		return fmt.Errorf("DB: Binding chain to user failed")
	}

	return nil

}

func (c *StoreContext) GetChainEntries(chain *model.Chain) []*model.Entry {

	entries := []*model.Entry{}

	c.db.Model(chain).Related(&entries, "Entries")

	return entries

}
