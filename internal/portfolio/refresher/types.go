package refresher

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

// AccountRefresher refreshes a single account — brokerage, wallet, imported.
type AccountRefresher interface {
	Refresh(ctx context.Context, account domain.Account) ([]*domain.Activity, error)
}

// BatchAccountRefresher refreshes all accounts of a type in one call — exchanges.
type BatchAccountRefresher interface {
	Refresh(ctx context.Context, provider string, accounts []domain.Account) ([]*domain.Activity, error)
}
