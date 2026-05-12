package syncer

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type TestBatchAccountSyncer struct{}

func (t TestBatchAccountSyncer) Sync(ctx context.Context, accts []domain.Account) error {
	return nil
}

func ResolveBatchSyncer(provider string) (BatchAccountSyncer, error) {
	return TestBatchAccountSyncer{}, nil
}
