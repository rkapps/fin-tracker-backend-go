package syncer

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

// BatchAccountSyncer — the only syncer interface needed.
// Exchange and Wallet always sync as a batch by provider/chain.
type BatchAccountSyncer interface {
	Sync(ctx context.Context, accounts []domain.Account) error
}
