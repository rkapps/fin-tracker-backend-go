package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// Holding represents a security holding
type HoldingSummary struct {
	Category          string          `json:"category"`
	Type              string          `json:"type"`
	AcctountID        string          `json:"accountId"`
	ParentAccountName string          `json:"parentAccountName"`
	AccountName       string          `json:"accountName"`
	AssetType         string          `json:"asset_type" bson:"asset_type"`
	Sector            string          `json:"sector"`
	Industry          string          `json:"industry"`
	Symbol            string          `json:"symbol"`
	Date              *time.Time      `json:"date"`
	Qty               decimal.Decimal `json:"qty"`
	Cost              decimal.Decimal `json:"cost"`
	CostValue         decimal.Decimal `json:"costValue"`
	PrLast            decimal.Decimal `json:"prLast"`
	PrDiffAmt         decimal.Decimal `json:"prDiffAmt"`
	PrDiffPerc        decimal.Decimal `json:"prDiffPerc"`
	MktValue          decimal.Decimal `json:"mktValue"`
	Dglamount         decimal.Decimal `json:"dglAmount"`
	Glamount          decimal.Decimal `json:"glAmount"`
	Glperc            decimal.Decimal `json:"glPerc"`
}
