package refresher

import (
	"context"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type TestAccountRefresher struct {
}

func (t TestAccountRefresher) Refresh(ctx context.Context, account domain.Account) ([]*domain.Activity, error) {
	actvs := []*domain.Activity{}
	return actvs, nil
}

type TestBatchAccountRefresher struct{}

func (t TestBatchAccountRefresher) Refresh(ctx context.Context, provider string, accounts []domain.Account) ([]*domain.Activity, error) {
	actvs := []*domain.Activity{}
	return actvs, nil
}

func ResolveRefresher(storage storage.StorageService, account domain.Account) (AccountRefresher, error) {
	slog.Debug("ResolveRefresher", "Account Cateogory", account.Category)
	switch account.Category {
	case domain.CategoryBrokerage, domain.CategoryRetirement:
		return ImportedAccountRefresher{storage}, nil
	}
	return TestAccountRefresher{}, nil
}

func ResolveBatchRefresher(provider string) (BatchAccountRefresher, error) {
	return TestBatchAccountRefresher{}, nil
}
