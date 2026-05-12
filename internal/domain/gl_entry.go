package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

type GLEntry struct {
	ID         string
	UID        string
	AccountID  string
	ActivityID string // traceability back to the source activity

	// classification
	TxnType ActivityType
	GLType  GLType // see below

	// asset
	Currency string
	Quantity decimal.Decimal

	// cost basis — what you paid
	CostBasis        decimal.Decimal // total cost basis for this lot
	CostBasisPerUnit decimal.Decimal

	// proceeds — what you received
	Proceeds        decimal.Decimal // total proceeds from disposal
	ProceedsPerUnit decimal.Decimal

	// gain/loss
	GainLoss      decimal.Decimal // Proceeds - CostBasis
	IsShortTerm   bool            // holding period < 1 year
	HoldingPeriod int             // days held

	// acquisition — for matching to disposal
	AcquiredDate time.Time
	DisposedDate time.Time

	// fees
	Fee         decimal.Decimal
	FeeCurrency string

	// metadata
	Notes string
}

// GLType classifies the financial event
type GLType string

const (
	GLTypeDisposal    GLType = "disposal"    // sell, trade, spend
	GLTypeAcquisition GLType = "acquisition" // buy, receive, earn
	GLTypeIncome      GLType = "income"      // staking, interest, rewards
	GLTypeTransfer    GLType = "transfer"    // non-taxable movement
	GLTypeFee         GLType = "fee"         // deductible fee
)

// Id returns the unique id for the ticker
func (a *GLEntry) Id() string {
	return a.ID
}

func (a *GLEntry) CollectionName() string {
	return ACTIVITY_COLLECTION_NAME
}
