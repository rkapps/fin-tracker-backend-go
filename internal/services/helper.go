package services

import (
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

func GetTickersMapforLots(storage storage.StorageService, lots []*domain.ActivityLot) map[string]domain.Ticker {
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
		ticker.PrLast = decimal.NewFromFloat(1.0)
		ticker.PrDiffAmt = decimal.Zero
		ticker.PrDiffPerc = decimal.Zero
	}

	return ticker

}
