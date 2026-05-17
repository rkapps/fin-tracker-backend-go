package services

import (
	"context"
	"fmt"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/dto"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/shopspring/decimal"
)

type PortfolioService struct {
	tickersService TickersService
	storage        storage.StorageService
	logConfig      *logger.Config
	logger         *logger.Logger
}

func NewPortfolioService(logConfig *logger.Config, tickersService TickersService, storage storage.StorageService) PortfolioService {
	plog := logConfig.For("portfolio.service")
	return PortfolioService{storage: storage, logConfig: logConfig, logger: plog}
}

func (p PortfolioService) GetHoldings(uid string, category string, atype string, acctIds []string) ([]*dto.HoldingSummary, error) {

	hldgs := []*dto.HoldingSummary{}
	hldgsm := make(map[string]*dto.HoldingSummary)

	acctIdsm := make(map[string]string)
	for _, acctId := range acctIds {
		acctIdsm[acctId] = acctId
	}

	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return hldgs, nil
	}
	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	lots, err := p.storage.GetActivityLots(uid)
	if err != nil {
		return hldgs, nil
	}

	// get tickermap
	tm := GetTickersMapforLots(p.storage, lots)

	for _, lot := range lots {
		if lot.Status != domain.LotStatusOpen {
			continue
		}
		acct := acctsm[lot.AccountID]
		if acct == nil {
			p.logger.Error("GetHoldings - Account not found", "AccountId", lot.AccountID, "LotId", lot.ID)
			// log.Println(lot)
			continue
		}

		filter := filterBankAccount(acctsm, lot.AccountID)
		if !filter {
			continue
		}
		filter = filterAccount(acctIdsm, acct, category, atype, acctIds)
		if !filter {
			continue
		}

		key := fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID, lot.Symbol)
		p.logger.Debug("GetHoldings", "Key", key, "Lot", lot.Qty)

		h := hldgsm[key]

		zero := decimal.NewFromFloat(0.0)
		if h == nil {
			h = &dto.HoldingSummary{}
			h.Category = string(acct.Category)
			h.Type = string(acct.Type)
			h.AccountName = acct.Name
			h.Acct_ID = lot.AccountID
			h.Symbol = lot.Symbol
			h.Qty = zero
			h.Cost = zero
			h.CostValue = zero
			h.MktValue = zero
			hldgs = append(hldgs, h)
			hldgsm[key] = h
		}
		h.Cost = lot.Cost
		h.Qty = h.Qty.Add(lot.Qty)
		h.CostValue = h.CostValue.Add(lot.CostValue)
		if !h.Qty.IsZero() {
			h.Cost = h.CostValue.Div(h.Qty)
		}

		ticker := tm[lot.Symbol]
		if len(ticker.Symbol) == 0 {
			ticker = GetTickerPriceDiff(tm, lot.Symbol)
			tm[lot.Symbol] = ticker
		}

		h.PrLast = ticker.PrLast
		h.PrDiffAmt = ticker.PrDiffAmt
		h.PrDiffPerc = ticker.PrDiffPerc
		h.MktValue = h.MktValue.Add(lot.Qty.Mul(ticker.PrLast))
		h.Dglamount = h.Dglamount.Add(lot.Qty.Mul(ticker.PrDiffAmt))
		h.Glamount = h.MktValue.Sub(h.CostValue)
		if !h.CostValue.IsZero() {
			h.Glperc = h.Glamount.Mul(decimal.NewFromFloat(100.0)).Div(h.CostValue)
		}

		p.logger.Trace("GetHoldings", "Holding", h.Qty)

	}

	p.logger.Info("GetHoldings", "Holdings", len(hldgs))

	return hldgs, nil
}

