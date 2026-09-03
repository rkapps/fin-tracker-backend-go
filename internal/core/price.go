package core

import (
	"fmt"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type PriceService struct {
	tstorage storage.TickerStorageService
	cstorage storage.CryptoStorageService
	pm       map[string]*domain.CryptoPrice
}

func NewPriceService(tstorage storage.TickerStorageService, cstorage storage.CryptoStorageService) PriceService {
	pm := make(map[string]*domain.CryptoPrice)
	return PriceService{tstorage, cstorage, pm}
}

func (ps PriceService) GetCryptoPrice(symbol string, date time.Time) (decimal.Decimal, error) {

	if strings.Compare(symbol, "USDT") == 0 ||
		strings.Compare(symbol, "USDC") == 0 ||
		strings.Compare(symbol, "DAI") == 0 {
		return decimal.NewFromFloat(1.0), nil
	}

	psymbol := symbol
	if strings.Compare(symbol, "WETH") == 0 || strings.Compare(symbol, "ETH2") == 0 {
		psymbol = "ETH"
	}
	if strings.Compare(symbol, "WPOL") == 0 {
		psymbol = "POL"
	}
	if strings.Compare(symbol, "mSOL") == 0 {
		psymbol = "SOL"
	}

	//Truncate date to the nearest hour
	d := 24 * 60 * time.Minute
	ndate := date.Truncate(d)
	ms := ndate.UTC().UnixMilli()

	key := fmt.Sprintf("%s-%d", symbol, ms)
	// log.Printf("Date: %v Key: %s", ndate, key)
	cp, ok := ps.pm[key]

	if !ok {
		th, err := ps.tstorage.GetTickerHistoryByDate(psymbol, ndate)
		if err != nil || th == nil {
			return decimal.Decimal{}, err
		}
		cp = &domain.CryptoPrice{}
		cp.Symbol = symbol
		cp.Ms = ms
		cp.Price = th.Close
		cp.SetId()
		ps.pm[key] = cp
	}

	return cp.Price, nil
}

func (ps PriceService) GetCryptoPrices() []*domain.CryptoPrice {
	prices := []*domain.CryptoPrice{}
	for _, price := range ps.pm {
		prices = append(prices, price)
	}
	return prices
}

func (ps PriceService) LoadCryptoPrices() {
	prices, _ := ps.cstorage.GetCryptoPrices()
	for _, price := range prices {
		ps.pm[price.ID] = price
	}

}
