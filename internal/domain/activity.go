package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type Activity struct {
	ID        string `json:"id"        bson:"_id"`
	UID       string `json:"-"         bson:"uid"`
	AccountID string `json:"accountId" bson:"accountId"`

	// identity
	TxnType ActivityType   `json:"txnType"  bson:"txnType"`
	Date    time.Time      `json:"date"     bson:"date"`
	Status  ActivityStatus `json:"status" bson:"status"` // pending, settled, cancelled

	// source traceability
	SourceID   string `json:"sourceId"   bson:"sourceId"`   // ID from broker/exchange/chain
	SourceType string `json:"sourceType" bson:"sourceType"` // "import", "api", "manual"

	// asset — what was transacted
	RcvSymbol   string          `json:"rcvSymbol"   bson:"rcvSymbol"`
	RcvQuantity decimal.Decimal `json:"rcvQuantity" bson:"rcvQuantity"`
	RcvPrice    decimal.Decimal `json:"rcvPrice"    bson:"rcvPrice"`  // price per unit at time of txn
	RcvAmount   decimal.Decimal `json:"rcvAmount"   bson:"rcvAmount"` // quantity * price
	RcvAccount  string          `json:"rcvAccount,omitempty" bson:"rcvAccount,omitempty"`
	RcvBalance  decimal.Decimal `json:"rcvBalance"   bson:"rcvBalance"` // quantity * price

	// consideration — what was exchanged
	// for buy: cash out. for sell: cash in. for trade: asset exchanged
	SentSymbol   string          `json:"sentSymbol"   bson:"sentSymbol"`
	SentQuantity decimal.Decimal `json:"sentQuantity" bson:"sentQuantity"`
	SentPrice    decimal.Decimal `json:"sentPrice"    bson:"sentPrice"`
	SentAmount   decimal.Decimal `json:"sentAmount"   bson:"sentAmount"`
	SentBalance  decimal.Decimal `json:"sentBalance"   bson:"sentBalance"` // quantity * price
	SentAccount  string          `json:"sentAccount,omitempty" bson:"sentAccount,omitempty"`

	// Value
	Value decimal.Decimal `json:"value"   bson:"value"` // actual value of the activity

	// costs
	Fee         decimal.Decimal `json:"fee"            bson:"fee"`
	FeeCurrency string          `json:"feeCurrency"    bson:"feeCurrency"`
	Commission  decimal.Decimal `json:"commission"     bson:"commission"`
	Tax         decimal.Decimal `json:"tax"            bson:"tax"` // foreign tax, withholding
	TaxCurrency string          `json:"taxCurrency"    bson:"taxCurrency"`

	// transfer routing — for internal account movements
	RcvAccountID  string `json:"rcvAccountId" bson:"rcvAccountId"`
	SentAccountID string `json:"sentAccountId"   bson:"sentAccountId"`

	// type-specific detail
	Detail ActivityDetail `json:"detail,omitempty"  bson:"detail,omitempty"`

	Notes string `json:"notes" bson:"notes"`
}

// Id returns the unique id for the ticker
func (a *Activity) Id() string {
	return a.ID
}

func (a *Activity) CollectionName() string {
	return ACTIVITY_COLLECTION_NAME
}

func (a *Activity) Debug() string {
	return fmt.Sprintf("%s-%s", a.TxnType, a.ID)
}

type ActivityType string

const (
	// trades
	ActivityTypeBuy   ActivityType = "buy"
	ActivityTypeSell  ActivityType = "sell"
	ActivityTypeTrade ActivityType = "trade" // crypto swap / FX

	// income
	ActivityTypeDividend ActivityType = "dividend"
	ActivityTypeInterest ActivityType = "interest"
	ActivityTypeIncome   ActivityType = "income" // staking, rewards, cashback

	// corporate actions
	ActivityTypeSplit   ActivityType = "split"
	ActivityTypeMerger  ActivityType = "merger"
	ActivityTypeSpinoff ActivityType = "spinoff"
	ActivityTypeReturn  ActivityType = "return_of_capital"

	// cash movements
	ActivityTypeDeposit  ActivityType = "deposit"
	ActivityTypeWithdraw ActivityType = "withdraw"
	ActivityTypeRollover ActivityType = "rollover"
	ActivityTypeTransfer ActivityType = "transfer"

	// costs
	ActivityTypeFee        ActivityType = "fee"
	ActivityTypeTax        ActivityType = "tax" // foreign withholding
	ActivityTypeCommission ActivityType = "commission"
)

type ActivityStatus string

const (
	ActivityStatusPending   ActivityStatus = "pending"
	ActivityStatusSettled   ActivityStatus = "settled"
	ActivityStatusCancelled ActivityStatus = "cancelled"
)

func (a Activity) IsIncome() bool {
	return a.TxnType == ActivityTypeDividend || a.TxnType == ActivityTypeInterest
}
