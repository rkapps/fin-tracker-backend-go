package domain

import (
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/shopspring/decimal"
)

type Ticker struct {
	ID                string                                `json:"id" bson:"id"`
	AssetType         string                                `json:"asset_type" bson:"asset_type"`
	Active            bool                                  `json:"active" bson:"active"`
	Symbol            string                                `json:"symbol" bson:"symbol"`
	Exchange          string                                `json:"exchange" bson:"exchange"`
	Name              string                                `json:"name" bson:"name"`
	Sector            string                                `json:"sector" bson:"sector"`
	Industry          string                                `json:"industry" bson:"industry"`
	Overview          string                                `json:"overview" bson:"overview"`
	Country           string                                `json:"country" bson:"country"`
	Currency          string                                `json:"currency" bson:"currency"`
	MarketCap         int64                                 `json:"market_cap" bson:"market_cap"`
	EPS               float64                               `json:"eps" bson:"eps"`
	PERatio           float64                               `json:"pe_ratio" bson:"pe_ratio"`
	PEGRatio          float64                               `json:"peg_ratio" bson:"peg_ratio"`
	PBRatio           float64                               `json:"pb_ratio" bson:"pb_ratio"`
	PSRatio           float64                               `json:"ps_ratio" bson:"ps_ratio"`
	DivAmt            float64                               `json:"dividend_amt" bson:"dividend_amt"`
	Yield             float64                               `json:"yield" bson:"yield"`
	ExDivDate         *time.Time                            `json:"ex_div_date" bson:"ex_div_date"`
	PayDate           *time.Time                            `json:"pay_date" bson:"pay_date"`
	PayRatio          float64                               `json:"pay_ratio" bson:"pay_ratio"`
	PrDate            *time.Time                            `json:"pr_date" bson:"pr_date"`
	PrOpen            decimal.Decimal                       `json:"pr_open" bson:"pr_open"`
	PrHigh            decimal.Decimal                       `json:"pr_high" bson:"pr_high"`
	PrLow             decimal.Decimal                       `json:"pr_low" bson:"pr_low"`
	PrClose           decimal.Decimal                       `json:"pr_close" bson:"pr_close"`
	PrLast            decimal.Decimal                       `json:"pr_last" bson:"pr_last"`
	PrPrev            decimal.Decimal                       `json:"pr_prev" bson:"pr_prev"`
	PrDiffAmt         decimal.Decimal                       `json:"pr_diff_amt" bson:"pr_diff_amt"`
	PrDiffPerc        decimal.Decimal                       `json:"pr_diff_perc" bson:"pr_diff_perc"`
	PrDiffPercSearch  float64                               `json:"-" bson:"pr_diff_perc_search"`
	Pr52WkHigh        decimal.Decimal                       `json:"pr_52_wk_high" bson:"pr_52_wk_high"`
	Pr52WkLow         decimal.Decimal                       `json:"pr_52_wk_low" bson:"pr_52_wk_low"`
	Performance       map[string]map[string]decimal.Decimal `json:"performance" bson:"performance"`
	PerformanceSearch map[string]map[string]float64         `json:"-" bson:"performance_search"`
	Technicals        map[string]map[string]decimal.Decimal `json:"technicals" bson:"technicals"`
	TechnicalsSearch  map[string]map[string]float64         `json:"technicals_search" bson:"technicals_search"`
	AvgVolume         int                                   `json:"avg_volume" bson:"avg_volume"`
	Volume            int                                   `json:"volume" bson:"volume"`
}

// Tickers is an array of tickers
type Tickers []*Ticker

// Implement methods for Mongorepository

// Id returns the unique id for the ticker
func (t *Ticker) Id() string {
	return t.ID
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
