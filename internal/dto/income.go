package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type IncomeResponse struct {
	Category   string `json:"category"`
	Type       string `json:"type"`
	AcctountID string `json:"accountId"`
	// ParentAccountName  string          `json:"parentAccountName"`
	AccountName string `json:"accountName"`
	Blockchain  string `json:"blockchain"`
	// AccountDisplayName string          `json:"accountDisplayName"`
	Symbol    string          `json:"symbol"`
	Date      time.Time       `json:"date"`
	Qty       decimal.Decimal `json:"qty"`
	Cost      decimal.Decimal `json:"cost"`
	CostValue decimal.Decimal `json:"costValue"`
}
