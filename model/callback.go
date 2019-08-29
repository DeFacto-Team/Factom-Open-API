package model

import (
	//	"github.com/jinzhu/gorm"
	//	log "github.com/sirupsen/logrus"
	"time"
)

type Callback struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"createdAt" form:"createdAt" query:"createdAt"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ID        int    `json:"id" gorm:"primary_key;unique;not null"`
	UserID    int    `json:"userId"`
	EntryHash string `json:"entryHash" validate:"required,hexadecimal,len=64" gorm:"not null"`
	URL       string `json:"url" validate:"required,url" gorm:"not null"`
	Entry     Entry  `json:"-" form:"-" query:"-" gorm:"foreignkey:entry_hash;association_foreignkey:entry_hash"`
	Result    int    `json:"-" form:"-" query:"-"`
}
