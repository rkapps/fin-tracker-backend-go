package storage

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/mongodb"
)

type MongoStorage struct {
	database *mongodb.MongoDatabase
}

type StorageService interface {
	DeleteTicker(id string) error
	GetUser(id string) (*domain.User, error)
	GetTicker(id string) (*domain.Ticker, error)
	GetTickerGroups() (domain.TickerGroups, error)
	GetTickerEmbeddings(symbol string) ([]*domain.TickerEmbedding, error)
	GetTickerHistory(symbol string) ([]*domain.TickerHistory, error)
	GetTickerSentiments(symbol string) ([]*domain.TickerSentiment, error)
	GetTickers(symbols []string) (domain.Tickers, error)
	SearchTicker(ts domain.TickerSearch) (domain.Tickers, error)
}