func (p PortfolioService) GetActivities(uid string, category string, atype string,
	acctIds []string, startDate time.Time, endDate time.Time) ([]dto.ActivityResponse, error) {

	ractvs := []dto.ActivityResponse{}

	acctIdsm := make(map[string]string)
	for _, acctId := range acctIds {
		acctIdsm[acctId] = acctId
	}

	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return ractvs, nil
	}
	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	actvs, err := p.storage.GetActivities(uid)
	if err != nil {
		return nil, err
	}

	var filter bool
	for _, actv := range actvs {

		acct := acctsm[actv.AccountID]
		if acct == nil {
			p.logger.Error("GetHoldings - Account not found", "AccountId", actv.AccountID, "AcvitityId", actv.ID)
			// log.Println(lot)
			continue
		}
		p.logger.Debug("GetActivities", "Actv", actv.Debug(), "Date", actv.Date)

		filter = utils.IsDateBetween(startDate, endDate, actv.Date)
		if !filter {
			continue
		}
		filter = filterAccount(acctIdsm, acct, category, atype, acctIds)
		if !filter {
			continue
		}
		ractv := dto.NewActivityResponseFromActivity(*acct, *actv)
		ractv.Value = actv.Value
		// ractv.RcvAccount = actv.RcvAccount
		// ractv.SentAccount = actv.SentAccount
		ractv.RcvBalance = actv.RcvBalance
		ractv.SentBalance = actv.SentBalance
		if actv.TxnType == domain.ActivityTypeDividend || actv.TxnType == domain.ActivityTypeInterest {
			ractv.Notes = fmt.Sprintf("For %s", actv.SentSymbol)
		}

		acct = acctsm[actv.RcvAccountID]
		if acct != nil {
			ractv.RcvAccount = acct.Name
		}
		acct = acctsm[actv.SentAccountID]
		if acct != nil {
			ractv.SentAccount = acct.Name
		}

		ractvs = append(ractvs, ractv)
	}

	p.logger.Debug("GetActivities", "Actvs", len(ractvs))

	return ractvs, nil
}

func (p PortfolioService) GetIncome(uid string, category string, atype string,
	acctIds []string, startDate time.Time, endDate time.Time) ([]dto.Income, error) {

	acctIdsm := make(map[string]string)
	for _, acctId := range acctIds {
		acctIdsm[acctId] = acctId
	}

	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return nil, fmt.Errorf("accounts not found")
	}
	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	actvs, err := p.storage.GetActivities(uid)
	if err != nil {
		return nil, fmt.Errorf("activites error")
	}
	var filter bool
	incomes := []dto.Income{}

	for _, actv := range actvs {

		if !actv.IsIncome() {
			continue
		}

		acct := acctsm[actv.RcvAccountID]
		if acct == nil {
			p.logger.Error("GetHoldings - Account not found", "AccountId", actv.AccountID, "AcvitityId", actv.ID)
			// log.Println(lot)
			continue
		}

		filter = utils.IsDateBetween(startDate, endDate, actv.Date)
		if !filter {
			continue
		}
		filter = filterAccount(acctIdsm, acct, category, atype, acctIds)
		if !filter {
			continue
		}
		filter = filterAccount(acctIdsm, acct, category, atype, acctIds)
		if !filter {
			continue
		}

		income := dto.Income{}
		income.Category = string(acct.Category)
		income.Type = string(acct.Type)
		income.AccountName = acct.Name
		income.Date = actv.Date
		income.Symbol = actv.RcvSymbol
		income.Qty = actv.RcvAmount
		income.CostValue = actv.SentAmount
		income.Cost = income.CostValue.Div(income.Qty)
		if actv.TxnType == domain.ActivityTypeDividend || actv.TxnType == domain.ActivityTypeInterest {
			income.Symbol = actv.SentSymbol
			income.Qty = decimal.NewFromFloat(1.0)
			income.Cost = actv.RcvAmount
			income.CostValue = actv.RcvAmount
		}

		incomes = append(incomes, income)
	}
	return incomes, nil
}

func (p PortfolioService) RefreshUserAccounts(ctx context.Context, uid string, simulate bool) error {
	p.logger.Info("RefreshAccounts", "UID", uid, "Simulate", simulate)
	portfolio := portfolio.NewPortfolio(p.storage, p.logConfig, p.logger)
	return portfolio.RefreshUserAccounts(ctx, uid, simulate)
}
func (p PortfolioService) SyncUserAccounts(ctx context.Context, uid string) error {
	p.logger.Trace("RefreshAccounts", "UID", uid)
	portfolio := portfolio.NewPortfolio(p.storage, p.logConfig, p.logger)
	return portfolio.SyncUserAccounts(ctx, uid)
}
