package model

import (
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"github.com/FactomProject/factom"
	"github.com/jinzhu/copier"
	"github.com/jinzhu/gorm"
	"github.com/lib/pq"
	"time"
)

const (
	// Statuses
	EntryCompleted  = "completed"
	EntryProcessing = "processing"
	EntryQueue      = "queue"

	// Factom Statuses
	FactomEntryUnknown         = "Unknown"
	FactomEntryNotConfirmed    = "NotConfirmed"
	FactomEntryTransactionACK  = "TransactionACK"
	FactomEntryDBlockConfirmed = "DBlockConfirmed"

	MaxEntrySize = 10240
)

// swagger:model
type Entry struct {
	// gorm.Model without ID
	CreatedAt time.Time  `json:"-" form:"-" query:"-"`
	UpdatedAt time.Time  `json:"-" form:"-" query:"-"`
	DeletedAt *time.Time `json:"-" form:"-" query:"-"`
	// model
	EntryHash   string         `json:"entryHash" form:"entryHash" query:"entryHash" validate:"required,hexadecimal,len=64" gorm:"primary_key;unique;not null"`
	ChainID     string         `json:"chainId" form:"chainId" query:"chainId" validate:"required,hexadecimal,len=64"`
	ExtIDs      pq.StringArray `json:"extIds" form:"extIds" query:"extIds" validate:"omitempty,dive,base64"`
	Content     string         `json:"content" form:"content" query:"content" validate:"omitempty,base64"`
	Status      string         `json:"status" form:"status" query:"status" validate:"omitempty,oneof=queue processing completed" gorm:"not null;default:'queue'"`
	EntryBlocks []*EBlock      `json:"-" form:"-" query:"-" gorm:"many2many:entries_e_blocks;"`
	FactomTime  *time.Time     `json:"createdAt"`
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

func (entry *Entry) ConvertToQueueParams() *QueueParams {

	params := &QueueParams{}
	copier.Copy(params, entry)
	return params

}

func (entry *Entry) Base64Decode() *Entry {

	entryDecoded := &Entry{}
	copier.Copy(entryDecoded, entry)

	content, _ := base64.StdEncoding.DecodeString(entry.Content)
	entryDecoded.Content = string(content)

	var extID []byte
	var extIDs []string

	for _, i := range entry.ExtIDs {
		extID, _ = base64.StdEncoding.DecodeString(i)
		extIDs = append(extIDs, string(extID))
	}

	entryDecoded.ExtIDs = extIDs

	return entryDecoded

}

func (entry *Entry) Base64Encode() *Entry {

	entryEncoded := &Entry{}
	copier.Copy(entryEncoded, entry)

	entryEncoded.Content = base64.StdEncoding.EncodeToString([]byte(entry.Content))

	var extIDs []string

	for _, i := range entry.ExtIDs {
		extIDs = append(extIDs, base64.StdEncoding.EncodeToString([]byte(i)))
	}

	entryEncoded.ExtIDs = extIDs

	return entryEncoded

}

func (entry *Entry) GetChain() *Chain {

	chain := &Chain{ChainID: entry.ChainID}

	return chain

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

func (entry *Entry) FillModelFromFactom() (*Entry, error) {

	fe, err := factom.GetEntry(entry.EntryHash)
	if err != nil {
		return nil, err
	}

	return NewEntryFromFactomModel(fe), nil

}

func (entry *Entry) GetStatusFromFactom() string {

	status, _ := factom.EntryRevealACK(entry.EntryHash, "", factom.ZeroHash)

	if status.EntryData.Status == FactomEntryDBlockConfirmed {
		return EntryCompleted
	}

	if status.EntryData.Status == FactomEntryTransactionACK {
		return EntryProcessing
	}

	return EntryQueue

}

func (entry *Entry) GetTimeFromFactom() (int64, error) {

	status, err := factom.EntryRevealACK(entry.EntryHash, "", factom.ZeroHash)

	if err != nil {
		return 0, err
	}

	return status.EntryData.BlockDate, nil

}

func (entry *Entry) BeforeUpdate(scope *gorm.Scope) error {

	dbEntry := &Entry{}

	// check entry on the local db
	scope.DB().First(&dbEntry, &Entry{EntryHash: entry.EntryHash})

	// this is fix for double creation of entry in the chain — queue processed and entry status should change to processing
	// but if entry already marked as completed, don't change it's status to processing (keep it completed)
	if dbEntry.Status == EntryCompleted {
		scope.SetColumn("Status", EntryCompleted)
	}

	return nil

}
