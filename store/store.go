package store

import (
	//"database/sql"
	//"encoding/json"
	//"errors"
	"fmt"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"

	"github.com/jinzhu/gorm"
	//	_ "github.com/jinzhu/gorm/dialects/postgres"
	//log "github.com/sirupsen/logrus"
)

type Store interface {
	Close() error

	CreateUser(user *model.User) error
	GetUserByAccessToken(token string) *model.User

	GetChain(chain *model.Chain) *model.Chain
	GetChainEntries(chain *model.Chain) []model.Entry
	CreateChain(chain *model.Chain) error
	UpdateChain(chain *model.Chain) error
	BindChainToUser(chain *model.Chain, user *model.User) error

	CreateEntry(entry *model.Entry) error
	CreateEBlock(eblock *model.EBlock) error

	/*
		Begin() (*sql.Tx, error)
		Commit(tx *sql.Tx) error
		Rollback(tx *sql.Tx) error

		CreateUser(tx *sql.Tx, user *model.User) (*model.User, error)
		GetUserByAccessToken(tx *sql.Tx, token string) (*model.User, error)
		UpdateUser(tx *sql.Tx, user *model.User) error

		GetChain(tx *sql.Tx, chain *model.Chain) (*model.Chain, error)
		CreateChain(tx *sql.Tx, chain *model.Chain) (*string, error)
		CreateRelationUserChain(tx *sql.Tx, user *model.User, chain *model.Chain) (*string, error)

		CreateEntry(tx *sql.Tx, entry *model.Entry) (*int, error)
	*/
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

func (c *StoreContext) CreateChain(chain *model.Chain) error {

	if c.db.Create(&chain).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Creating chain failed")

}

func (c *StoreContext) CreateEntry(entry *model.Entry) error {

	if c.db.Create(&entry).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Creating entry failed")

}

func (c *StoreContext) CreateEBlock(eblock *model.EBlock) error {

	if c.db.Create(&eblock).RowsAffected > 0 {
		return nil
	}
	return fmt.Errorf("DB: Creating entryblock failed")

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

func (c *StoreContext) GetChainEntries(chain *model.Chain) []model.Entry {

	entries := []model.Entry{}

	c.db.Model(chain).Related(&entries, "Entries")

	return entries

}

/*
// Begin SQL tx
func (c *StoreContext) Begin() (*sql.Tx, error) {
	return c.db.Begin()
}

// Commit SQL tx
func (c *StoreContext) Commit(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	return tx.Commit()
}

// Rollback SQL tx
func (c *StoreContext) Rollback(tx *sql.Tx) error {
	if tx == nil {
		return errors.New("tx is nil")
	}
	return tx.Rollback()
}

// Create user
func (c *StoreContext) CreateUser(tx *sql.Tx, user *model.User) (*model.User, error) {
	var query = "INSERT INTO users (name, access_token) VALUES($1, $2) RETURNING id;"
	var id int
	var err error
	if tx != nil {
		err = tx.QueryRow(query, user.Name, user.AccessToken).Scan(&id)
	} else {
		err = c.db.QueryRow(query, user.Name, user.AccessToken).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	user.ID = id
	return user, nil
}

// Find user to authentificate
func (c *StoreContext) GetUserByAccessToken(tx *sql.Tx, token string) (*model.User, error) {
	var query = "SELECT id, name, usage, usage_limit FROM users WHERE access_token=$1 AND status=1;"
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, token)
	} else {
		row = c.db.QueryRow(query, token)
	}
	user := &model.User{}
	if err := row.Scan(&user.ID, &user.Name, &user.Usage, &user.UsageLimit); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		} else {
			return nil, nil
		}
	}
	return user, nil
}

// Update user
func (c *StoreContext) UpdateUser(tx *sql.Tx, user *model.User) error {
	query := "UPDATE users SET name=$1, usage=$2, usage_limit=$3 WHERE id=$4;"
	var res sql.Result
	var err error
	if tx != nil {
		res, err = tx.Exec(query, user.Name, user.Usage, user.UsageLimit, user.ID)
	} else {
		res, err = c.db.Exec(query, user.Name, user.Usage, user.UsageLimit, user.ID)
	}
	if err != nil {
		return err
	}
	if a, err := res.RowsAffected(); err != nil {
		return err
	} else if a == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Get chain by ChainID
func (c *StoreContext) GetChain(tx *sql.Tx, chain *model.Chain) (*model.Chain, error) {
	var query = "SELECT chainid, content, extids, status FROM chains WHERE chainid=$1;"
	var row *sql.Row
	if tx != nil {
		row = tx.QueryRow(query, chain.ChainID)
	} else {
		row = c.db.QueryRow(query, chain.ChainID)
	}
	var extids string
	res := &model.Chain{}
	if err := row.Scan(&res.ChainID, &res.Content, &extids, &res.Status); err != nil {
		if err != sql.ErrNoRows {
			return nil, err
		} else {
			return nil, nil
		}
	}
	json.Unmarshal([]byte(extids), &res.ExtIDs)
	return res, nil
}

// Create chain
func (c *StoreContext) CreateChain(tx *sql.Tx, chain *model.Chain) (*string, error) {
	var query = "INSERT INTO chains (chainid, content, extids, status, sync) VALUES($1, $2, $3, $4, $5) ON CONFLICT (chainid) DO UPDATE SET chainid = $1, status = $4 RETURNING chainid;"
	extids, _ := json.Marshal(chain.ExtIDs)
	var chainid string
	var err error
	if tx != nil {
		err = tx.QueryRow(query, chain.ChainID, chain.Content, extids, chain.Status, chain.Sync).Scan(&chainid)
	} else {
		err = c.db.QueryRow(query, chain.ChainID, chain.Content, extids, chain.Status, chain.Sync).Scan(&chainid)
	}
	if err != nil {
		return nil, err
	}
	return &chainid, nil
}

// Create user-chain relation
func (c *StoreContext) CreateRelationUserChain(tx *sql.Tx, user *model.User, chain *model.Chain) (*string, error) {
	var query = "INSERT INTO users_chains (user_id, chainid) VALUES($1, $2) ON CONFLICT (user_id, chainid) DO UPDATE SET user_id = $1, chainid = $2 RETURNING chainid;"
	var chainid string
	var err error
	if tx != nil {
		err = tx.QueryRow(query, user.ID, chain.ChainID).Scan(&chainid)
	} else {
		err = c.db.QueryRow(query, user.ID, chain.ChainID).Scan(&chainid)
	}
	if err != nil {
		return nil, err
	}
	return &chainid, nil
}

// Create entry
func (c *StoreContext) CreateEntry(tx *sql.Tx, entry *model.Entry) (*int, error) {
	var query = "INSERT INTO entries (entryhash) VALUES($1) ON CONFLICT (entryhash) DO UPDATE SET entryhash = $1 RETURNING id;"
	//	entrydata, _ := json.Marshal(entry)
	var id int
	var err error
	if tx != nil {
		err = tx.QueryRow(query, entry.EntryHash).Scan(&id)
	} else {
		err = c.db.QueryRow(query, entry.EntryHash).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}
*/
