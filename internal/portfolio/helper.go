package portfolio

import (
	"fmt"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

// group accounts by provider
func groupAccountsByProvider(accounts []domain.Account) map[string][]domain.Account {
	accountm := make(map[string][]domain.Account)

	//TODO
	return accountm
}

// get account and symbol key
func getAccountSymbolKey(account, currency string) string {
	return fmt.Sprintf("%s:%s", account, currency)
}

func GetHoldingsKey(byAccount bool, acct domain.Account, symbol string) string {

	var key string
	if byAccount {
		key = fmt.Sprintf("%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, acct.ID)
	} else {
		key = fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, acct.ID, symbol)
	}

	return key
}

// defaultLotMatchingMethod returns the default method for an account category.
func defaultLotMatchingMethod(category domain.AccountCategory) domain.LotMatchingMethod {
	switch category {
	case domain.CategoryCrypto:
		return domain.LotMatchingHIFO
	default:
		return domain.LotMatchingFIFO
	}
}

func filterAccount(acctIdsm map[string]string, acct *domain.Account) bool {

	if len(acctIdsm) > 0 {
		if _, ok := acctIdsm[acct.ID]; !ok {
			return false
		}
	}

	return true
}

func filterBankAccount(acctIdsm map[string]*domain.Account, acctId string) bool {

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
	accts []*domain.Account, acctIds []string, lots []*domain.ActivityLot) ([]*domain.HoldingSummary, error) {

	hldgs := []*domain.HoldingSummary{}
	hldgsm := make(map[string]*domain.HoldingSummary)

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

		filter := filterBankAccount(acctsm, lot.AccountID)
		if !filter {
			continue
		}

		filter = filterAccount(acctIdsm, acct)
		if !filter {
			continue
		}

		if byAccount {
			key = fmt.Sprintf("%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID)
		} else {
			key = fmt.Sprintf("%s-%s-%s-%s-%s", acct.Category, acct.Type, acct.Name, lot.AccountID, lot.Symbol)
		}

		logger.Debug("GetHoldings", "Key", key, "Lot", lot.Qty)

		ticker := tm[lot.Symbol]
		if len(ticker.Symbol) == 0 {
			ticker = GetTickerPriceDiff(tm, lot.Symbol)
			tm[lot.Symbol] = ticker
		}

		h := hldgsm[key]

		zero := decimal.NewFromFloat(0.0)
		if h == nil {
			h = &domain.HoldingSummary{}
			h.Category = string(acct.Category)
			h.Type = string(acct.Type)
			h.AccountName = acct.Name
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
		h.Qty = h.Qty.Add(lot.Qty)
		h.CostValue = h.CostValue.Add(lot.CostValue)
		if !h.Qty.IsZero() {
			h.Cost = h.CostValue.Div(h.Qty)
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

		logger.Trace("GetHoldings", "Holding", h.Qty)

	}

	logger.Info("GetHoldings", "Holdings", len(hldgs))

	return hldgs, nil

}

func GetTickersMapforLots(storage storage.TickerStorageService, lots []*domain.ActivityLot) map[string]domain.Ticker {
	tm := make(map[string]domain.Ticker)

	tsymbols := []string{}
	tsymbolsm := make(map[string]string)
	for _, lot := range lots {

		symbol := lot.Symbol
		if strings.Compare(symbol, "ETH2") == 0 ||
			strings.Compare(symbol, "WETH") == 0 {
			symbol = "ETH"
		} else if strings.Compare(symbol, "mSOL") == 0 {
			symbol = "SOL"
		}
		if _, ok := tsymbolsm[symbol]; ok {
			continue
		}
		tsymbols = append(tsymbols, symbol)
		tsymbolsm[symbol] = symbol
	}

	ts, _ := storage.GetTickers(tsymbols)
	for _, ticker := range ts {
		tm[ticker.Symbol] = *ticker
	}

	return tm
}

func GetTickerPriceDiff(tm map[string]domain.Ticker, symbol string) domain.Ticker {

	var ticker domain.Ticker
	if strings.Compare(symbol, "ETH2") == 0 ||
		strings.Compare(symbol, "WETH") == 0 {
		ticker = tm["ETH"]
	} else if strings.Compare(symbol, "mSOL") == 0 {
		ticker = tm["SOL"]
	} else {
		ticker = tm[symbol]
	}
	if len(ticker.Symbol) == 0 {
		ticker = domain.Ticker{}
		ticker.Symbol = symbol
		ticker.PrLast = decimal.Zero
		ticker.PrDiffAmt = decimal.Zero
		ticker.PrDiffPerc = decimal.Zero
	}
	if strings.Compare(symbol, "USD") == 0 {
		ticker.AssetType = "Cash"
		ticker.PrLast = decimal.NewFromFloat(1.0)
		ticker.PrDiffAmt = decimal.Zero
		ticker.PrDiffPerc = decimal.Zero
	}

	return ticker

}
