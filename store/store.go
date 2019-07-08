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

	CreateUser(user *model.User) (*model.User, error)
	GetUser(user *model.User) *model.User
	GetUsers(user *model.User) []*model.User
	UpdateUser(user *model.User) error
	DeleteUser(user *model.User) error

	GetChain(chain *model.Chain) *model.Chain
	GetChains(chain *model.Chain) []*model.Chain
	GetUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int)
	SearchUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int)
	GetChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int, sort string) ([]*model.Entry, int)
	SearchChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int, sort string) ([]*model.Entry, int)
	CreateChain(chain *model.Chain) error
	UpdateChain(chain *model.Chain) error
	UpdateChainsWhere(sql string, chain *model.Chain) error
	BindChainToUser(chain *model.Chain, user *model.User) error

	GetEntry(entry *model.Entry, sort string) *model.Entry
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
func NewStore(conf *config.Config, applyMigration bool) (Store, error) {

	storeConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Store.Host, conf.Store.Port, conf.Store.User, conf.Store.Password, conf.Store.DBName,
	)

	db, err := gorm.Open("postgres", storeConfig)
	if err != nil {
		return nil, err
	}

	if conf.API.Logging && conf.API.LogLevel >= 6 {
		db.LogMode(true)
	}

	if applyMigration == true {
		log.Info("Store: applying SQL migrations")

		migrations := &migrate.FileMigrationSource{
			Dir: "migrations",
		}

		n, err := migrate.Exec(db.DB(), "postgres", migrations, migrate.Up)
		if err != nil {
			log.Fatal(err)
		}
		log.Info("Store: applied ", n, " migration(s)")
	}

	return &Context{db}, nil

}

// Close store
func (c *Context) Close() error {

	return c.db.Close()

}

func (c *Context) CreateUser(user *model.User) (*model.User, error) {

	if c.db.Create(&user).RowsAffected > 0 {
		return user, nil
	}

	return nil, fmt.Errorf("Creating user failed")

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

	if user.Status == 0 {
		c.db.Model(user).Update("status", user.Status)
	}

	if user.Usage == 0 {
		c.db.Model(user).Update("usage", user.Usage)
	}

	if user.UsageLimit == 0 {
		c.db.Model(user).Update("usage_limit", user.UsageLimit)
	}

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

func (c *Context) GetUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int) {

	orderString := fmt.Sprintf("factom_time %s, created_at %s", sort, sort)

	res := []*model.Chain{}

	c.db.Order(orderString).Where(chain).Model(user).Related(&res, "Chains")
	total := len(res)

	if start > 0 || total > limit {
		log.Warn("Second DB Request")
		c.db.Offset(start).Limit(limit).Order(orderString).Where(chain).Model(user).Related(&res, "Chains")
	}

	return res, total

}

func (c *Context) SearchUserChains(chain *model.Chain, user *model.User, start int, limit int, sort string) ([]*model.Chain, int) {

	orderString := fmt.Sprintf("factom_time %s, created_at %s", sort, sort)

	res := []*model.Chain{}

	where := &model.Chain{}
	if chain.Status != "" {
		where.Status = chain.Status
	}

	c.db.Order(orderString).Where("ext_ids @> ?", chain.ExtIDs).Where(where).Model(user).Related(&res, "Chains")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Order(orderString).Where("ext_ids @> ?", chain.ExtIDs).Where(where).Model(user).Related(&res, "Chains")
	}
	return res, total

}

func (c *Context) CreateChain(chain *model.Chain) error {

	if err := c.db.FirstOrCreate(&chain).Error; err != nil {
		return err
	}
	return nil

}

func (c *Context) GetEntry(entry *model.Entry, sort string) *model.Entry {

	var orderString string
	if sort != "" {
		orderString = fmt.Sprintf("factom_time %s, created_at %s", sort, sort)
	}

	res := &model.Entry{}
	if c.db.Order(orderString).First(&res, entry).RecordNotFound() {
		return nil
	}
	return res

}

func (c *Context) GetChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int, sort string) ([]*model.Entry, int) {

	orderString := fmt.Sprintf("factom_time %s, created_at %s", sort, sort)

	res := []*model.Entry{}

	where := &model.Entry{}
	if entry.Status != "" {
		where.Status = entry.Status
	}

	c.db.Order(orderString).Model(chain).Where(where).Related(&res, "Entries")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Order(orderString).Model(chain).Where(where).Related(&res, "Entries")
	}
	return res, total

}

func (c *Context) SearchChainEntries(chain *model.Chain, entry *model.Entry, start int, limit int, sort string) ([]*model.Entry, int) {

	orderString := fmt.Sprintf("factom_time %s, created_at %s", sort, sort)

	res := []*model.Entry{}

	where := &model.Entry{}
	if entry.Status != "" {
		where.Status = entry.Status
	}

	c.db.Order(orderString).Where("ext_ids @> ?", entry.ExtIDs).Where(where).Model(chain).Related(&res, "Entries")
	total := len(res)

	if start > 0 || total > limit {
		c.db.Offset(start).Limit(limit).Order(orderString).Where("ext_ids @> ?", entry.ExtIDs).Where(where).Model(chain).Related(&res, "Entries")
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
