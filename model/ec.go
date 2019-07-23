package model

import (
	"github.com/FactomProject/factom"
	"math/rand"
)

type EC struct {
	EsAddress string `json:"esAddress" form:"esAddress" query:"esAddress"`
	ECAddress string `json:"ecAddress" form:"ecAddress" query:"ecAddress"`
	Balance   int64  `json:"balance" form:"balance" query:"balance"`
}

// Returns Es/EC strings & balance from Es string
func GetEC(EsAddress string) *EC {

	ECAddress, err := factom.GetECAddress(EsAddress)
	if err != nil {
		return nil
	}
	return &EC{EsAddress: EsAddress, ECAddress: ECAddress.PubString()}

}

// Generate new Es/EC keypair
func GenerateEC() *EC {

	key := make([]byte, 32)
	rand.Read(key)

	newAddress, err := factom.MakeECAddress(key)
	if err != nil {
		return nil
	}
	return &EC{EsAddress: newAddress.SecString(), ECAddress: newAddress.PubString(), Balance: 0}

}

func (ec *EC) GetBalanceFromFactom() {

	balance, err := factom.GetECBalance(ec.ECAddress)

	if err != nil {
		balance = 0
	}

	ec.Balance = balance

}
