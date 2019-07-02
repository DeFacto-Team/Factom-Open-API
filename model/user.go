package model

import (
	"github.com/liip/sheriff"
	"time"
)

type User struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ID          int      `json:"id" form:"id" query:"id" validate:"required" gorm:"primary_key;unique;not null"`
	Name        string   `json:"name" form:"name" query:"name" validate:"required" gorm:"not null" groups:"api"`
	AccessToken string   `json:"accessToken" form:"accessToken" query:"accessToken" validate:"required" gorm:"unique;not null" groups:"api"`
	Usage       int      `json:"usage" form:"usage" query:"usage" groups:"api"`
	UsageLimit  int      `json:"usageLimit" form:"usageLimit" query:"usageLimit" groups:"api"`
	Status      int      `json:"status" form:"status" query:"status" gorm:"not null;default:1"`
	Chains      []*Chain `json:"-" form:"-" query:"-" gorm:"many2many:users_chains;"`
}

type Users struct {
	Items []*User
}

func (user *User) FilterStruct(groups []string) (interface{}, error) {

	o := &sheriff.Options{
		Groups: groups,
	}

	data, err := sheriff.Marshal(o, user)

	if err != nil {
		return nil, err
	}

	return data, nil

}
