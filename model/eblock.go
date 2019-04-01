package model

import (
	"github.com/FactomProject/factom"
)

// swagger:model
type EBlock struct {
	KeyMR               string   `json:"keymr" gorm:"primary_key;unique;not null"`
	BlockSequenceNumber int64    `json:"blocksequencenumber"`
	ChainID             string   `json:"chainid"`
	PrevKeyMR           string   `json:"prevkeymr"`
	Timestamp           int64    `json:"timestamp"`
	DBHeight            int64    `json:"dbheight"`
	Entries             []*Entry `json:"-" form:"-" query:"-" gorm:"many2many:entries_e_blocks;"`
}

func NewEBlockFromFactomModel(ebhash string, fe *factom.EBlock) *EBlock {

	eblock := EBlock{}
	eblock.KeyMR = ebhash

	if fe == nil {
		fe, _ = factom.GetEBlock(ebhash)
	}

	eblock.BlockSequenceNumber = fe.Header.BlockSequenceNumber
	eblock.ChainID = fe.Header.ChainID
	eblock.DBHeight = fe.Header.DBHeight
	eblock.PrevKeyMR = fe.Header.PrevKeyMR
	eblock.Timestamp = fe.Header.Timestamp

	return &eblock

}
