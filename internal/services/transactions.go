package services

import (
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type TransactionsService struct {
	storage storage.StorageService
}

func NewTransactionsService(storage storage.StorageService) TransactionsService {
	return TransactionsService{storage: storage}
}

func (s TransactionsService) SearchTransactions(user domain.User, startDate time.Time, endDate time.Time, searchText string) (domain.Transactions, error) {
	return s.storage.SearchTransactions(user.ID, startDate, endDate, searchText)
}

func (s TransactionsService) ImportTransactions(user domain.User, startDate time.Time, endDate time.Time, txns []*domain.Transaction) error {
	return s.storage.ImportTransactions(user.ID, startDate, endDate, txns)
}

func (s TransactionsService) SummaryTransactions(user domain.User, startDate time.Time, endDate time.Time) ([]domain.TransactionAgg, error) {
	return s.storage.SummaryTransactions(user.ID, startDate, endDate)
}
