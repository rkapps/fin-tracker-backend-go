package domain

import "time"

// TickerHistory holds historical data for tickers
type TickerAlpha struct {
	ID   string    `json:"id" bson:"id"`
	Key  string    `json:"key" bson:"key"`
	N    int       `json:"n" bson:"n"`
	Date time.Time `json:"date" bson:"date"`
}

// Id returns the unique id for the ticker
func (ti *TickerAlpha) Id() string {
	return ti.ID
}

func (ti *TickerAlpha) CollectionName() string {
	return TICKER_ALPHA_COLLECTION_NAME
}
