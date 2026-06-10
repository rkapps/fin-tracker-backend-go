package portfolio

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio/refresher"
	"golang.org/x/sync/errgroup"
)

func (p Portfolio) RefreshUserAccounts(ctx context.Context, uid string, simulate bool) error {

	var err error
	user, err := p.storage.GetUser(uid)
	if err != nil {
		return fmt.Errorf("User record does not exist")
	}

	p.logger.Info("RefreshUserAccounts", "UID", uid, "CurrencyCode", user.CurrencyCode)
	p.logger.Trace("RefreshUserAccounts", "UID", uid, "Simulate", simulate)
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
	gl := NewGainLoss(accts, user.LotMatchingMethod, simulate, p.logConfig)

	glResult, err := gl.Run(ctx, actvs)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "Run", err)
		return fmt.Errorf("error running gainloss")
	}

	asumys, err := p.summarizeData(uid, accts, glResult.Actvs, glResult.Lots)
	if err != nil {
		p.logger.Error("RefreshUserAccounts", "SummarizeData", err)
		return fmt.Errorf("error summarizing data")
	}

	if !simulate {
		err = p.saveData(uid, asumys, glResult.Actvs, glResult.Lots)
	}

	// gain loss here
	return err
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
			refresher, err := refresher.ResolveRefresher(p.storage, account, p.logConfig)
			if err != nil {
				return err
			}
			slog.Debug("refreshUserActivities", "refresher", refresher, "error", err)
			result, err := refresher.Refresh(ctx, account, p.logConfig)

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

func (p Portfolio) summarizeData(uid string, accts []*domain.Account, actvs []*domain.Activity, lots []*domain.ActivityLot) ([]*domain.AccountSummary, error) {

	asumys := []*domain.AccountSummary{}
	user, err := p.storage.GetUser(uid)
	if err != nil {
		return asumys, err
	}

	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	asummarym := make(map[string]*domain.AccountSummary)
	for _, actv := range actvs {

		acct, ok := acctsm[actv.AccountID]
		if !ok {
			p.logger.Error("summarizeData", "account not found", actv.AccountID)
			continue
		}
		key := GetHoldingsKey(true, *acct, "")
		asummary := asummarym[key]
		if asummary == nil {
			asummary = &domain.AccountSummary{}
			asummary.AccountID = actv.AccountID
			asummary.AccountName = acct.Name
			asummary.UID = uid
			asummary.Date = time.Now()
			asummary.ID = uuid.New().String()
			asummary.SectorHldgs = make(map[string]*domain.AccountSummaryValue)
			asummary.AssetTypeHlgds = make(map[string]*domain.AccountSummaryValue)
			asummarym[key] = asummary
		}

		if actv.IsIncome() {
			asummary.Income = asummary.Income.Add(actv.RcvAmount)
		} else if actv.IsDeposit() {
			asummary.Deposits = asummary.Deposits.Add(actv.RcvAmount)
		} else if actv.IsWithdrawal() {
			asummary.Withdrawals = asummary.Withdrawals.Add(actv.SentAmount)
		}

		asummary.NetDeposits = asummary.Deposits.Sub(asummary.Withdrawals)
	}

	// get holding with symbol to get cash
	hldgs, err := GetHoldings(p.tstorage, p.logger, false, accts, []string{}, lots)
	if err != nil {
		return asumys, fmt.Errorf("getholdings error: %v", err)
	}

	for _, hldg := range hldgs {

		acct, ok := acctsm[hldg.AcctountID]
		if !ok {
			p.logger.Error("summarizeData", "account not found", hldg.AcctountID)
			continue
		}
		key := GetHoldingsKey(true, *acct, "")
		asummary, ok := asummarym[key]
		if !ok {
			continue
		}
		asummary.Category = string(acct.Category)
		asummary.Type = string(acct.Type)
		asummary.AccountID = hldg.AcctountID
		asummary.ParentAccountName = hldg.ParentAccountName

		if hldg.Symbol == user.CurrencyCode {
			p.logger.Info("Cash", "account", hldg.AccountName)
			asummary.Cash = asummary.Cash.Add(hldg.CostValue)
		} else {
			asummary.CostValue = asummary.CostValue.Add(hldg.CostValue)
			asummary.MarketValue = asummary.MarketValue.Add(hldg.MktValue)

			// add sector map
			sectorSummary, ok := asummary.SectorHldgs[hldg.Sector]
			if !ok {
				sectorSummary = &domain.AccountSummaryValue{}
				asummary.SectorHldgs[hldg.Sector] = sectorSummary
			}
			sectorSummary.CostValue = sectorSummary.CostValue.Add(hldg.CostValue)
			sectorSummary.MktValue = sectorSummary.MktValue.Add(hldg.MktValue)

		}

		// add asset type map
		assetSummary, ok := asummary.AssetTypeHlgds[hldg.AssetType]
		if !ok {
			assetSummary = &domain.AccountSummaryValue{}
			// assetSummary.AssetType = hldg.AssetType
			asummary.AssetTypeHlgds[hldg.AssetType] = assetSummary
		}
		assetSummary.CostValue = assetSummary.CostValue.Add(hldg.CostValue)
		assetSummary.MktValue = assetSummary.MktValue.Add(hldg.MktValue)

	}

	for _, asummary := range asummarym {
		asumys = append(asumys, asummary)
	}

	return asumys, nil
}

func (p Portfolio) saveData(uid string, asumys []*domain.AccountSummary, actvs []*domain.Activity, lots []*domain.ActivityLot) error {

	ids := []string{}

	// clear and delete activitysummaries
	clear(ids)
	oasumys, err := p.storage.GetAccountSummaries(uid)
	if err != nil {
		return err
	}
	for _, oasum := range oasumys {
		ids = append(ids, oasum.ID)
	}
	p.storage.DeleteAccountSummaries(ids)

	clear(ids)
	// get and delete activities
	oactvs, _ := p.storage.GetActivities(uid)
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

	err = p.storage.DeleteActivityLots(ids)
	if err != nil {
		return err
	}

	err = p.storage.SaveAccountSummaries(asumys)
	if err != nil {
		return err
	}

	err = p.storage.SaveActivities(actvs)
	if err != nil {
		return err
	}

	return p.storage.SaveActivityLots(lots)
}
