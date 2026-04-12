package mongo

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/storage-backend-go/core"
	"github.com/rkapps/storage-backend-go/mongodb"
)

type MongoStorage struct {
	database *mongodb.MongoDatabase
}

func NewMongoStorage(database *mongodb.MongoDatabase) storage.StorageService {
	return MongoStorage{database}
}

func (s MongoStorage) context() context.Context {
	return context.Background()
}

func (s MongoStorage) users() core.Repository[string, *domain.User] {
	return mongodb.GetMongoRepository[string, *domain.User](s.database)
}

func (s MongoStorage) tickers() core.Repository[string, *domain.Ticker] {
	return mongodb.GetMongoRepository[string, *domain.Ticker](s.database)
}

func (s MongoStorage) tickerHistory() core.Repository[string, *domain.TickerHistory] {
	return mongodb.GetMongoRepository[string, *domain.TickerHistory](s.database)
}

func (s MongoStorage) tickerSentiment() core.Repository[string, *domain.TickerSentiment] {
	return mongodb.GetMongoRepository[string, *domain.TickerSentiment](s.database)
}

func (s MongoStorage) tickerEmbedding() core.Repository[string, *domain.TickerEmbedding] {
	return mongodb.GetMongoRepository[string, *domain.TickerEmbedding](s.database)
}

func (s MongoStorage) transaction() core.Repository[string, *domain.Transaction] {
	return mongodb.GetMongoRepository[string, *domain.Transaction](s.database)
}
