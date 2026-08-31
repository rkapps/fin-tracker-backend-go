package domain

import (
	"bytes"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rkapps/storage-backend-go/mongodb"
	"github.com/shopspring/decimal"
	"go.mongodb.org/mongo-driver/v2/bson"
)

type GLType string

const (
	GLTypeDisposal GLType = "disposal" // sell, trade, orphan-send
	GLTypeIncome   GLType = "income"   // interest, dividend, reward, staking
)

type TaxTerm string

const (
	TermShort TaxTerm = "short_term"
	TermLong  TaxTerm = "long_term"
)

type GLDetail interface {
	Validate() error
	GetType() GLType
}

type GLEntry struct {
	ID         string          `json:"id" bson:"id"`
	UID        string          `json:"-" bson:"uid"`
	AccountID  string          `json:"accountId" bson:"accountId"`
	ActivityID string          `json:"activityId" bson:"activityId"`
	LotID      string          `json:"lotId" bson:"lotId"`
	TxnType    ActivityType    `json:"txnType" bson:"txnType"`
	GLType     GLType          `json:"glType" bson:"glType"`
	Symbol     string          `json:"symbol" bson:"symbol"`
	Quantity   decimal.Decimal `json:"quantity" bson:"quantity"`
	Detail     GLDetail        `json:"detail" bson:"detail"`
	Notes      string          `json:"notes,omitempty" bson:"notes,omitempty"`
	TaxDate    time.Time       `json:"taxDate" bson:"taxDate"`
	CreatedAt  time.Time       `json:"createdAt" bson:"createdAt"`
}

func (e *GLEntry) Id() string             { return e.ID }
func (e *GLEntry) CollectionName() string { return GL_ENTRY_COLLECTION_NAME }

// ---- Detail variants ----

type GLDisposalDetail struct {
	CostBasis        decimal.Decimal `json:"costBasis" bson:"costBasis"`
	CostBasisPerUnit decimal.Decimal `json:"costBasisPerUnit" bson:"costBasisPerUnit"`
	Proceeds         decimal.Decimal `json:"proceeds" bson:"proceeds"`
	ProceedsPerUnit  decimal.Decimal `json:"proceedsPerUnit" bson:"proceedsPerUnit"`
	GainLoss         decimal.Decimal `json:"gainLoss" bson:"gainLoss"`
	TaxableGainLoss  decimal.Decimal `json:"taxableGainLoss" bson:"taxableGainLoss"`
	Term             TaxTerm         `json:"term" bson:"term"`
	HoldingPeriod    int             `json:"holdingPeriod" bson:"holdingPeriod"`
	AcquiredDate     time.Time       `json:"acquiredDate" bson:"acquiredDate"`
}

func (d *GLDisposalDetail) Validate() error {
	if d.CostBasis.IsZero() && d.Proceeds.IsZero() {
		return fmt.Errorf("disposal detail requires costBasis or proceeds")
	}
	return nil
}
func (d *GLDisposalDetail) GetType() GLType { return GLTypeDisposal }

type GLIncomeDetail struct {
	Income        decimal.Decimal `json:"income" bson:"income"`
	TaxableIncome decimal.Decimal `json:"taxableIncome" bson:"taxableIncome"`
	ReceivedDate  time.Time       `json:"receivedDate" bson:"receivedDate"`
}

func (d *GLIncomeDetail) Validate() error {
	if d.Income.IsZero() {
		return fmt.Errorf("income detail requires a non-zero income amount")
	}
	return nil
}
func (d *GLIncomeDetail) GetType() GLType { return GLTypeIncome }

// ---- Dispatch, mirroring Account.UnmarshalBSON / UnmarshalJSON exactly ----

func newGLDetail(glType GLType) (GLDetail, error) {
	switch glType {
	case GLTypeDisposal:
		return &GLDisposalDetail{}, nil
	case GLTypeIncome:
		return &GLIncomeDetail{}, nil
	default:
		return nil, fmt.Errorf("unknown gl type: %s", glType)
	}
}

func (e *GLEntry) UnmarshalBSON(data []byte) error {
	reg := mongodb.GetBsonRegistryForDecimal()

	type glEntryFields struct {
		ID         string          `bson:"id"`
		UID        string          `bson:"uid"`
		AccountID  string          `bson:"accountId"`
		ActivityID string          `bson:"activityId"`
		LotID      string          `bson:"lotId"`
		TxnType    ActivityType    `bson:"txnType"`
		GLType     GLType          `bson:"glType"`
		Symbol     string          `bson:"symbol"`
		Quantity   decimal.Decimal `bson:"quantity"`
		Notes      string          `bson:"notes,omitempty"`
		CreatedAt  time.Time       `bson:"createdAt"`
		TaxDate    time.Time       `bson:"taxDate"`
	}

	var fields glEntryFields
	dec := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(data)))
	dec.SetRegistry(reg)
	if err := dec.Decode(&fields); err != nil {
		return err
	}

	e.ID = fields.ID
	e.UID = fields.UID
	e.AccountID = fields.AccountID
	e.ActivityID = fields.ActivityID
	e.LotID = fields.LotID
	e.TxnType = fields.TxnType
	e.GLType = fields.GLType
	e.Symbol = fields.Symbol
	e.Quantity = fields.Quantity
	e.Notes = fields.Notes
	e.CreatedAt = fields.CreatedAt
	e.TaxDate = fields.TaxDate

	var raw bson.Raw = data
	detailRaw, err := raw.LookupErr("detail")
	if err != nil {
		return err
	}
	detail, err := newGLDetail(e.GLType)
	if err != nil {
		return err
	}

	dec2 := bson.NewDecoder(bson.NewDocumentReader(bytes.NewReader(detailRaw.Value)))
	dec2.SetRegistry(reg)
	if err := dec2.Decode(detail); err != nil {
		return err
	}
	e.Detail = detail

	return nil
}

func (e *GLEntry) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	var glType GLType
	if err := json.Unmarshal(raw["glType"], &glType); err != nil {
		return err
	}

	detail, err := newGLDetail(glType)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(raw["detail"], detail); err != nil {
		return err
	}

	delete(raw, "detail")
	tempData, _ := json.Marshal(raw)

	type Alias GLEntry
	if err := json.Unmarshal(tempData, (*Alias)(e)); err != nil {
		return err
	}
	e.Detail = detail
	return nil
}
