package wallet

import (
	//"encoding/json"
	"fmt"
	"github.com/DeFacto-Team/Factom-Open-API/config"
	"github.com/FactomProject/factom"
	log "github.com/sirupsen/logrus"
)

const (
	ChainECCost = 10
)

type Wallet interface {
	GetEC() *factom.ECAddress
	CommitRevealEntry(entry *factom.Entry) (string, error)
	CommitRevealChain(chain *factom.Chain) (string, error)
}

type WalletContext struct {
	ec *factom.ECAddress
}

func NewWallet(conf *config.Config) (Wallet, error) {

	// setup EC pub-priv keypair from Es address
	ECAddress, err := factom.GetECAddress(conf.Factom.EsAddress)
	if err != nil {
		return nil, fmt.Errorf("INVALID Es address set in config: %s", conf.Factom.EsAddress)
	} else {
		balance, _ := factom.GetECBalance(ECAddress.PubString())
		log.Info("Using EC address: ", ECAddress, ", balance=", balance)
		if balance == 0 {
			log.Warn("EC address balance is 0 EC. Please top up your EC address to let API create chains & entries on the blockchain.")
		}
	}

	return &WalletContext{ECAddress}, nil

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

func (c *WalletContext) CommitRevealEntry(entry *factom.Entry) (string, error) {

	// calculate entry cost
	cost, err := factom.EntryCost(entry)
	if err != nil {
		log.Error("Can not calculate Entry Cost")
		return "", err
	}

	// check if EC balance enought for tx
	if res := c.checkBalance(cost); res == false {
		err = fmt.Errorf("Not enough Entry Credits to create entry")
		log.Error(err)
		return "", err
	}

	// commit+reveal entry
	_, err = factom.CommitEntry(entry, c.GetEC())
	if err != nil {
		log.Error(err)
		return "", err
	}
	resp, err := factom.RevealEntry(entry)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return resp, nil

}

func (c *WalletContext) CommitRevealChain(chain *factom.Chain) (string, error) {

	// calculate entry cost
	cost, err := factom.EntryCost(chain.FirstEntry)
	if err != nil {
		log.Error("Can not calculate Entry Cost")
		return "", err
	}

	// check if EC balance enought for tx
	if res := c.checkBalance(cost + ChainECCost); res == false {
		err = fmt.Errorf("Not enough Entry Credits to create chain")
		log.Error(err)
		return "", err
	}

	// commit chain
	_, err = factom.CommitChain(chain, c.GetEC())
	resp, err := factom.RevealChain(chain)
	if err != nil {
		log.Error(err)
		return "", err
	}

	return resp, nil

}
