package model

import (
	"encoding/base64"
	"github.com/lib/pq"
	"time"
	//	"encoding/json"
	//	"encoding/hex"

	"github.com/FactomProject/factom"
	//log "github.com/sirupsen/logrus"
)

// swagger:model
type Chain struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ChainID            string         `json:"chainid" form:"chainid" query:"chainid" validate:"required,hexadecimal,len=64" gorm:"primary_key;unique;not null"`
	ExtIDs             pq.StringArray `json:"extids" form:"extids" query:"extids" validate:"required,dive,base64"`
	Content            string         `json:"-" form:"-" query:"-" sql:"-" validate:"required,base64"`
	Status             string         `json:"status" form:"status" query:"status" validate:"omitempty,oneof=queue processing completed"`
	Synced             bool           `json:"synced" form:"synced" query:"synced" gorm:"not null;default:false"`
	EarliestEntryBlock string         `json:"-" form:"-" query:"-"`
	LatestEntryBlock   string         `json:"-" form:"-" query:"-"`
	Entries            []Entry        `json:"-" form:"-" query:"-" gorm:"foreignkey:chain_id"`
}

type ChainWithLinks struct {
	*Chain
	Links []string `json:"links" form:"links" query:"links" validate:""`
}

const (
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
	entry.EntryHash = entry.Hash()
	entry.Status = chain.Status
	return entry

}

func (chain *Chain) Base64Decode() *Chain {

	content, _ := base64.StdEncoding.DecodeString(chain.Content)
	chain.Content = string(content)

	var extID []byte
	var extIDs []string

	for _, i := range chain.ExtIDs {
		extID, _ = base64.StdEncoding.DecodeString(i)
		extIDs = append(extIDs, string(extID))
	}

	chain.ExtIDs = extIDs

	return chain

}

func (chain *Chain) Base64Encode() *Chain {

	chain.Content = base64.StdEncoding.EncodeToString([]byte(chain.Content))

	var extIDs []string

	for _, i := range chain.ExtIDs {
		extIDs = append(extIDs, base64.StdEncoding.EncodeToString([]byte(i)))
	}

	chain.ExtIDs = extIDs

	return chain

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

func (chain *Chain) GetStatusFromFactom() (string, string) {

	status, err := factom.GetChainHeadAndStatus(chain.ChainID)
	if err != nil {
		return ChainQueue, ""
	}

	if status.ChainHead == "" {
		return ChainProcessing, ""
	}

	return ChainCompleted, status.ChainHead

}
