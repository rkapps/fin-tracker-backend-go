package dto

import (
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

type ActivityResponse struct {
	ID              string          `json:"id"`
	TxnType         string          `json:"txnType"`
	Date            *time.Time      `json:"date"`
	RcvAccountName  string          `json:"rcvAccountName"`
	RcvSymbol       string          `json:"rcvSymbol"`
	RcvAmount       decimal.Decimal `json:"rcvAmount"`
	RcvValue        decimal.Decimal `json:"rcvValue"`
	RcvBalance      decimal.Decimal `json:"rcvBalance"`
	SentAccountName string          `json:"sentAccountName"`
	SentSymbol      string          `json:"sentSymbol"`
	SentAmount      decimal.Decimal `json:"sentAmount"`
	SentValue       decimal.Decimal `json:"sentValue"`
	SentBalance     decimal.Decimal `json:"sentBalance"`
	Value           decimal.Decimal `json:"value"`
	Gl_Amount       decimal.Decimal `json:"glAmount"`
	FeeAmount       decimal.Decimal `json:"feeAmount"`
	FeeSymbol       string          `json:"feeSymbol"`
	Notes           string          `json:"notes"`
	Tag             string          `json:"tag"`
}

func NewActivityResponseFromActivity(acct domain.Account, actv domain.Activity) ActivityResponse {
	ractv := ActivityResponse{}
	ractv.ID = actv.ID
	ractv.TxnType = string(actv.TxnType)
	ractv.Date = &actv.Date
	ractv.Notes = actv.Notes
	ractv.FeeAmount = actv.Fee
	ractv.FeeSymbol = actv.FeeCurrency
	ractv.RcvAccountName = acct.Name
	ractv.RcvSymbol = actv.RcvSymbol
	ractv.RcvAmount = actv.RcvQuantity
	ractv.SentSymbol = actv.SentSymbol
	ractv.SentAmount = actv.SentQuantity
	ractv.Notes = actv.Notes

	return ractv
}
