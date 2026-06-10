package mongo

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/storage-backend-go/core"
	"github.com/rkapps/storage-backend-go/mongodb"
)

// FinTracker Mongo Storage
type FinTrackerMongoStorage struct {
	database *mongodb.MongoDatabase
}

func NewFinTrackerMongoStorage(database *mongodb.MongoDatabase) storage.FinTrackerStorageService {
	return FinTrackerMongoStorage{database}
}

func (s FinTrackerMongoStorage) context() context.Context {
	return context.Background()
}

func (s FinTrackerMongoStorage) accounts() core.Repository[string, *domain.Account] {
	return mongodb.GetMongoRepository[string, *domain.Account](s.database)
}

func (s FinTrackerMongoStorage) accountCredentials() core.Repository[string, *domain.AccountCredential] {
	return mongodb.GetMongoRepository[string, *domain.AccountCredential](s.database)
}

func (s FinTrackerMongoStorage) accountSyncStates() core.Repository[string, *domain.AccountSyncState] {
	return mongodb.GetMongoRepository[string, *domain.AccountSyncState](s.database)
}

func (s FinTrackerMongoStorage) accountSummaries() core.Repository[string, *domain.AccountSummary] {
	return mongodb.GetMongoRepository[string, *domain.AccountSummary](s.database)
}

func (s FinTrackerMongoStorage) acitivyImports() core.Repository[string, *domain.ActivityImport] {
	return mongodb.GetMongoRepository[string, *domain.ActivityImport](s.database)
}

func (s FinTrackerMongoStorage) acitivities() core.Repository[string, *domain.Activity] {
	return mongodb.GetMongoRepository[string, *domain.Activity](s.database)
}
func (s FinTrackerMongoStorage) acitivityLots() core.Repository[string, *domain.ActivityLot] {
	return mongodb.GetMongoRepository[string, *domain.ActivityLot](s.database)
}

func (s FinTrackerMongoStorage) transaction() core.Repository[string, *domain.Transaction] {
	return mongodb.GetMongoRepository[string, *domain.Transaction](s.database)
}

func (s FinTrackerMongoStorage) users() core.Repository[string, *domain.User] {
	return mongodb.GetMongoRepository[string, *domain.User](s.database)
}

// Ticker Mongo Storage
type TickerMongoStorage struct {
	database *mongodb.MongoDatabase
}

func NewTickerMongoStorage(database *mongodb.MongoDatabase) storage.TickerStorageService {
	return TickerMongoStorage{database}
}
func (s TickerMongoStorage) context() context.Context {
	return context.Background()
}
func (s TickerMongoStorage) tickers() core.Repository[string, *domain.Ticker] {
	return mongodb.GetMongoRepository[string, *domain.Ticker](s.database)
}

func (s TickerMongoStorage) tickerHistory() core.Repository[string, *domain.TickerHistory] {
	return mongodb.GetMongoRepository[string, *domain.TickerHistory](s.database)
}

func (s TickerMongoStorage) tickerSentiment() core.Repository[string, *domain.TickerSentiment] {
	return mongodb.GetMongoRepository[string, *domain.TickerSentiment](s.database)
}

func (s TickerMongoStorage) tickerEmbedding() core.Repository[string, *domain.TickerEmbedding] {
	return mongodb.GetMongoRepository[string, *domain.TickerEmbedding](s.database)
}
