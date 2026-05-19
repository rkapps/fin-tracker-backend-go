package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type AccountSummary struct {
	ID                string
	UID               string
	AccountID         string `json:"accountId"`
	Date              time.Time
	AccountName       string `json:"accountName"`
	Category          string `json:"category"`
	Type              string `json:"type"`
	ParentAccountName string `json:"parentAccountName"`
	Deposits          decimal.Decimal
	Withdrawals       decimal.Decimal
	NetDeposits       decimal.Decimal
	Income            decimal.Decimal
	Realizedgl        decimal.Decimal
	Cash              decimal.Decimal
	SectorHldgs       map[string]*AccountSummaryValue
	AssetTypeHlgds    map[string]*AccountSummaryValue
	CostValue         decimal.Decimal
	MarketValue       decimal.Decimal
}

type AccountSummaryValue struct {
	CostValue decimal.Decimal `json:"costValue"`
	MktValue  decimal.Decimal `json:"mktValue"`
}

// Id returns the unique id for the ticker
func (a *AccountSummary) Id() string {
	return a.ID
}

func (a *AccountSummary) CollectionName() string {
	return ACCOUNT_SUMMARY_COLLECTION_NAME
}
