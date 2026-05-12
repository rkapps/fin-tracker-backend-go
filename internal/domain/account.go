package domain

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/charmbracelet/log"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type Account struct {
	ID                string            `json:"id" bson:"id"`
	UID               string            `json:"-" bson:"uid"`
	Name              string            `json:"name" bson:"name"`
	Active            bool              `json:"active" bson:"active"`
	Category          AccountCategory   `json:"category" bson:"category"`
	Type              AccountType       `json:"type" bson:"type"`
	AlternateNames    []string          `json:"alternateNames,omitempty" bson:"alternateNames,omitempty"`
	Detail            AccountDetail     `json:"detail" bson:"detail"`
	TaxStatus         TaxStatus         `json:"taxStatus" bson:"taxStatus"`
	LotMatchingMethod LotMatchingMethod `json:"lotMatchingMethod" bson:"lotMatchingMethod"`
	CreatedAt         time.Time         `json:"createdAt" bson:"createdAt"`
	UpdatedAt         time.Time         `json:"updatedAt" bson:"updatedAt"`
}

// Id returns the unique id for the ticker
func (a *Account) Id() string {
	return a.ID
}

func (a *Account) CollectionName() string {
	return ACCOUNT_COLLECTION_NAME
}

type Accounts []*Account

type AccountCategory string

const (
	CategoryCash       AccountCategory = "cash"
	CategoryBrokerage  AccountCategory = "brokerage"
	CategoryRetirement AccountCategory = "retirement"
	CategoryHSA        AccountCategory = "hsa"
	CategoryCrypto     AccountCategory = "crypto"
	Category529        AccountCategory = "529"
)

// AccountType
type AccountType string

const (
	// Cash Management types
	TypeChecking    AccountType = "checking"
	TypeSavings     AccountType = "savings"
	TypeMoneyMarket AccountType = "money_market"

	// Retirement types
	TypeTraditionalIRA AccountType = "traditional_ira"
	TypeRolloverIRA    AccountType = "rollover_ira"
	TypeRoth           AccountType = "roth"

	// Crypto types
	TypeExchange  AccountType = "exchange"
	TypeImported  AccountType = "imported"
	TypeHotWallet AccountType = "hot_wallet"

	// Empty for simple categories
	TypeRegular AccountType = "regular" // For Brokerage, HSA, 529
)

type TaxStatus string

const (
	TaxStatusTaxable  TaxStatus = "taxable"
	TaxStatusDeferred TaxStatus = "tax_deferred"
	TaxStatusFree     TaxStatus = "tax_free"
	TaxStatusExempt   TaxStatus = "tax_exempt"
)

// MarshalBSON for Account
func (a *Account) MarshalBSON() ([]byte, error) {
	type Alias Account

	doc := struct {
		*Alias `bson:",inline"`
	}{
		Alias: (*Alias)(a),
	}

	return bson.Marshal(doc)
}

// UnmarshalBSON for Account
func (a *Account) UnmarshalBSON(data []byte) error {

	// slog.Debug("Account", "UnmarshalBSOn", data)
	// First pass: unmarshal everything except Details
	type Alias Account
	aux := &struct {
		*Alias `bson:",inline"`
		Detail bson.Raw `bson:"detail"`
	}{
		Alias: (*Alias)(a),
	}

	if err := bson.Unmarshal(data, aux); err != nil {
		slog.Debug("Account", "UnmarshalBSOn", err)
		return err
	}
	log.Debug("Account", "UnmarshalBSOn", a)

	// Second pass: unmarshal Details based on AccountType
	var detail AccountDetail
	switch a.Category {
	case CategoryBrokerage, CategoryRetirement, CategoryHSA:
		detail = &BrokerageDetail{}
	case CategoryCrypto:
		detail = &CryptoDetail{}
	case CategoryCash:
		detail = &BankDetail{}
	case Category529:
		detail = &EducationDetail{}
	default:
		return fmt.Errorf("unknown account type: %s", a.Type)
	}

	if err := bson.Unmarshal(aux.Detail, detail); err != nil {
		return err
	}

	a.Detail = detail
	return nil
}

func (a *Account) UnmarshalJSON(data []byte) error {
	// First: unmarshal everything into a map to get accountType
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	slog.Debug("Account-UnmarshalJSON", "raw", raw)

	// Get accountType
	var acategory AccountCategory
	if err := json.Unmarshal(raw["category"], &acategory); err != nil {
		slog.Debug("Account-UnmarshalJSON", "Category", err)
		return err
	}
	slog.Debug("Account", "UnmarshalBSOn", a)

	// Determine detail type
	var detail AccountDetail
	switch acategory {
	case CategoryBrokerage, CategoryRetirement, CategoryHSA:
		detail = &BrokerageDetail{}
	case CategoryCrypto:
		detail = &CryptoDetail{}
	case CategoryCash:
		detail = &BankDetail{}
	case Category529:
		detail = &EducationDetail{}

	// ...
	default:
		return fmt.Errorf("unknown account type: %s", acategory)
	}

	// Unmarshal detail
	if err := json.Unmarshal(raw["detail"], detail); err != nil {
		return err
	}
	slog.Debug("Account-UnmarshalJSON", "detail", detail)

	// Unmarshal other fields (exclude detail)
	delete(raw, "detail") // Remove detail from map

	// Reconstruct JSON without detail field
	tempData, _ := json.Marshal(raw)

	type Alias Account
	if err := json.Unmarshal(tempData, (*Alias)(a)); err != nil {
		return err
	}

	a.Detail = detail
	return nil
}
