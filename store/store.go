package store

import (
	"fmt"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"

	"github.com/jinzhu/gorm"
	"github.com/rubenv/sql-migrate"
	log "github.com/sirupsen/logrus"
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
	GetUserChains(chain *model.Chain, user *model.User, start int, limit int) ([]*model.Chain, int)
	SearchUserChains(chain *model.Chain, user *model.User, start int, limit int) ([]*model.Chain, int)
	GetChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int) ([]*model.Entry, int)
	SearchChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int) ([]*model.Entry, int)
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
type Context struct {
	db *gorm.DB
}

// Create new store
func NewStore(conf *config.Config) (Store, error) {

	storeConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Store.Host, conf.Store.Port, conf.Store.User, conf.Store.Password, conf.Store.DBName,
	)

	db, err := gorm.Open("postgres", storeConfig)
	if err != nil {
		return nil, err
	}

	if conf.API.Logging && conf.LogLevel >= 6 {
		db.LogMode(true)
	}

	log.Info("Store: applying SQL migrations")

	migrations := &migrate.FileMigrationSource{
		Dir: "store/migrations",
	}

	n, err := migrate.Exec(db.DB(), "postgres", migrations, migrate.Up)
	if err != nil {
		log.Fatal(err)
	}
	log.Info("Store: applied ", n, " migration(s)")

	return &Context{db}, nil

}

// Close store
func (c *Context) Close() error {

	return c.db.Close()

}

func (c *Context) CreateUser(user *model.User) error {

	if c.db.Create(&user).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("Creating user failed")

}

func (c *Context) GetUser(user *model.User) *model.User {

	res := &model.User{}
	if c.db.First(&res, user).RecordNotFound() {
		return nil
	}
	return res

}

func (c *Context) GetUsers(user *model.User) []*model.User {

	res := []*model.User{}
	c.db.Where(user).Find(&res)
	return res

}

func (c *Context) UpdateUser(user *model.User) error {

	if c.db.Model(&user).Updates(user).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating user failed")

}

func (c *Context) DeleteUser(user *model.User) error {

	if c.db.Delete(&user).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Deletion user failed")

}

func (c *Context) DisableUserUsageLimit(user *model.User) error {

	if c.db.Model(user).Update("usage_limit", 0).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("DB: Updating user limit failed")

}

func (c *Context) GetChain(chain *model.Chain) *model.Chain {

	res := &model.Chain{}
	if c.db.First(&res, chain).RecordNotFound() {
		return nil
	}
	return res

}

func (c *Context) GetChains(chain *model.Chain) []*model.Chain {

	res := []*model.Chain{}
	c.db.Where(chain).Find(&res)
	return res

}

func (c *Context) GetUserChains(chain *model.Chain, user *model.User, start int, limit int) ([]*model.Chain, int) {

	res := []*model.Chain{}

	c.db.Where(chain).Model(user).Related(&res, "Chains")
	total := len(res)

	if start > 0 || total > limit {
		log.Warn("Second DB Request")
		c.db.Offset(start).Limit(limit).Where(chain).Model(user).Related(&res, "Chains")
	}

	return res, total

}

func (c *Context) SearchUserChains(chain *model.Chain, user *model.User, start int, limit int) ([]*model.Chain, int) {

	res := []*model.Chain{}

	where := &model.Chain{}
	if chain.Status != "" {
		where.Status = chain.Status
	}

	c.db.Where("ext_ids @> ?", chain.ExtIDs).Where(where).Model(user).Related(&res, "Chains")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Where("ext_ids @> ?", chain.ExtIDs).Where(where).Model(user).Related(&res, "Chains")
	}
	return res, total

}

func (c *Context) CreateChain(chain *model.Chain) error {

	if err := c.db.FirstOrCreate(&chain).Error; err != nil {
		return err
	}
	return nil

}

func (c *Context) GetEntry(entry *model.Entry) *model.Entry {

	res := &model.Entry{}
	if c.db.First(&res, entry).RecordNotFound() {
		return nil
	}
	return res

}

func (c *Context) GetChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int) ([]*model.Entry, int) {

	res := []*model.Entry{}

	where := &model.Entry{}
	if entry.Status != "" {
		where.Status = entry.Status
	}

	c.db.Model(chain).Where(where).Related(&res, "Entries")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Model(chain).Where(where).Related(&res, "Entries")
	}
	return res, total

}

func (c *Context) SearchChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int) ([]*model.Entry, int) {

	res := []*model.Entry{}

	where := &model.Entry{}
	if entry.Status != "" {
		where.Status = entry.Status
	}

	c.db.Where("ext_ids @> ?", entry.ExtIDs).Where(where).Model(chain).Related(&res, "Entries")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Where("ext_ids @> ?", entry.ExtIDs).Where(where).Model(chain).Related(&res, "Entries")
	}
	return res, total

}

func (c *Context) CreateEntry(entry *model.Entry) error {

	assign := model.Entry{}
	assign.Status = entry.Status
	if entry.FactomTime != nil {
		assign.FactomTime = entry.FactomTime
	}

	if err := c.db.Assign(assign).FirstOrCreate(&entry).Error; err != nil {
		return err
	}
	return nil

}

func (c *Context) UpdateEntry(entry *model.Entry) error {

	if c.db.Model(&entry).Updates(entry).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating entry failed")

}

func (c *Context) CreateEBlock(eblock *model.EBlock) error {

	if err := c.db.FirstOrCreate(&eblock).Error; err != nil {
		return err
	}
	return nil

}

func (c *Context) UpdateChain(chain *model.Chain) error {

	if c.db.Model(&chain).Updates(chain).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating chain failed")

}

func (c *Context) UpdateChainsWhere(sql string, chain *model.Chain) error {

	c.db.Model(model.Chain{}).Where(sql).Updates(chain)

	return nil

}

func (c *Context) BindChainToUser(chain *model.Chain, user *model.User) error {

	c.db.Model(user).Association("Chains").Append(chain)

	return nil

}

func (c *Context) BindEntryToEBlock(entry *model.Entry, eblock *model.EBlock) error {

	c.db.Model(eblock).Association("Entries").Append(entry)

	return nil

}

func (c *Context) GetQueue(queue *model.Queue) []*model.Queue {

	res := []*model.Queue{}
	c.db.Where(queue).Find(&res)
	return res

}

func (c *Context) GetQueueWhere(sql string) []*model.Queue {

	res := []*model.Queue{}
	c.db.Where(sql).Find(&res)
	return res

}

func (c *Context) GetQueueItem(queue *model.Queue) *model.Queue {

	res := &model.Queue{}
	if c.db.First(&res, queue).RecordNotFound() {
		return nil
	}
	return res

}

func (c *Context) CreateQueue(queue *model.Queue) error {

	if c.db.Create(&queue).RowsAffected > 0 {
		return nil
	}

	return fmt.Errorf("Creating queue failed")

}

func (c *Context) UpdateQueue(queue *model.Queue) error {

	if c.db.Model(&queue).Updates(queue).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Updating queue failed")

}

func (c *Context) DeleteQueue(queue *model.Queue) error {

	if c.db.Delete(&queue).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Deletion queue failed")

}
