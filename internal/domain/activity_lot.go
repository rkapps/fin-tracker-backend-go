package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type ActivityLot struct {
	ID        string `json:"id"      bson:"_id"`
	UID       string `json:"-"       bson:"uid"`
	AccountID string `json:"acctId"  bson:"accountId"`

	// traceability
	ActivityID     string `json:"actvId"       bson:"actvId"`     // acquisition activity
	SellActivityID string `json:"sellActvId"   bson:"sellActvId"` // disposal activity
	LotSeq         int    `json:"lotSeq"       bson:"lotSeq"`     // FIFO ordering

	// asset
	Symbol string     `json:"symbol"  bson:"symbol"`
	Date   *time.Time `json:"date"    bson:"date"` // acquisition date

	// quantity tracking
	OrigQty decimal.Decimal `json:"origQty" bson:"origQty"` // quantity at creation
	Qty     decimal.Decimal `json:"qty"     bson:"qty"`     // remaining quantity

	// cost basis
	Cost      decimal.Decimal `json:"cost"      bson:"cost"`      // cost per unit
	CostValue decimal.Decimal `json:"costValue" bson:"costValue"` // Cost * OrigQty

	// acquisition fee
	Fee decimal.Decimal `json:"fee" bson:"fee"`
	// transfer tracking
	SendQty  decimal.Decimal `json:"sendQty"  bson:"sendQty"`
	SendDate *time.Time      `json:"sendDate" bson:"sendDate"`

	// disposal tracking
	SaleQty   decimal.Decimal `json:"saleQty"   bson:"saleQty"`
	SaleDate  *time.Time      `json:"saleDate"  bson:"saleDate"`
	SalePrice decimal.Decimal `json:"salePrice" bson:"salePrice"`
	SaleFee   decimal.Decimal `json:"saleFee"   bson:"saleFee"`

	// lifecycle
	Status LotStatus `json:"status" bson:"status"`
}

// LotStatus tracks the lifecycle of a lot
type LotStatus string

const (
	LotStatusOpen        LotStatus = "open"        // has remaining quantity
	LotStatusClosed      LotStatus = "closed"      // fully consumed by disposal
	LotStatusTransferred LotStatus = "transferred" // moved to another account
)

// Id returns the unique id for the ticker
func (a *ActivityLot) Id() string {
	return a.ID
}

func (a *ActivityLot) CollectionName() string {
	return ACTIVITY_LOT_COLLECTION_NAME
}

func (a *ActivityLot) Debug() string {
	return fmt.Sprintf("%s-%v-%v", a.Symbol, a.Qty, a.CostValue)
}

// NewLotFromActivity creates an ActivityLot from an Activity.
func NewLotFromActivity(activity Activity) *ActivityLot {
	return &ActivityLot{
		UID:        activity.UID,
		AccountID:  activity.AccountID,
		ActivityID: activity.ID,
		Symbol:     activity.RcvSymbol,
		Date:       &activity.Date,
		OrigQty:    activity.RcvQuantity,
		Qty:        activity.RcvQuantity,
		Cost:       activity.RcvPrice,
		CostValue:  activity.RcvAmount,
		Fee:        activity.Fee,
		Status:     LotStatusOpen,
	}
}

type LotMatchingMethod string

const (
	LotMatchingHIFO LotMatchingMethod = "hifo"
	LotMatchingFIFO LotMatchingMethod = "fifo"
	LotMatchingLIFO LotMatchingMethod = "lifo"
)
