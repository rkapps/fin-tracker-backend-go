package domain

import "time"

// Ticker
type TickerControl struct {
	ID     string `json:"id" bson:"id"`
	Symbol string `json:"symbol" bson:"symbol"`
	// Exchange            string     `json:"exchange" bson:"exchange"`
	// CreateDate          *time.Time `json:"create_date" bson:"create_date"`
	// InactiveDate        *time.Time `json:"inactive_date" bson:"inactive_date"`
	LastSyncAt          *time.Time `json:"last_sync_at" bson:"last_sync_at"`
	LastHistorySyncAt   *time.Time `json:"last_history_sync_at" bson:"last_history_sync_at"`
	LastIndicatorSyncAt *time.Time `json:"last_indicator_sync_at" bson:"last_indicator_sync_at"`
	LastSentimentSyncAt *time.Time `json:"last_sentiment_sync_at" bson:"last_sentiment_sync_at"`
	LastEmbeddingSyncAt *time.Time `json:"last_embedding_sync_at" bson:"last_embedding_sync_at"`
}

func (tc *TickerControl) Id() string {
	return tc.ID
}

func (tc *TickerControl) CollectionName() string {
	return TICKER_CONTROL_COLLECTION_NAME
}
