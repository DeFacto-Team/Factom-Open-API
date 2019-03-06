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
	return entry

}

func (chain *Chain) ConvertToFactomModel() *factom.Chain {

	fc := &factom.Chain{}
	fc.ChainID = chain.ChainID
	fc.FirstEntry.ChainID = chain.ChainID
	fc.FirstEntry.Content = []byte(chain.Content)
	for _, i := range chain.ExtIDs {
		fc.FirstEntry.ExtIDs = append(fc.FirstEntry.ExtIDs, []byte(i))
	}
	return fc

}
