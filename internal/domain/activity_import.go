package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type ActivityImport struct { // Fixed capitalization
	ID        string `json:"id" bson:"_id"`
	UID       string `json:"-" bson:"uid"`
	AccountID string `json:"accountId" bson:"accountId"`

	// // Source metadata
	// ImportBatchID string    `json:"importBatchId,omitempty" bson:"importBatchId,omitempty"`
	// ImportedAt    time.Time `json:"importedAt" bson:"importedAt"`
	// Source        string    `json:"source" bson:"source"`
	// Transaction data
	Hash      string     `json:"hash,omitempty" bson:"hash,omitempty"`
	TxnWallet string     `json:"txnWallet,omitempty" bson:"txnWallet,omitempty"`
	TxnType   string     `json:"txnType" bson:"txnType"`
	Date      *time.Time `json:"date" bson:"date"`

	RcvAccount  string          `json:"rcvAccount,omitempty" bson:"rcvAccount,omitempty"`
	RcvAddress  string          `json:"rcvAddress,omitempty" bson:"rcvAddress,omitempty"`
	RcvCurrency string          `json:"rcvCurrency,omitempty" bson:"rcvCurrency,omitempty"`
	RcvAmount   decimal.Decimal `json:"rcvAmount" bson:"rcvAmount"`

	SentAccount  string          `json:"sentAccount,omitempty" bson:"sentAccount,omitempty"`
	SentAddress  string          `json:"sentAddress,omitempty" bson:"sentAddress,omitempty"`
	SentCurrency string          `json:"sentCurrency,omitempty" bson:"sentCurrency,omitempty"`
	SentAmount   decimal.Decimal `json:"sentAmount" bson:"sentAmount"`
	SentPrice    decimal.Decimal `json:"sentPrice,omitempty" bson:"sentPrice,omitempty"`
	SentBalance  decimal.Decimal `json:"sentBalance,omitempty" bson:"sentBalance,omitempty"`

	GlAmount    decimal.Decimal `json:"glAmount,omitempty" bson:"glAmount,omitempty"`
	Fee         decimal.Decimal `json:"fee,omitempty" bson:"fee,omitempty"`
	FeeCurrency string          `json:"feeCurrency,omitempty" bson:"feeCurrency,omitempty"`
	Notes       string          `json:"notes,omitempty" bson:"notes,omitempty"`

	// // Processing
	// ProcessedID   string `json:"processedId,omitempty" bson:"processedId,omitempty"`
	// ProcessStatus string `json:"processStatus" bson:"processStatus"`
	// ErrorMessage  string `json:"errorMessage,omitempty" bson:"errorMessage,omitempty"`
}

// Id returns the unique id for the ticker
func (a *ActivityImport) Id() string {
	return a.ID
}

func (a *ActivityImport) CollectionName() string {
	return ACTIVITY_IMPORT_COLLECTION_NAME
}
