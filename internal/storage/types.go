package storage

import (
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type StorageService interface {

	// Ticker
	DeleteTicker(id string) error
	GetTicker(id string) (*domain.Ticker, error)
	GetTickerGroups() (domain.TickerGroups, error)
	GetTickerEmbeddings(symbol string) ([]*domain.TickerEmbedding, error)
	GetTickerHistory(symbol string) ([]*domain.TickerHistory, error)
	GetTickerSentiments(symbol string) ([]*domain.TickerSentiment, error)
	GetTickers(symbols []string) (domain.Tickers, error)
	SearchTicker(ts domain.TickerSearch) (domain.Tickers, error)

	//Transaction
	ImportTransactions(userId string, startDate time.Time, endDate time.Time, transactions []*domain.Transaction) error
	SearchTransactions(userId string, startDate time.Time, endDate time.Time, searchText string) (domain.Transactions, error)
	SummaryTransactions(userId string, startDate time.Time, endDate time.Time) ([]domain.TransactionAgg, error)

	//User
	GetUser(id string) (*domain.User, error)
	SaveUser(user *domain.User) error
}
