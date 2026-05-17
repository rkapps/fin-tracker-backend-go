package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type Income struct {
	Category          string          `json:"category"`
	Type              string          `json:"type"`
	Acct_ID           string          `json:"acctId"`
	ParentAccountName string          `json:"parentAccountName"`
	AccountName       string          `json:"accountName"`
	Symbol            string          `json:"symbol"`
	Date              time.Time       `json:"date"`
	Qty               decimal.Decimal `json:"qty"`
	Cost              decimal.Decimal `json:"cost"`
	CostValue         decimal.Decimal `json:"costValue"`
}
