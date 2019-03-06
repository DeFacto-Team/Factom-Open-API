package model

import (
	"github.com/FactomProject/factom"
)

// swagger:model
type Chain struct {
	ChainID string   `json:"chainid" form:"chainid" query:"chainid" validate:"required,hexadecimal,len=64"`
	ExtIDs  []string `json:"extids" form:"extids" query:"extids" validate:"required"`
	Content string   `json:"content" form:"content" query:"content" validate:"required"`
}

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
