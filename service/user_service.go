package service

import (
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	//	log "github.com/sirupsen/logrus"
)

type UserService interface {
	CreateUser(user *model.User) (*model.User, error)
	GetUserByAccessToken(token string) (*model.User, error)
	UpdateUser(user *model.User) error
}

func NewUserService(store store.Store) UserService {
	return &UserServiceContext{store: store}
}

type UserServiceContext struct {
	store store.Store
}

func (c *UserServiceContext) CreateUser(user *model.User) (*model.User, error) {

	tx, err := c.store.Begin()

	if err != nil {
		return nil, err
	}
	created, err := c.store.CreateUser(tx, user)
	if err != nil {
		c.store.Rollback(tx)
		return nil, err
	}
	if err = c.store.Commit(tx); err != nil {
		return nil, err
	}
	return created, nil
}

func (c *UserServiceContext) GetUserByAccessToken(token string) (*model.User, error) {
	return c.store.GetUserByAccessToken(nil, token)
}

func (c *UserServiceContext) UpdateUser(user *model.User) error {

	tx, err := c.store.Begin()
	if err != nil {
		return err
	}
	err = c.store.UpdateUser(tx, user)
	if err != nil {
		c.store.Rollback(tx)
		return err
	}
	if err = c.store.Commit(tx); err != nil {
		return err
	}
	return nil

}
