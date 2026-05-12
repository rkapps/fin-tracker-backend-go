package services

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/dto"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type PortfolioService struct {
	tickersService TickersService
	storage        storage.StorageService
	logConfig      *logger.Config
	logger         *logger.Logger
}

func NewPortfolioService(logConfig *logger.Config, tickersService TickersService, storage storage.StorageService) PortfolioService {
	plog := logConfig.For("portfolio")
	return PortfolioService{storage: storage, logConfig: logConfig, logger: plog}
}

func (p PortfolioService) GetHoldings(uid string) ([]*dto.HoldingSummary, error) {

	hldgs := []*dto.HoldingSummary{}
	hldgsm := make(map[string]*dto.HoldingSummary)
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
		key := fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID, lot.Symbol)
		p.logger.Debug("GetHoldings", "Key", key, "Lot", lot.Qty)

		h := hldgsm[key]

		zero := decimal.NewFromFloat(0.0)
		if h == nil {
			h = &dto.HoldingSummary{}
			h.Group = string(acct.Category)
			h.Category = string(acct.Type)
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

		p.logger.Debug("GetHoldings", "Holding", h.Qty)

	}

	p.logger.Debug("GetHoldings", "Holdings", hldgs)

	return hldgs, nil
}

func (p PortfolioService) GetActivities(uid string) ([]dto.ActivityResponse, error) {

	ractvs := []dto.ActivityResponse{}

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
	for _, actv := range actvs {

		acct := acctsm[actv.AccountID]
		if acct == nil {
			p.logger.Error("GetHoldings - Account not found", "AccountId", actv.AccountID, "AcvitityId", actv.ID)
			// log.Println(lot)
			continue
		}
		ractv := dto.NewActivityResponseFromActivity(*acct, *actv)
		ractv.Value = actv.Value
		ractv.RcvBalance = actv.RcvBalance
		ractv.SentBalance = actv.SentBalance
		ractvs = append(ractvs, ractv)
	}
	return ractvs, nil
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
