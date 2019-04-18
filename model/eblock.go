package model

import (
	"github.com/FactomProject/factom"
)

type EBlock struct {
	KeyMR               string   `json:"keyMr" gorm:"primary_key;unique;not null"`
	BlockSequenceNumber int64    `json:"blockSequenceNumber"`
	ChainID             string   `json:"chainId"`
	PrevKeyMR           string   `json:"prevKeyMr"`
	Timestamp           int64    `json:"timestamp"`
	DBHeight            int64    `json:"dbHeight"`
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
