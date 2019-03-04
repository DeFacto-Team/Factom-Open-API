package wallet

import (
	//"encoding/json"
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/FactomProject/factom"
	log "github.com/sirupsen/logrus"
)

type Wallet interface {
	GetEC() *factom.ECAddress
	CommitRevealEntry(entry *factom.Entry) (*factom.Entry, error)
}

type WalletContext struct {
	ec *factom.ECAddress
}

func NewWallet(conf *config.Config) (Wallet, error) {

	// setup EC pub-priv keypair from Es address
	ec_address, err := factom.GetECAddress(conf.Factom.EsAddress)
	if err != nil {
		return nil, fmt.Errorf("INVALID Es address set in config: %s", conf.Factom.EsAddress)
	} else {
		balance, _ := factom.GetECBalance(ec_address.PubString())
		log.Info("Using EC address: ", ec_address, ", balance=", balance)
		if balance == 0 {
			log.Warn("EC address balance is 0 EC. Please top up your EC address to let API create chains & entries on the blockchain.")
		}
	}

	return &WalletContext{ec_address}, nil

}

func (c *WalletContext) GetEC() *factom.ECAddress {
	return c.ec
}

func (c *WalletContext) checkBalance(cost int8) bool {

	balance, _ := factom.GetECBalance(c.ec.PubString())
	if balance < int64(cost) {
		return false
	}

	return true

}

func (c *WalletContext) CommitRevealEntry(entry *factom.Entry) (*factom.Entry, error) {

	// calculate entry cost
	cost, err := factom.EntryCost(entry)
	if err != nil {
		err := fmt.Errorf("Can not calculate Entry Cost")
		return nil, err
	}

	// check if EC balance enought for tx
	if res := c.checkBalance(cost); res == false {
		return nil, fmt.Errorf("Not enough Entry Credits to create entry")
	}

	// commit entry
	commitResult, err := factom.CommitEntry(entry, c.GetEC())
	if err != nil {
		return nil, err
	}
	log.Info("Commit: ", commitResult)

	// reveal entry
	revealResult, err := factom.RevealEntry(entry)
	if err != nil {
		return nil, err
	}
	log.Info("Reveal: ", revealResult)

	return entry, nil

}
