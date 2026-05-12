package portfolio

import (
	"context"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio/syncer"
	"golang.org/x/sync/errgroup"
)

func (p Portfolio) SyncUserAccounts(ctx context.Context, uid string) error {

	slog.Info("RefreshUser", "UID", uid)
	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return err
	}

	// filter — only exchange and wallet accounts
	var syncable []domain.Account
	for _, account := range accts {
		switch account.Type {
		case domain.TypeExchange, domain.TypeHotWallet:
			syncable = append(syncable, *account)
		}
	}

	// group by provider/chain and fan out
	g, ctx := errgroup.WithContext(ctx)

	for provider, providerAccounts := range groupAccountsByProvider(syncable) {
		providerAccounts := providerAccounts
		provider := provider
		g.Go(func() error {
			syncer, err := syncer.ResolveBatchSyncer(provider)
			if err != nil {
				return err
			}
			return syncer.Sync(ctx, providerAccounts)
		})
	}

	return g.Wait()
}
