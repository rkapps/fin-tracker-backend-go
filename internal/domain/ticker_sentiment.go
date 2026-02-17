package domain

import (
	"time"

	"github.com/shopspring/decimal"
)

// TickerHistory holds historical data for tickers
type TickerSentiment struct {
	ID             string          `json:"id" bson:"id"`
	Symbol         string          `json:"symbol" bson:"symbol"`
	Date           time.Time       `json:"date" bson:"date"`
	Title          string          `json:"title" bson:"title"`
	Url            string          `json:"url" bson:"url"`
	Summary        string          `json:"summary" bson:"summary"`
	Source         string          `json:"source" bson:"source"`
	SourceCategory string          `json:"source_category" bson:"source_category"`
	SourceDomain   string          `json:"source_domain" bson:"source_domain"`
	RelevanceScore decimal.Decimal `json:"relevance_score" bson:"relevance_score"`
	Score          decimal.Decimal `json:"score" bson:"score"`
	Label          string          `json:"label" bson:"label"`
}

// Id returns the unique id for the ticker
func (th *TickerSentiment) Id() string {
	return th.ID
}

func (th *TickerSentiment) CollectionName() string {
	return TICKER_SENTIMENT_COLLECTION_NAME
}
