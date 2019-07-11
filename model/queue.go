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

type Queue struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"createdAt" form:"createdAt" query:"createdAt"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ID          int        `json:"id" form:"id" query:"id" gorm:"primary_key;unique;not null"`
	UserID      int        `json:"-" form:"-" query:"-"`
	Action      string     `json:"action" form:"action" query:"action"`
	Params      []byte     `json:"-" form:"-" query:"-"`
	Error       string     `json:"error" form:"error" query:"error"`                   // factomd request error
	Result      string     `json:"result" form:"result" query:"result"`                // factomd request result
	ProcessedAt *time.Time `json:"processedAt" form:"processedAt" query:"processedAt"` // time when sent to Factom without error, otherwise null
	NextTryAt   *time.Time `json:"nextTryAt" form:"nextTryAt" query:"nextTryAt"`       // by default null, set when processing failed to postpone next attempt
	TryCount    int        `json:"tryCount" form:"tryCount" query:"tryCount"`
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
