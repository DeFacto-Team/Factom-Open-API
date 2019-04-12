package model

import (
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	log "github.com/sirupsen/logrus"
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
	Action      string
	Params      []byte
	Error       string     // factomd request error
	Result      string     // factomd request result
	ProcessedAt *time.Time // time when sent to Factom without error, otherwise null
	NextTryAt   *time.Time // by default null, set when processing failed to postpone next attempt
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

func (queue *Queue) AfterCreate(db *gorm.DB) {

	user := &User{}
	db.First(&user, &User{ID: queue.UserID})

	user.Usage++

	if queue.Action == QueueActionChain {
		user.Usage++
	}

	if db.Model(&user).Updates(user).RowsAffected == 0 {
		log.Error("Update usage for user ", user.Name, " failed")
	}

}
