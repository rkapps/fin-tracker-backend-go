package portfolio

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio/refresher"
	"golang.org/x/sync/errgroup"
)

func (p Portfolio) RefreshUserAccounts(ctx context.Context, uid string, simulate bool) error {

	user, err := p.storage.GetUser(uid)
	if err != nil {
		return fmt.Errorf("User record does not exist")
	}

	p.logger.Trace("RefreshUserAccounts", "UID", uid)
	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return fmt.Errorf("error getting user accounts: %v", err)
	}

	actvs, err := p.refreshUserActivities(ctx, accts)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "Error", err)
		return fmt.Errorf("error refreshing user activities")
	}

	p.logger.Info("RefreshUserAccounts", "Activities", len(actvs))
	gl := NewGainLoss(accts, user.LotMatchingMethod, simulate)

	glResult, err := gl.Run(ctx, actvs)

	p.saveData(uid, actvs, glResult.Lots)
	// gain loss here
	return nil
}

func (p Portfolio) refreshUserActivities(ctx context.Context, accts []*domain.Account) ([]*domain.Activity, error) {

	// split accounts by pattern — per account vs per type batch
	var (
		singleAccounts []domain.Account
		batchAccounts  []domain.Account
	)

	for _, account := range accts {
		if account.Category == domain.CategoryCash {
			continue
		}
		switch account.Type {
		case domain.TypeExchange, domain.TypeHotWallet:
			batchAccounts = append(batchAccounts, *account)
		default:
			singleAccounts = append(singleAccounts, *account)
		}
	}

	var (
		mu         sync.Mutex
		activities []*domain.Activity
	)

	g, ctx := errgroup.WithContext(ctx)

	// fan out per account — brokerage, wallet, imported
	for _, account := range singleAccounts {
		account := account
		g.Go(func() error {
			refresher, err := refresher.ResolveRefresher(p.storage, account)
			if err != nil {
				return err
			}
			slog.Debug("refreshUserActivities", "refresher", refresher, "error", err)
			result, err := refresher.Refresh(ctx, account)

			if err != nil {
				return err
			}
			mu.Lock()
			activities = append(activities, result...)
			mu.Unlock()
			return nil
		})
	}

	// one call for all exchange accounts — grouped by exchange type
	// TODO: group exchangeAccounts by exchange provider (e.g. Coinbase, Binance)
	//       each provider gets one Refresh(ctx, []account) call
	for provider, providerAccounts := range groupAccountsByProvider(batchAccounts) {
		providerAccounts := providerAccounts
		provider := provider
		g.Go(func() error {
			refresher, err := refresher.ResolveBatchRefresher(provider)
			if err != nil {
				return err
			}
			result, err := refresher.Refresh(ctx, provider, providerAccounts)
			if err != nil {
				return err
			}
			mu.Lock()
			activities = append(activities, result...)
			mu.Unlock()
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	return activities, nil
}

func (p Portfolio) saveData(uid string, actvs []*domain.Activity, lots []*domain.ActivityLot) error {

	// get and delete activities
	oactvs, _ := p.storage.GetActivities(uid)
	ids := []string{}
	for _, oactv := range oactvs {
		ids = append(ids, oactv.ID)
	}
	p.storage.DeleteActivities(ids)

	// clear and delete activitylogs
	clear(ids)
	olots, _ := p.storage.GetActivityLots(uid)
	for _, olot := range olots {
		ids = append(ids, olot.ID)
	}
	p.storage.DeleteActivityLots(ids)

	p.storage.SaveActivities(actvs)
	p.storage.SaveActivityLots(lots)

	return nil
}
