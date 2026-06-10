package storage

import (
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

// const (
// 	FINTRACKER_DB_NAME = "rustic_finance"
// )

type FinTrackerStorageService interface {

	//Accounts
	DeleteAccount(uid string, id string) error
	DeleteAccountSummaries(ids []string) error
	DeleteActivities(ids []string) error
	DeleteActivityLots(ids []string) error
	DeleteImortedActivities(ids []string) error
	GetAccount(uid string, id string) (*domain.Account, error)
	GetAccounts(uid string) (domain.Accounts, error)
	GetAccountSummaries(uid string) ([]*domain.AccountSummary, error)
	GetAccountCredential(uid string, id string) (*domain.AccountCredential, error)
	GetAccountSyncState(uid string, id string) (*domain.AccountSyncState, error)
	GetActivities(uid string) ([]*domain.Activity, error)
	GetActivitiesForAccount(uid string, acctId string) ([]*domain.Activity, error)
	GetActivityLots(uid string) ([]*domain.ActivityLot, error)
	GetActivityLotsForAccount(uid string, acctId string) ([]*domain.ActivityLot, error)
	GetImortedActivities(uid string, acctId string) ([]*domain.ActivityImport, error)

	SaveAccount(acct *domain.Account) error
	SaveAccountCredential(acct *domain.AccountCredential) error
	SaveAccountSyncState(acct *domain.AccountSyncState) error
	SaveAccountSummaries(asumys []*domain.AccountSummary) error
	SaveImportedActivities(actvs []*domain.ActivityImport) error
	SaveActivities(actvs []*domain.Activity) error
	SaveActivityLots(lots []*domain.ActivityLot) error

	//Transaction
	ImportTransactions(userId string, startDate time.Time, endDate time.Time, transactions []*domain.Transaction) error
	SearchTransactions(userId string, startDate time.Time, endDate time.Time, searchText string) (domain.Transactions, error)
	SummaryTransactions(userId string, startDate time.Time, endDate time.Time) ([]domain.TransactionAgg, error)

	//User
	GetUsers() []*domain.User
	GetUser(id string) (*domain.User, error)
	SaveUser(user *domain.User) error
}

type TickerStorageService interface {

	// Ticker
	DeleteTicker(id string) error
	GetTicker(id string) (*domain.Ticker, error)
	GetTickerGroups() (domain.TickerGroups, error)
	GetTickerEmbeddings(symbol string) ([]*domain.TickerEmbedding, error)
	GetTickerHistory(symbol string) ([]*domain.TickerHistory, error)
	GetTickerSentiments(symbol string) ([]*domain.TickerSentiment, error)
	GetTickers(symbols []string) (domain.Tickers, error)
	GetTickerPrice(symbol string) (decimal.Decimal, error)
	SearchTicker(ts domain.TickerSearch) (domain.Tickers, error)
}
