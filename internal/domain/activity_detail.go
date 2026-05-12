package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// ActivityDetail holds type-specific fields
// Each account category + type has its own detail struct
type ActivityDetail interface {
	DetailType() string
}

// BrokerageActivityDetail — stocks, ETFs, options
type BrokerageActivityDetail struct {
	CUSIP       string `json:"cusip"       bson:"cusip"`       // security identifier
	ISIN        string `json:"isin"        bson:"isin"`        // international identifier
	Exchange    string `json:"exchange"    bson:"exchange"`    // NYSE, NASDAQ
	Description string `json:"description" bson:"description"` // security name
	// TODO: add options fields (strike, expiry, putCall) if needed
}

func (b BrokerageActivityDetail) DetailType() string { return "brokerage" }

// DividendActivityDetail — dividend specific
type DividendActivityDetail struct {
	CUSIP          string          `json:"cusip"          bson:"cusip"`
	ISIN           string          `json:"isin"           bson:"isin"`
	GrossAmount    decimal.Decimal `json:"grossAmount"    bson:"grossAmount"`
	ForeignTax     decimal.Decimal `json:"foreignTax"     bson:"foreignTax"`
	ForeignCountry string          `json:"foreignCountry" bson:"foreignCountry"`
	IsQualified    bool            `json:"isQualified"    bson:"isQualified"` // qualified dividend tax treatment
	IsOrdinary     bool            `json:"isOrdinary"     bson:"isOrdinary"`
}

// WalletActivityDetail — on-chain blockchain transactions
type WalletActivityDetail struct {
	Hash        string          `json:"hash"        bson:"hash"`
	FromAddress string          `json:"fromAddress" bson:"fromAddress"`
	ToAddress   string          `json:"toAddress"   bson:"toAddress"`
	Blockchain  string          `json:"blockchain"  bson:"blockchain"`
	GasPrice    decimal.Decimal `json:"gasPrice"    bson:"gasPrice"`
	GasUsed     decimal.Decimal `json:"gasUsed"     bson:"gasUsed"`
	GasTotal    decimal.Decimal `json:"gasTotal"    bson:"gasTotal"` // gasPrice * gasUsed
	// TODO: add NFT fields if needed
	// TODO: add DeFi protocol fields (pool, liquidity) if needed
}

func (w WalletActivityDetail) DetailType() string { return "wallet" }

// ExchangeActivityDetail — centralised exchange transactions
type ExchangeActivityDetail struct {
	Exchange    string `json:"exchange"    bson:"exchange"` // coinbase, binance
	OrderID     string `json:"orderId"     bson:"orderId"`
	TradeID     string `json:"tradeId"     bson:"tradeId"`
	TradingPair string `json:"tradingPair" bson:"tradingPair"` // BTC/USD
	OrderType   string `json:"orderType"   bson:"orderType"`   // market, limit
	// TODO: add staking fields if needed
}

func (e ExchangeActivityDetail) DetailType() string { return "exchange" }

// CorporateActionDetail — splits, mergers, spinoffs
type CorporateActionDetail struct {
	Description   string          `json:"description"  bson:"description"`
	Ratio         decimal.Decimal `json:"ratio"        bson:"ratio"` // split ratio e.g 2:1
	CUSIP         string          `json:"cusip"        bson:"cusip"`
	ISIN          string          `json:"isin"         bson:"isin"`
	NewSymbol     string          `json:"newSymbol"    bson:"newSymbol"` // post merger/spinoff symbol
	OldSymbol     string          `json:"oldSymbol"    bson:"oldSymbol"` // pre merger/spinoff symbol
	EffectiveDate time.Time       `json:"effectiveDate" bson:"effectiveDate"`
}

func (c CorporateActionDetail) DetailType() string { return "corporate_action" }

// TransferActivityDetail — internal account movements
type TransferActivityDetail struct {
	FromAccountID string `json:"fromAccountId" bson:"fromAccountId"`
	ToAccountID   string `json:"toAccountId"   bson:"toAccountId"`
	FromAddress   string `json:"fromAddress"   bson:"fromAddress"` // crypto
	ToAddress     string `json:"toAddress"     bson:"toAddress"`   // crypto
	Reference     string `json:"reference"     bson:"reference"`   // wire ref, check number
}

func (t TransferActivityDetail) DetailType() string { return "transfer" }

// RolloverActivityDetail — retirement account rollovers
type RolloverActivityDetail struct {
	SourceInstitution string    `json:"sourceInstitution" bson:"sourceInstitution"`
	SourceAccountType string    `json:"sourceAccountType" bson:"sourceAccountType"` // 401k, IRA
	RolloverType      string    `json:"rolloverType"      bson:"rolloverType"`      // direct, indirect
	CheckNumber       string    `json:"checkNumber"       bson:"checkNumber"`
	ReceivedDate      time.Time `json:"receivedDate"      bson:"receivedDate"`
}

func (r RolloverActivityDetail) DetailType() string { return "rollover" }

// FeeActivityDetail — fees, commissions, taxes
type FeeActivityDetail struct {
	FeeType           string `json:"feeType"     bson:"feeType"` // adр, foreign_tax, management, custody
	Description       string `json:"description" bson:"description"`
	RelatedActivityID string `json:"relatedActivityId" bson:"relatedActivityId"` // e.g fee against a dividend
}

func (f FeeActivityDetail) DetailType() string { return "fee" }
