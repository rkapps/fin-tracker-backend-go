package portfolios

import (
	"context"
	"rkapps/fin-tracker-backend-go/internal/portfolios/accounts"
	"rkapps/fin-tracker-backend-go/internal/portfolios/user"
)

type Service interface {
	LoadAccounts(ctx context.Context, user user.User, accts accounts.Accounts) error
}
