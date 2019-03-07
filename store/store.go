package store

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/DeFacto-Team/Factom-Open-API/model"
)

type Store interface {
	Close() error
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
}

// Контекст стореджа
type StoreContext struct {
	db *sql.DB
}

// Create new store
func NewStore(conf *config.Config) (Store, error) {
	storeConfig := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		conf.Store.Host, conf.Store.Port, conf.Store.User, conf.Store.Password, conf.Store.Dbname,
	)
	db, err := sql.Open("postgres", storeConfig)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &StoreContext{db}, nil
}

// Close store
func (c *StoreContext) Close() error {
	return c.db.Close()
}

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
	var query = "INSERT INTO chains (chainid, content, extids) VALUES($1, $2, $3) ON CONFLICT (chainid) DO UPDATE SET chainid = $1 RETURNING chainid;"
	extids, _ := json.Marshal(chain.ExtIDs)
	var chainid string
	var err error
	if tx != nil {
		err = tx.QueryRow(query, chain.ChainID, chain.Content, extids).Scan(&chainid)
	} else {
		err = c.db.QueryRow(query, chain.ChainID, chain.Content, extids).Scan(&chainid)
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
	var query = "INSERT INTO entries (entryhash, entrydata) VALUES($1, $2) ON CONFLICT (entryhash) DO UPDATE SET entrydata = $2 RETURNING id;"
	entrydata, _ := json.Marshal(entry)
	var id int
	var err error
	if tx != nil {
		err = tx.QueryRow(query, entry.EntryHash, entrydata).Scan(&id)
	} else {
		err = c.db.QueryRow(query, entry.EntryHash, entrydata).Scan(&id)
	}
	if err != nil {
		return nil, err
	}
	return &id, nil
}
