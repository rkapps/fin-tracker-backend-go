package core

import (
	"encoding/json"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/dto"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

func GetHoldingsKey(byAccount bool, acct domain.Account, symbol string) string {

	var key string
	if byAccount {
		key = fmt.Sprintf("%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, acct.ID)
	} else {
		key = fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, acct.ID, symbol)
	}

	return key
}

func FilterAccount(acctIdsm map[string]string, acct *domain.Account) bool {

	if len(acctIdsm) > 0 {
		if _, ok := acctIdsm[acct.ID]; !ok {
			return false
		}
	}

	return true
}

func FilterBankAccount(acctIdsm map[string]*domain.Account, acctId string) bool {

	acct, ok := acctIdsm[acctId]
	if !ok {
		return false
	}

	if acct.Category == domain.CategoryCash {
		return false
	}

	return true
}

func GetHoldings(storage storage.TickerStorageService, logger *logger.Logger, byAccount bool,
	accts []*domain.Account, acctIds []string, lots []*domain.ActivityLot) ([]*dto.HoldingResponse, error) {

	hldgs := []*dto.HoldingResponse{}
	hldgsm := make(map[string]*dto.HoldingResponse)

	acctIdsm := make(map[string]string)
	for _, acctId := range acctIds {
		acctIdsm[acctId] = acctId
	}

	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	// get tickermap
	tm := GetTickersMapforLots(storage, lots)
	var key string
	for _, lot := range lots {
		if lot.Status != domain.LotStatusOpen {
			continue
		}
		acct := acctsm[lot.AccountID]
		if acct == nil {
			logger.Error("GetHoldings - Account not found", "AccountId", lot.AccountID, "LotId", lot.ID)
			// log.Println(lot)
			continue
		}

		filter := FilterBankAccount(acctsm, lot.AccountID)
		if !filter {
			continue
		}

		filter = FilterAccount(acctIdsm, acct)
		if !filter {
			continue
		}

		if byAccount {
			key = fmt.Sprintf("%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID)
		} else {
			key = fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID, lot.Symbol)
		}

		logger.Debug("GetHoldings", "Key", key, "Lot", lot.Amount)

		ticker := tm[lot.Symbol]
		if len(ticker.Symbol) == 0 {
			ticker = GetTickerPriceDiff(tm, lot.Symbol)
			tm[lot.Symbol] = ticker
		}

		h := hldgsm[key]

		zero := decimal.NewFromFloat(0.0)
		if h == nil {
			h = &dto.HoldingResponse{}
			h.Category = string(acct.Category)
			h.Type = string(acct.Type)
			h.AccountName = acct.Name
			h.Blockchain = acct.Blockchain()
			// h.AccountDisplayName = acct.ID
			h.AcctountID = lot.AccountID
			h.Symbol = lot.Symbol
			h.AssetType = ticker.AssetType
			h.Sector = ticker.Sector
			h.Industry = ticker.Industry
			h.Qty = zero
			h.Cost = zero
			h.CostValue = zero
			h.MktValue = zero
			hldgs = append(hldgs, h)
			hldgsm[key] = h
		}
		h.Cost = lot.Cost
		h.Qty = h.Qty.Add(lot.Amount)
		h.CostValue = h.CostValue.Add(lot.CostValue)
		if !h.Qty.IsZero() {
			h.Cost = h.CostValue.Div(h.Qty)
		}

		h.PrLast = ticker.PrLast
		h.PrDiffAmt = ticker.PrDiffAmt
		h.PrDiffPerc = ticker.PrDiffPerc
		h.MktValue = h.MktValue.Add(lot.Amount.Mul(ticker.PrLast))
		h.Dglamount = h.Dglamount.Add(lot.Amount.Mul(ticker.PrDiffAmt))
		h.Glamount = h.MktValue.Sub(h.CostValue)
		if !h.CostValue.IsZero() {
			h.Glperc = h.Glamount.Mul(decimal.NewFromFloat(100.0)).Div(h.CostValue)
		}

		logger.Trace("GetHoldings", "Holding", h.Qty)
	}

	uhldgs := []*dto.HoldingResponse{}
	for _, hldg := range hldgs {
		// if hldg.MktValue.LessThan(decimal.NewFromFloat(1.0)) {
		// 	continue
		// }
		uhldgs = append(uhldgs, hldg)
	}

	logger.Info("GetHoldings", "Holdings", len(uhldgs))
	return uhldgs, nil

}

// domain/account_detail.go — extract the dispatch so both Account.UnmarshalJSON
// and handler code can use it without duplicating the switch
func NewAccountDetail(category domain.AccountCategory, raw json.RawMessage) (domain.AccountDetail, error) {
	var detail domain.AccountDetail
	switch category {
	case domain.CategoryBrokerage, domain.CategoryRetirement, domain.CategoryHSA:
		detail = &domain.BrokerageDetail{}
	case domain.CategoryCrypto:
		detail = &domain.CryptoDetail{}
	case domain.CategoryCash:
		detail = &domain.BankDetail{}
	case domain.Category529:
		detail = &domain.EducationDetail{}
	default:
		return nil, fmt.Errorf("unknown account category: %s", category)
	}
	if err := json.Unmarshal(raw, detail); err != nil {
		return nil, err
	}
	return detail, nil
}
