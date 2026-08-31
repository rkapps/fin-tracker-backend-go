package portfolio

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"golang.org/x/sync/errgroup"
)

func (p Portfolio) SyncUserAccounts(ctx context.Context, uid string) error {

	p.logger.Info("SyncUserAccounts", "UID", uid)

	accts, err := p.accountsStorage.GetAccounts(uid)
	if err != nil {
		return err
	}

	// get syncable accounts
	saccts := domain.Accounts{}
	for _, acct := range accts {
		saccts = append(saccts, acct)

		// astate, _ := p.accountsStorage.GetAccountSyncState(acct.UID, acct.ID)
		// if astate != nil && astate.Refresh {
		// 	saccts = append(saccts, acct)
		// }
	}

	acreds := p.groupAccountsWithCredentialsByProvider(saccts)
	p.logger.Info("SyncUserAccounts", "Grouped Accounts", len(acreds))

	// group by provider/chain and fan out
	g, ctx := errgroup.WithContext(ctx)

	for provider, providerAccounts := range acreds {
		p.logger.Info("SyncUserAccounts", "Provider", provider, "Accounts", len(providerAccounts))
		// providerAccounts := providerAccounts
		// provider := provider
		g.Go(func() error {
			syncer, err := ResolveBatchSyncer(p.syncRegistry, provider, p.logConfig, p.logger)
			if err != nil {
				// p.logger.Error("SyncUserAccounts", "ResolveBatchSyncer", err)
				// return err
				return nil
			}
			return syncer.Sync(ctx, p.providerStorage, providerAccounts)
		})
	}

	return g.Wait()
}

// AccountSyncer — the only syncer interface needed.
// Exchange and Wallet always sync as a batch by provider/chain.
type AccountSyncer interface {
	Sync(ctx context.Context, storage storage.ProviderStorageService, acreds []domain.AccountWithCredential) error
}

func ResolveBatchSyncer(
	syncRegistry core.SyncRegistry,
	provider string,
	logConfig *logger.Config,
	logger *logger.Logger,
) (AccountSyncer, error) {

	slog := logConfig.For("portfolio.syncer")
	sProvider, err := syncRegistry.Get(provider)

	if err != nil {
		// slog.Error("ResolveBatchSyner", "RegistryError", err)
		return nil, err
	}
	return &accountSyncProvider{
		syncProvider: sProvider,
		logConfig:    logConfig,
		logger:       slog,
	}, nil
}

// The generic implementation — one for ALL providers.
type accountSyncProvider struct {
	syncProvider core.SyncProvider // Name, Streams, FetchRaws
	logConfig    *logger.Config
	logger       *logger.Logger
}

func (s *accountSyncProvider) Sync(ctx context.Context, providerStorage storage.ProviderStorageService, acreds []domain.AccountWithCredential) error {
	var errs []error

	// s.logger.Info("Sync", "Accounts", len(acreds))
	if global, ok := s.syncProvider.(core.GlobalSyncProvider); ok {
		s.logger.Debug("Sync", "GlobalProvider", s.syncProvider.Name(), "Stream", global.GlobalStreams())
		for _, stream := range global.GlobalStreams() {
			if strings.Compare("ethereum", s.syncProvider.Name()) != 0 {
				continue
			}
			if err := s.syncGlobalStream(ctx, providerStorage, global, stream); err != nil {
				errs = append(errs, fmt.Errorf("%s/global/%s: %w", s.syncProvider.Name(), stream, err))
			}
		}
	}

	for _, acred := range acreds { // serial: shared rate limit / nonce per provider
		s.logger.Debug("Sync", "Account", acred.Account.ID)
		if strings.Compare("ethereum", s.syncProvider.Name()) != 0 {
			continue
		}
		for _, stream := range s.syncProvider.Streams() {
			s.logger.Debug("Sync", "Provider-Stream", fmt.Sprintf("%s-%s", s.syncProvider.Name(), stream))
			if err := s.syncStream(ctx, providerStorage, acred, stream); err != nil {
				errs = append(errs, fmt.Errorf("%s/%s/%s: %w",
					s.syncProvider.Name(), acred.Account.ID, stream, err))
				// keep going — isolate account failures
			}
		}
	}
	return errors.Join(errs...)
}

func (s *accountSyncProvider) syncGlobalStream(
	ctx context.Context,
	storage storage.ProviderStorageService,
	global core.GlobalSyncProvider,
	stream string,
) error {

	cursor, _ := storage.LoadCursor(ctx, s.syncProvider.Name(), s.syncProvider.Name(), s.syncProvider.Name(), stream) // no UID/AccountID — global data
	if cursor == nil {
		cursor = &domain.SyncCursor{
			UID:       s.syncProvider.Name(),
			AccountID: s.syncProvider.Name(),
			ID:        s.syncProvider.Name(),
			Provider:  s.syncProvider.Name(),
			Stream:    stream,
		}
	}

	for {
		page, err := global.FetchGlobalRaw(ctx, stream, cursor.Cursor)
		if err != nil {
			s.logger.Error("syncGlobalStream", "FetchGlobalRaw", err)
			return err
		}
		s.logger.Info("syncGlobalStream", "Provider-Stream", fmt.Sprintf("%s-%s", s.syncProvider.Name(), stream), "Items", len(page.Items))

		if len(page.Items) > 0 {
			if err := storage.UpsertRaw(ctx, "", "", s.syncProvider.Name(), page.Items); err != nil {
				s.logger.Error("syncGlobalStream", "UpsertRaw", err)
				return err
			}
		}

		cursor.Cursor = page.NextCursor
		cursor.BackfillDone = !page.HasMore
		cursor.LastSyncedAt = time.Now()

		if err := storage.SaveCursor(ctx, cursor); err != nil {
			s.logger.Error("syncGlobalStream", "SaveCursor", err)
			return err
		}
		if !page.HasMore {
			return nil
		}
	}
}

func (s *accountSyncProvider) syncStream(ctx context.Context,
	storage storage.ProviderStorageService,
	acred domain.AccountWithCredential, stream string,
) error {

	cursor, _ := storage.LoadCursor(ctx, acred.Account.UID, acred.Account.ID, s.syncProvider.Name(), stream)
	// create a default cursor if not available
	if cursor == nil {
		cursor = &domain.SyncCursor{
			UID:       acred.Account.UID,
			ID:        acred.Account.ID,
			AccountID: acred.Account.ID,
			Provider:  s.syncProvider.Name(),
			Stream:    stream,
		}
	}

	for {
		page, err := s.syncProvider.FetchRaw(ctx, acred, stream, cursor.Cursor)
		if err != nil {
			s.logger.Error("Sync", "FetchRaw", err)
			return err
		}
		s.logger.Info("syncStream", "Provider-Stream", fmt.Sprintf("%s-%s", s.syncProvider.Name(), stream), "Items", len(page.Items))
		if len(page.Items) > 0 {
			if err := storage.UpsertRaw(ctx, acred.Account.UID, acred.Account.ID, s.syncProvider.Name(), page.Items); err != nil {
				s.logger.Error("Sync", "UpsertRaw", err)
				return err
			}
		}

		cursor.Cursor = page.NextCursor // update the string field, not the whole struct
		cursor.BackfillDone = !page.HasMore
		cursor.LastSyncedAt = time.Now()

		if err := storage.SaveCursor(ctx, cursor); err != nil {
			s.logger.Error("Sync", "SaveCursor", err)
			return err
		}
		if !page.HasMore {
			return nil
		}
	}
}
