package model

import (
	"time"
)

type User struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ID          int      `json:"-" form:"-" query:"-" validate:"required" gorm:"primary_key;unique;not null"`
	Name        string   `json:"name" form:"name" query:"name" validate:"required" gorm:"unique;not null"`
	AccessToken string   `json:"accessToken" form:"accessToken" query:"accessToken" validate:"required" gorm:"unique;not null"`
	Usage       int      `json:"usage" form:"usage" query:"usage"`
	UsageLimit  int      `json:"usageLimit" form:"usageLimit" query:"usageLimit"`
	Status      int      `json:"-" form:"-" query:"-"`
	Chains      []*Chain `json:"-" form:"-" query:"-" gorm:"many2many:users_chains;"`
}
