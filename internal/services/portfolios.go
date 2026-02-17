package services

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type PortfoliosService struct {
	storage storage.StorageService
}

func NewPortfoliosService(storage storage.StorageService) PortfoliosService {
	return PortfoliosService{storage: storage}
}

func (s PortfoliosService) GetUser(id string) (*domain.User, error) {
	return s.storage.GetUser(id)
}

func (s PortfoliosService) LoadAccounts(ctx context.Context, user domain.User, accts accounts.Accounts) error {

	ids := []string{}
	for _, acct := range accts {
		acct.UID = user.ID
		acct.SetId()
		ids = append(ids, acct.ID)
	}
	// acctColl := mongodb.NewMongoRepository[*accounts.Account](*s.client)
	// return acctColl.BulkWrite(ctx, ids, accts)
	return nil
}
