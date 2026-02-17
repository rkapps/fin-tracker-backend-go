package domain

// TickerHistory holds historical data for tickers
type TickerEmbedding struct {
	ID            string    `json:"id" bson:"id"`
	Symbol        string    `json:"symbol" bson:"symbol"`
	SentimentId   string    `json:"sentiment_id" bson:"sentiment_id"`
	EmbeddingText string    `json:"embedding_text" bson:"embedding_text"`
	Vector        []float64 `json:"vector" bson:"vector"`
}

// Id returns the unique id for the ticker
func (th *TickerEmbedding) Id() string {
	return th.ID
}

func (th *TickerEmbedding) CollectionName() string {
	return TICKER_EMBEDDING_COLLECTION_NAME
}
