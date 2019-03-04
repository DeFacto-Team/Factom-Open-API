package model

import (
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
)

const (
	MaxEntrySize = 10240
	// Entry statuses
	StatusCompleted  = "completed"
	StatusProcessing = "processing"
	StatusQueue      = "queue"
)

// swagger:model
type Entry struct {
	EntryHash string   `json:"entryhash" form:"entryhash" query:"entryhash" validate:"required,hexadecimal,len=64"`
	ChainID   string   `json:"chainid" form:"chainid" query:"chainid" validate:"required,hexadecimal,len=64"`
	ExtIDs    []string `json:"extids" form:"extids" query:"extids" validate:""`
	Content   string   `json:"content" form:"content" query:"content" validate:"required"`
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
