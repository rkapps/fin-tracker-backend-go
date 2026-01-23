package portfolios

import (
	"context"
	"rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"rkapps/fin-tracker-backend-go/internal/portfolios/user"

	"github.com/rkapps/storage-backend-go/mongodb"
)

type PortfoliosService struct {
	client *mongodb.MongoClient
}

func NewMongoService(client *mongodb.MongoClient) Service {

	return PortfoliosService{
		client: client,
	}
}

func (s PortfoliosService) LoadAccounts(ctx context.Context, user user.User, accts accounts.Accounts) error {

	ids := []string{}
	for _, acct := range accts {
		acct.UID = user.ID
		acct.SetId()
		ids = append(ids, acct.ID)
	}
	acctColl := mongodb.NewMongoRepository[*accounts.Account](*s.client)
	return acctColl.BulkWrite(ctx, ids, accts)
}
