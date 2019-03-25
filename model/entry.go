package model

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/lib/pq"
	"time"
)

const (
	// Statuses
	EntryCompleted  = "completed"
	EntryProcessing = "processing"
	EntryQueue      = "queue"
	MaxEntrySize    = 10240
)

// swagger:model
type Entry struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	EntryHash  string         `json:"entryhash" form:"entryhash" query:"entryhash" validate:"required,hexadecimal,len=64" gorm:"primary_key;unique;not null"`
	ChainID    string         `json:"chainid" form:"chainid" query:"chainid" validate:"required,hexadecimal,len=64"`
	ExtIDs     pq.StringArray `json:"extids" form:"extids" query:"extids" validate:""`
	Content    string         `json:"content" form:"content" query:"content" validate:"required"`
	Status     string         `json:"status" form:"status" query:"status" validate:"omitempty,oneof=queue processing completed" gorm:"not null;default:'queue'"`
	EntryBlock string         `json:"-" form:"-" query:"-"`
}

func NewEntryFromFactomModel(fe *factom.Entry) *Entry {

	entry := Entry{}
	entry.ChainID = fe.ChainID
	entry.Content = string(fe.Content)
	for _, i := range fe.ExtIDs {
		entry.ExtIDs = append(entry.ExtIDs, string(i))
	}
	entry.EntryHash = entry.Hash()
	return &entry

}

func (entry *Entry) ConvertToFactomModel() *factom.Entry {

	fe := &factom.Entry{}
	fe.ChainID = entry.ChainID
	fe.Content = []byte(entry.Content)
	for _, i := range entry.ExtIDs {
		fe.ExtIDs = append(fe.ExtIDs, []byte(i))
	}
	return fe

}

func (entry *Entry) Hash() string {

	fe := entry.ConvertToFactomModel()
	return hex.EncodeToString(fe.Hash())

}

func (entry *Entry) ECCost() int {

	fe := entry.ConvertToFactomModel()
	cost, err := factom.EntryCost(fe)
	if err != nil {
		return -1
	}
	return int(cost)

}

func (entry *Entry) Size() int {

	size := len(entry.Content)

	if len(entry.ExtIDs) > 0 {
		for _, extid := range entry.ExtIDs {
			size += len(extid)
		}
	}

	return size

}

func (entry *Entry) Fit10KB() (bool, error) {

	if entry.Size() < MaxEntrySize {
		return true, nil
	}
	return false, fmt.Errorf("Entry cannot be larger than 10KB")

}
