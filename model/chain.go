package model

import (
	//	"encoding/hex"
	"github.com/FactomProject/factom"
	//log "github.com/sirupsen/logrus"
)

// swagger:model
type Chain struct {
	ChainID string   `json:"chainid" form:"chainid" query:"chainid" validate:"required,hexadecimal,len=64"`
	ExtIDs  []string `json:"extids" form:"extids" query:"extids" validate:"required"`
	Content string   `json:"content" form:"content" query:"content" validate:"required"`
	Status  string   `json:"status" form:"status" query:"status" validate:"omitempty,oneof=queue processing completed"`
}

const (
	// Statuses
	ChainCompleted  = "completed"
	ChainProcessing = "processing"
	ChainQueue      = "queue"
)

func (chain *Chain) ConvertToEntryModel() *Entry {

	entry := &Entry{}
	entry.ExtIDs = chain.ExtIDs
	entry.Content = chain.Content
	if chain.ChainID != "" {
		entry.ChainID = chain.ChainID
	}
	return entry

}

func (chain *Chain) ConvertToFactomModel() *factom.Chain {

	fe := chain.ConvertToEntryModel().ConvertToFactomModel()

	fc := factom.NewChain(fe)
	return fc

}

func (chain *Chain) ID() string {

	fc := chain.ConvertToFactomModel()
	return fc.ChainID

}

func (chain *Chain) FirstEntryHash() string {

	entry := chain.ConvertToEntryModel()
	return entry.Hash()

}

func (chain *Chain) Exists() bool {

	return factom.ChainExists(chain.ChainID)

}

func (chain *Chain) GetStatusFromFactom() string {

	status, err := factom.GetChainHeadAndStatus(chain.ChainID)
	if err != nil {
		return ChainQueue
	}

	if status.ChainInProcessList == true {
		return ChainProcessing
	}

	return ChainCompleted

}
