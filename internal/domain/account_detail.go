package domain

import "errors"

// Account tdetails
type AccountDetail interface {
	Validate(accountType AccountType) error
	GetType() AccountCategory
}

type BankDetail struct {
	Institution   string
	RoutingNumber string
	AccountNumber string
}

type EducationDetail struct {
	Institution     string `json:"institution" bson:"institution"`
	AccountNumber   string `json:"accountNumber" bson:"accountNumber"`
	BeneficiaryName string `json:"beneficiaryName" bson:"beneficiaryName"`
	StateProgram    string `json:"stateProgram,omitempty" bson:"stateProgram,omitempty"`
}

type BrokerageDetail struct {
	Institution   string `json:"institution" bson:"institution"`
	AccountNumber string `json:"accountNumber" bson:"accountNumber"`
}

type CryptoDetail struct {
	// For wallets (cold/hot)
	Blockchain string `json:"blockchain,omitempty" bson:"blockchain,omitempty"` // "ethereum", "solana"
	Address    string `json:"address,omitempty" bson:"address,omitempty"`

	// For exchanges
	Exchange      string `json:"exchange,omitempty" bson:"exchange,omitempty"` // "coinbase", "kraken"
	AccountNumber string `json:"accountNumber" bson:"accountNumber"`
}

func (b *BankDetail) Validate(accountType AccountType) error {
	if b.Institution == "" || b.AccountNumber == "" {
		return errors.New("institution and account number required")
	}
	return nil
}

func (b *BankDetail) GetType() AccountCategory {
	return CategoryCash
}

func (b *EducationDetail) Validate(accountType AccountType) error {
	if b.Institution == "" || b.AccountNumber == "" {
		return errors.New("institution and account number required")
	}
	return nil
}

func (b *EducationDetail) GetType() AccountCategory {
	return Category529
}

func (b *BrokerageDetail) Validate(accountType AccountType) error {
	if b.Institution == "" || b.AccountNumber == "" {
		return errors.New("institution and account number required")
	}
	return nil
}

func (b *BrokerageDetail) GetType() AccountCategory {
	return CategoryBrokerage
}

func (c *CryptoDetail) Validate(accountType AccountType) error {
	switch accountType {
	case TypeExchange:
		if c.Exchange == "" {
			return errors.New("exchange required")
		}
	case TypeImported, TypeHotWallet:
		if c.Blockchain == "" || c.Address == "" {
			return errors.New("blockchain and address required")
		}
	}
	return nil
}

func (c *CryptoDetail) GetType() AccountCategory {
	return CategoryCrypto
}
