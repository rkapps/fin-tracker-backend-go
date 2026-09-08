package dto

import (
	"github.com/shopspring/decimal"
)

type IncomeResponse struct {
	Category   string            `json:"category"`
	Type       string            `json:"type"`
	Blockchain string            `json:"blockchain"`
	AccountId  string            `json:"accountId"`
	Symbol     string            `json:"symbol"`
	Year       int               `json:"year"`
	Values     []decimal.Decimal `json:"values"`
}
