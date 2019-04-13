package model

import (
	"encoding/base64"
	"github.com/lib/pq"
	"time"

	"github.com/FactomProject/factom"
	"github.com/jinzhu/copier"
	//	log "github.com/sirupsen/logrus"
)

// swagger:model
type Chain struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	ChainID            string         `json:"chainId" form:"chainId" query:"chainId" validate:"required,hexadecimal,len=64" gorm:"primary_key;unique;not null"`
	ExtIDs             pq.StringArray `json:"extIds" form:"extIds" query:"extIds" validate:"required,dive,base64"`
	Content            string         `json:"-" form:"content" query:"content" sql:"-" validate:"omitempty,base64"`
	Status             string         `json:"status" form:"status" query:"status" validate:"omitempty,oneof=queue processing completed"`
	Synced             *bool          `json:"synced" form:"synced" query:"synced" gorm:"not null;default:false"`
	EarliestEntryBlock string         `json:"-" form:"-" query:"-"`
	LatestEntryBlock   string         `json:"-" form:"-" query:"-"`
	Entries            []Entry        `json:"-" form:"-" query:"-" gorm:"foreignkey:chain_id"`
	WorkerID           int            `json:"-" form:"-" query:"-" gorm:"not null;default:-1"`
	SentToPool         *bool          `json:"-" form:"-" query:"-" gorm:"not null;default:false"`
	FactomTime         time.Time      `json:"createdAt"`
}

type ChainWithLinks struct {
	*Chain
	Links []Link `json:"links" form:"links" query:"links" validate:""`
}

type Chains struct {
	Items []*Chain
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

func (chain *Chain) ConvertToQueueParams() *QueueParams {

	params := &QueueParams{}
	copier.Copy(params, chain)
	return params

}

func (chain *Chain) Base64Decode() *Chain {

	chainDecoded := &Chain{}
	copier.Copy(chainDecoded, chain)

	content, _ := base64.StdEncoding.DecodeString(chain.Content)
	chainDecoded.Content = string(content)

	var extID []byte
	var extIDs []string

	for _, i := range chain.ExtIDs {
		extID, _ = base64.StdEncoding.DecodeString(i)
		extIDs = append(extIDs, string(extID))
	}

	chainDecoded.ExtIDs = extIDs

	return chainDecoded

}

func (chain *Chain) Base64Encode() *Chain {

	chainEncoded := &Chain{}
	copier.Copy(chainEncoded, chain)

	chainEncoded.Content = base64.StdEncoding.EncodeToString([]byte(chain.Content))

	var extIDs []string

	for _, i := range chain.ExtIDs {
		extIDs = append(extIDs, base64.StdEncoding.EncodeToString([]byte(i)))
	}

	chainEncoded.ExtIDs = extIDs

	return chainEncoded

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

func (chain *Chain) ConvertToChainWithLinks() *ChainWithLinks {

	resp := &ChainWithLinks{Chain: chain}

	resp.Links = append(resp.Links, Link{Rel: "entries", Href: "/chains/" + chain.ChainID + "/entries"})
	resp.Links = append(resp.Links, Link{Rel: "firstEntry", Href: "/chains/" + chain.ChainID + "/entries/first"})
	resp.Links = append(resp.Links, Link{Rel: "lastEntry", Href: "/chains/" + chain.ChainID + "/entries/last"})

	return resp

}

func (chains Chains) ConvertToChainsWithLinks() []*ChainWithLinks {

	resp := []*ChainWithLinks{}

	for _, v := range chains.Items {
		resp = append(resp, v.ConvertToChainWithLinks())
	}

	return resp

}
