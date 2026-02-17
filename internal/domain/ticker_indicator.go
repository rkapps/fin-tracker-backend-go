package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TickerHistory holds historical data for tickers
type TickerIndicator struct {
	ID       string    `json:"id" bson:"id"`
	Date     time.Time `json:"date" bson:"date"`
	Metadata struct {
		Symbol      string `json:"symbol" bson:"symbol"`
		Type        string `json:"type" bson:"type"`     // "sma", "ema", "rsi", "macd"
		Period      int    `json:"period" bson:"period"` // 50, 200, 14, etc.
		Exchange    string `json:"exchange" bson:"exchange"`
		Granularity string `json:"granularity" bson:"granularity"` // "1d", "1h", "5m", "1m"
	} `json:"metadata" bson:"metadata"`
	Value decimal.Decimal `bson:"value"` // measurement

}

// Id returns the unique id for the ticker
func (ti *TickerIndicator) Id() string {
	return ti.ID
}

func (ti *TickerIndicator) CollectionName() string {
	return TICKER_INDICATOR_COLLECTION_NAME
}
