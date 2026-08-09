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
	storage        storage.FinTrackerStorageService
	logConfig      *logger.Config
	logger         *logger.Logger
}

func NewPortfolioService(logConfig *logger.Config, tickersService TickersService, storage storage.FinTrackerStorageService) PortfolioService {
	plog := logConfig.For("portfolio.service")
	return PortfolioService{tickersService: tickersService, storage: storage, logConfig: logConfig, logger: plog}
}

func (p PortfolioService) GetSummary(uid string) ([]*domain.AccountSummary, error) {
	return p.storage.GetAccountSummaries(uid)
}

func (p PortfolioService) GetHoldings(uid string, category string, atype string, acctIds []string) ([]*dto.HoldingSummary, error) {

	hldgs := []*dto.HoldingSummary{}
	var err error
	accts, err := p.storage.GetAccounts(uid)
	if err != nil {
		return hldgs, nil
	}

	lots, err := p.storage.GetActivityLots(uid)
	if err != nil {
		return hldgs, nil
	}
	p.logger.Info("GetHoldings", "lots", len(lots), "ticker storage", p.tickersService.storage)
	return portfolio.GetHoldings(p.tickersService.storage, p.logger, false, accts, acctIds, lots)

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
			ractv.RcvAccount = acct.ID
		}
		acct = acctsm[actv.SentAccountID]
		if acct != nil {
			ractv.SentAccount = acct.ID
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
		income.AcctountID = acct.ID
		income.AccountName = acct.Name
		income.AccountDisplayName = acct.ID

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
	portfolio := portfolio.NewPortfolio(p.storage, p.tickersService.storage, p.logConfig, p.logger)
	return portfolio.RefreshUserAccounts(ctx, uid, simulate)
}
func (p PortfolioService) SyncUserAccounts(ctx context.Context, uid string) error {
	p.logger.Trace("RefreshAccounts", "UID", uid)
	portfolio := portfolio.NewPortfolio(p.storage, p.tickersService.storage, p.logConfig, p.logger)
	return portfolio.SyncUserAccounts(ctx, uid)
}
