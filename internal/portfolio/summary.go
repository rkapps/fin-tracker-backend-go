package portfolio

import (
	"fmt"
	"time"

	uuid "github.com/google/uuid"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func (p Portfolio) summarizeData(uid string,
	accts []*domain.Account, actvs []*domain.Activity, lots []*domain.ActivityLot,
) ([]*domain.AccountSummary, error) {

	asumys := []*domain.AccountSummary{}
	user, err := p.userStorage.GetUser(uid)
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
		key := core.GetHoldingsKey(true, *acct, "")
		asummary := asummarym[key]
		if asummary == nil {
			asummary = &domain.AccountSummary{}
			asummary.AccountID = actv.AccountID
			asummary.AccountName = acct.Name
			asummary.AccountDisplayName = acct.ID
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
		} else if actv.IsBuyDeposit() {
			// bought directly using backaccount
			asummary.Deposits = asummary.Deposits.Add(actv.SentAmount)
		} else if actv.IsWithdrawal() {
			asummary.Withdrawals = asummary.Withdrawals.Add(actv.SentAmount)
		}

		asummary.NetDeposits = asummary.Deposits.Sub(asummary.Withdrawals)
	}

	// get holding with symbol to get cash
	hldgs, err := core.GetHoldings(p.tickersStorage, p.logger, false, accts, []string{}, lots)
	if err != nil {
		return asumys, fmt.Errorf("getholdings error: %v", err)
	}

	for _, hldg := range hldgs {

		acct, ok := acctsm[hldg.AcctountID]
		if !ok {
			p.logger.Error("summarizeData", "account not found", hldg.AcctountID)
			continue
		}
		key := core.GetHoldingsKey(true, *acct, "")
		asummary, ok := asummarym[key]
		if !ok {
			continue
		}
		asummary.Category = string(acct.Category)
		asummary.Type = string(acct.Type)
		asummary.AccountID = hldg.AcctountID
		// asummary.AccountDisplayName = hldg.AccountDisplayName
		// asummary.ParentAccountName = hldg.ParentAccountName

		if hldg.Symbol == user.CurrencyCode {
			p.logger.Debug("Cash", "account", hldg.AccountName)
			asummary.Cash = asummary.Cash.Add(hldg.CostValue)
		} else {

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

		asummary.CostValue = asummary.CostValue.Add(hldg.CostValue)
		asummary.MarketValue = asummary.MarketValue.Add(hldg.MktValue)

	}

	for _, asummary := range asummarym {
		asumys = append(asumys, asummary)
	}

	return asumys, nil
}
