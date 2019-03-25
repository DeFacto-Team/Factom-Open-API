package service

import (
	"github.com/DeFacto-Team/Factom-Open-API/model"
	"github.com/DeFacto-Team/Factom-Open-API/store"
	//	log "github.com/sirupsen/logrus"
)

type UserService interface {
	CreateUser(user *model.User) error
	GetUserByAccessToken(token string) *model.User
	UpdateUser(user *model.User) error
}

func NewUserService(store store.Store) UserService {
	return &UserServiceContext{store: store}
}

type UserServiceContext struct {
	store store.Store
}

func (c *UserServiceContext) CreateUser(user *model.User) error {

	err := c.store.CreateUser(user)
	if err != nil {
		return err
	}

	return nil
}

func (c *UserServiceContext) GetUserByAccessToken(token string) *model.User {
	return c.store.GetUserByAccessToken(token)
}

func (c *UserServiceContext) UpdateUser(user *model.User) error {

	/*
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
	*/
	return nil

}
