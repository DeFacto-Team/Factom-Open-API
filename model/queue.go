package model

import (
	"github.com/lib/pq"
	"time"
)

const (
	QueueActionChain = "chain"
	QueueActionEntry = "entry"
)

// swagger:model
type Queue struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ID          int `gorm:"primary_key;unique;not null"`
	UserID      int
	Action      string `validate:"oneof=newchain,newentry"`
	Params      []byte
	Error       string
	Result      string
	ProcessedAt *time.Time
	NextTryAt   *time.Time
	TryCount    int
}

type QueueParams struct {
	Content string
	ExtIDs  pq.StringArray
	ChainID string
}

func (Queue) TableName() string {
	return "queue"
}
