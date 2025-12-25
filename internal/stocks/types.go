package stocks

import (
	"rkapps/fin-tracker-backend-go/internal/utils"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (

	//Fields
	FIELD_ID                 = "id"
	FIELD_SYMBOL             = "symbol"
	FIELD_EXCHANGE           = "exchange"
	FIELD_NAME               = "name"
	FIELD_SECTOR             = "sector"
	FIELD_INDUSTRY           = "industry"
	FIELD_OVERVIEW           = "overview"
	FIELD_ACTIVE             = "active"
	FIELD_MARKETCAP          = "marketCap"
	FIELD_YIELD              = "yield"
	FIELD_PRDIFFPERC_SEARCH  = "prDiffPercSearch"
	FIELD_PERFORMANCE_SEARCH = "performanceSearch"
	FIELD_DIFF               = "diff"
	FIELD_PRICE              = "price"
	FIELD_STRATEGIES         = "strategies"

	//Collections
	TICKER_CONTROL_COLLECTION_NAME = "ticker_control"
	TICKER_COLLECTION_NAME         = "ticker"
	TICKER_HISTORY_COLLECTION_NAME = "ticker_history"

	//ExNasdaq defines the string NASDAQ
	ExNasdaq string = "NASDAQ"
	//ExNyse defines the string NYSE
	ExNyse string = "NYSE"
	//ExNyseArca defines the string NYSEARCA
	ExNyseArca string = "NYSEARCA"
	//ExCurrency defines the string CURRENCY
	ExCurrency string = "CURRENCY"
	//ExOtc defines the string OTC
	ExOtc string = "OTC"

	//SMA defines the string SMA
	SMA string = "SMA"
	//EMA defines the string EMA
	EMA string = "EMA"
	//RSI defines the string RSI
	RSI string = "RSI"
)

var (
	//PerfPeriods defines the performance periods
	PerfPeriods []string = []string{"1W", "1M", "3M", "6M", "YTD", "1Y", "2Y", "3Y", "5Y"}
	//RSIPeriods defines the RSI periods
	RSIPeriods []int = []int{5, 9, 14, 20, 26}
	//SMAPeriods defines the SMA periods
	SMAPeriods []int = []int{5, 10, 20, 50, 100, 200}
	//EMAPeriods defines the EMA periods
	EMAPeriods []int = []int{5, 9, 12, 21, 26, 50, 200}
)

// Ticker
type TickerControl struct {
	ID                 string     `json:"id" bson:"id"`
	Symbol             string     `json:"symbol" bson:"symbol"`
	Exchange           string     `json:"exchange" bson:"exchange"`
	CreateDate         *time.Time `json:"create_date" bson:"create_date"`
	InactiveDate       *time.Time `json:"inactive_date" bson:"inactive_date"`
	UpdatedDate        *time.Time `json:"updated_date" bson:"updated_date"`
	HistoryUpdatedDate *time.Time `json:"history_updated_date" bson:"history_updated_date"`
}

func (tc *TickerControl) Id() string {
	return tc.ID
}

// SetId sets the unique id for the ticket
func (tc *TickerControl) SetId() {
	tc.ID = tc.Exchange + ":" + tc.Symbol
}

func (tc *TickerControl) CollectionName() string {
	return TICKER_CONTROL_COLLECTION_NAME
}

// Ticker
type Ticker struct {
	ID          string                                `json:"id" bson:"id"`
	Symbol      string                                `json:"symbol" bson:"symbol"`
	Exchange    string                                `json:"exchange" bson:"exchange"`
	Name        string                                `json:"name" bson:"name"`
	Sector      string                                `json:"sector" bson:"sector"`
	Industry    string                                `json:"industry" bson:"industry"`
	Overview    string                                `json:"overview" bson:"overview"`
	Active      bool                                  `json:"active" bson:"active"`
	AvgVolume   int                                   `json:"avgVolume" bson:"avgVolume"`
	DivAmt      float64                               `json:"divAmt" bson:"divAmt"`
	EPS         float64                               `json:"eps" bson:"eps"`
	ExDivDate   *time.Time                            `json:"exDivDate" bson:"exDivDate"`
	MarketCap   int64                                 `json:"marketCap" bson:"marketCap"`
	PayDate     *time.Time                            `json:"payDate" bson:"payDate"`
	PayRatio    float64                               `json:"payRatio" bson:"payRatio"`
	PERatio     float64                               `json:"peRatio" bson:"peRatio"`
	PEGRatio    float64                               `json:"pegRatio" bson:"pegRatio"`
	PBRatio     float64                               `json:"pbRatio" bson:"pbRatio"`
	PSRatio     float64                               `json:"psRatio" bson:"psRatio"`
	PrDate      *time.Time                            `json:"prDate" bson:"prDate"`
	PrOpen      decimal.Decimal                       `json:"prOpen" bson:"prOpen"`
	PrHigh      decimal.Decimal                       `json:"prHigh" bson:"prHigh"`
	PrLow       decimal.Decimal                       `json:"prLow" bson:"prLow"`
	PrClose     decimal.Decimal                       `json:"prClose" bson:"prClose"`
	PrLast      decimal.Decimal                       `json:"prLast" bson:"prLast"`
	PrPrev      decimal.Decimal                       `json:"prPrev" bson:"prPrev"`
	PrDiffAmt   decimal.Decimal                       `json:"prDiffAmt" bson:"prDiffAmt"`
	PrDiffPerc  decimal.Decimal                       `json:"prDiffPerc" bson:"prDiffPerc"`
	Pr52WkHigh  float64                               `json:"pr52WkHigh" bson:"pr52WkHigh"`
	Pr52WkLow   float64                               `json:"pr52WkLow" bsone:"pr52WkLow"`
	Volume      int                                   `json:"volume" bson:"volume"`
	Tpeg1Y      float64                               `json:"tpeg1Y" bson:"tpeg1Y"`
	Yield       float64                               `json:"yield" bson:"yield"`
	Performance map[string]map[string]decimal.Decimal `json:"performance" bson:"performance"`
	Technicals  map[string]map[string]decimal.Decimal `json:"technicals" bson:"technicals"`
	// PrLastSearch float64                       `json:"-" bson:"prLastSearch"`

	Strategies        []string                      `json:"-" bson:"strategies"`
	PrDiffPercSearch  float64                       `json:"-" bson:"prDiffPercSearch"`
	PerformanceSearch map[string]map[string]float64 `json:"-" bson:"performanceSearch"`
}

// Tickers is an array of tickers
type Tickers []*Ticker

// Implement methods for Mongorepository

// Id returns the unique id for the ticker
func (t *Ticker) Id() string {
	return t.ID
}

// SetId sets the unique id for the ticket
func (t *Ticker) SetId() {
	t.ID = t.Exchange + ":" + t.Symbol
}

func (t *Ticker) CollectionName() string {
	return TICKER_COLLECTION_NAME
}

// IsStock returns true if the ticker is a NAsdaq, nyse, nysearca
func (t Ticker) IsStock() bool {
	return strings.Compare(t.Exchange, ExNasdaq) == 0 ||
		strings.Compare(t.Exchange, ExNyse) == 0 ||
		strings.Compare(t.Exchange, ExNyseArca) == 0 ||
		strings.Compare(t.Exchange, ExOtc) == 0
}

// IsCrypto returns true if the ticker is a cryptocurrency
func (t Ticker) IsCrypto() bool {
	return strings.Compare(t.Exchange, "CURRENCY") == 0
}

// IsMutf returns true if the ticker is a Mutual Fund
func (t Ticker) IsMutf() bool {
	return strings.Compare(t.Exchange, "MUTF") == 0
}

// SetPriceDiff sets the price difference between the last price and previous price
func (t *Ticker) SetPriceDiff() {
	t.PrDiffAmt, t.PrDiffPerc = utils.PriceDiff(t.PrLast, t.PrPrev)
}

// TickerSearch defines search criteria
type TickerSearch struct {
	Function     string   `json:"function"`
	Strategies   []string `json:"strategies"`
	Sectors      []string `json:"sectors"`
	Industries   []string `json:"industries"`
	SearchText   string   `json:"searchText"`
	PerfPeriod   string   `json:"perfPeriod"`
	FromPerfPerc float64  `json:"fromPerfPerc"`
	ToPerfPerc   float64  `json:"toPerfPerc"`
	FromYield    float64  `json:"fromYield"`
	ToYield      float64  `json:"toYield"`
	RsiPeriod    string   `json:"rsiPeriod"`
	FromRsi      int      `json:"fromRsi"`
	ToRsi        int      `json:"toRsi"`
	PrAbove      bool     `json:"prAbove"`
	PrMA         string   `json:"prMa"`
	PrPeriod     string   `json:"prPeriod"`
}

// TickerHistory holds historical data for tickers
type TickerHistory struct {
	ID       string    `bson:"id"`
	Date     time.Time `bson:"date"`
	Metadata struct {
		Symbol   string `bson:"symbol"`
		Exchange string `bson:"exchange"`
	} `bson:"metadata"`
	Open        decimal.Decimal            `bson:"open"`
	High        decimal.Decimal            `bson:"high"`
	Low         decimal.Decimal            `bson:"low"`
	Close       decimal.Decimal            `bson:"close"`
	Volume      decimal.Decimal            `bson:"volume"`
	SplitFactor float64                    `bson:"splitFactor"`
	SMA         map[string]decimal.Decimal `bson:"sma"`
	EMA         map[string]decimal.Decimal `bson:"ema"`
	RSI         map[string]decimal.Decimal `bson:"rsi"`
}

// Id returns the unique id for the ticker
func (th *TickerHistory) Id() string {
	return th.ID
}

// SetId sets the unique id for the ticket
func (th *TickerHistory) SetId() {
	th.ID = th.Metadata.Exchange + ":" + th.Metadata.Symbol
}

func (th *TickerHistory) CollectionName() string {
	return TICKER_HISTORY_COLLECTION_NAME
}
