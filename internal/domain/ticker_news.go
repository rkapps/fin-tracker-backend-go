package domain

import "time"

// TickerNews holds metadata for news feed
type TickerNews struct {
	ID          string    `json:"id" bson:"id"`
	Symbol      string    `json:"symbol" firestore:"symbol,omitempty"`
	Date        time.Time `json:"date" firestore:"date,omitempty"`
	URL         string    `json:"url" firestore:"url,omitempty"`
	Title       string    `json:"title" firestore:"title,omitempty"`
	Description string    `json:"description" firestore:"description,omitempty"`
	Source      string    `json:"source" firestore:"source,omitempty"`
}

// Id returns the unique id for the ticker
func (ti *TickerNews) Id() string {
	return ti.ID
}

func (ti *TickerNews) CollectionName() string {
	return TICKER_NEWS_COLLECTION_NAME
}
