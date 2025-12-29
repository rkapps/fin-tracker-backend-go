package accounts

import (
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"
)

const (
	ACCOUNT_COL          = "account"
	WALLET_TYPE_WALLET   = "Wallet"
	WALLET_TYPE_EXCHANGE = "Exchange"
	WALLET_TYPE_IMPORTED = "Imported Wallet"
)

type Account struct {
	ID             string     `json:"id" bson:"id"`
	UID            string     `json:"-"`
	Group          string     `json:"group" bson:"group"`
	Category       string     `json:"category" bson:"category"`
	Name           string     `json:"name" bson:"name"`
	AlternateNames string     `json:"alternateNames" bson:"alternateNames"`
	WalletType     string     `json:"walletType" bson:"wallettype"`
	Blockchain     string     `json:"blockchain" bson:"blockchain"`
	Address        string     `json:"address" bson:"address"`
	Active         bool       `json:"active" bson:"active"`
	LastSyncDate   *time.Time `json:"lastSyncDate"`
	Resync         bool       `json:"resync"`
	Refresh        bool       `json:"refresh"`
}

type Accounts []*Account

func (u *Account) Id() string {
	return u.ID
}

// SetId sets the unique id for the ticket
func (u *Account) SetId() {

	idStr := fmt.Sprintf("%s-%s%s%s%s%s%s", u.UID, u.Group, u.Category, u.Name, u.WalletType, u.Blockchain, u.Address)
	h := sha1.New()
	h.Write([]byte(idStr))
	u.ID = hex.EncodeToString(h.Sum(nil))
}

func (u *Account) CollectionName() string {
	return ACCOUNT_COL
}
