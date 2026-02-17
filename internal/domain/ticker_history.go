package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TickerHistory holds historical data for tickers
type TickerHistory struct {
	ID       string    `json:"id" bson:"id"`
	Date     time.Time `json:"date" bson:"date"`
	Metadata struct {
		Symbol      string `json:"symbol" bson:"symbol"`
		Exchange    string `json:"exchange" bson:"exchange"`
		Granularity string `json:"granularity" bson:"granularity"`
	} `json:"metadata" bson:"metadata"`
	Open   decimal.Decimal `json:"open" bson:"open"`
	High   decimal.Decimal `json:"high" bson:"high"`
	Low    decimal.Decimal `json:"low" bson:"low"`
	Close  decimal.Decimal `json:"close" bson:"close"`
	Volume decimal.Decimal `json:"volume" bson:"volume"`
}

// Id returns the unique id for the ticker
func (th *TickerHistory) Id() string {
	return th.ID
}

func (th *TickerHistory) CollectionName() string {
	return TICKER_HISTORY_COLLECTION_NAME
}
