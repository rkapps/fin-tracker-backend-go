package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

// Holding represents a security holding
type HoldingSummary struct {
	Group             string          `json:"group"`
	Category          string          `json:"category"`
	Acct_ID           string          `json:"acctId"`
	ParentAccountName string          `json:"parentAccountName"`
	AccountName       string          `json:"accountName"`
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
