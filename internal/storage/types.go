package storage

import (
	"context"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

type AccountStorageService interface {
	GetAccount(uid, id string) (*domain.Account, error)
	GetAccounts(uid string) (domain.Accounts, error)
	SaveAccount(acct *domain.Account) error
	DeleteAccount(uid, id string) error
	DeleteAccountCredential(uid, id string) error
	GetAccountCredentials(uid string) ([]*domain.AccountCredential, error)
	GetAccountCredential(uid, id string) (*domain.AccountCredential, error)
	SaveAccountCredential(cred *domain.AccountCredential) error

	GetAccountState(uid, id string) (*domain.AccountState, error)
	SaveAccountState(state *domain.AccountState) error

	GetAccountSummaries(uid string) ([]*domain.AccountSummary, error)
	SaveAccountSummaries(summaries []*domain.AccountSummary) error
	DeleteAccountSummaries(ids []string) error

	// Activities
	GetActivities(uid string) ([]*domain.Activity, error)
	GetActivitiesForAccount(uid, acctID string) ([]*domain.Activity, error)
	SaveActivities(acts []*domain.Activity) error
	DeleteActivities(ids []string) error

	GetActivityAdjs(uid string) ([]*domain.ActivityAdj, error)
	GetActivityLots(uid string) ([]*domain.ActivityLot, error)
	GetActivityLotsForAccount(uid, acctID string) ([]*domain.ActivityLot, error)
	SaveActivityLots(lots []*domain.ActivityLot) error
	DeleteActivityLots(ids []string) error

	GetImportedActivities(uid, acctID string) ([]*domain.ActivityImport, error)
	SaveImportedActivities(acts []*domain.ActivityImport) error
	DeleteImportedActivities(ids []string) error

	// Glentry
	DeleteGlEntries(ids []string) error
	GetGlEntries(uid string) ([]*domain.GLEntry, error)
	SaveGlEntries(glEntries []*domain.GLEntry) error
}

type TransactionStorageService interface {
	ImportTransactions(uid string, start, end time.Time, txns []*domain.Transaction) error
	SearchTransactions(uid string, start, end time.Time, searchText string) (domain.Transactions, error)
	SummaryTransactions(uid string, start, end time.Time) ([]domain.TransactionAgg, error)
}

type UserStorageService interface {
	GetUsers() []*domain.User
	GetUser(id string) (*domain.User, error)
	SaveUser(user *domain.User) error
}

type ProviderStorageService interface {
	DeleteAllCursors(ctx context.Context, uid, id string) error
	DeleteAllRawItems(ctx context.Context, uid, id string) error
	LoadAllCursors(ctx context.Context, uid, acctID string) ([]*domain.SyncCursor, error)
	LoadAllRawItems(ctx context.Context, uid, acctID string) ([]*domain.RawItem, error)

	LoadCursor(ctx context.Context, uid, acctID, provider, stream string) (*domain.SyncCursor, error)
	SaveCursor(ctx context.Context, cursor *domain.SyncCursor) error
	UpsertRaw(ctx context.Context, uid, acctID, provider string, items []domain.RawItem) error
	UnprocessedRaw(ctx context.Context, uid, acctID string, transformVersion int) ([]domain.RawItem, error)
	MarkProcessed(ctx context.Context, rawIDs []string, transformVersion int) error
}

type CryptoStorageService interface {
	GetCryptoPrices() ([]*domain.CryptoPrice, error)
	GetCryptoSpams() ([]*domain.CryptoSpam, error)
	SaveCryptoPrices([]*domain.CryptoPrice) error
}

type TickerStorageService interface {

	// Ticker
	DeleteTicker(id string) error
	GetTicker(id string) (*domain.Ticker, error)
	GetTickerGroups() (domain.TickerGroups, error)
	GetTickerEmbeddings(symbol string) ([]*domain.TickerEmbedding, error)
	GetTickerHistory(symbol string) ([]*domain.TickerHistory, error)
	GetTickerHistoryByDate(symbol string, date time.Time) (*domain.TickerHistory, error)
	GetTickerSentiments(symbol string) ([]*domain.TickerSentiment, error)
	GetTickers(symbols []string) (domain.Tickers, error)
	GetTickerPrice(symbol string) (decimal.Decimal, error)
	SearchTicker(ts domain.TickerSearch) (domain.Tickers, error)
}
