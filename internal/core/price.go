package core

import (
	"fmt"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type PriceService struct {
	tstorage storage.TickerStorageService
	pstorage storage.ProviderStorageService
	pm       map[string]*domain.CryptoPrice
}

func NewPriceService(tstorage storage.TickerStorageService, pstorage storage.ProviderStorageService) PriceService {
	pm := make(map[string]*domain.CryptoPrice)
	return PriceService{tstorage, pstorage, pm}
}

func (ps PriceService) GetCryptoPrice(symbol string, date time.Time) (decimal.Decimal, error) {

	//Truncate date to the nearest hour
	d := 24 * 60 * time.Minute
	ndate := date.Truncate(d)
	ms := ndate.UTC().UnixMilli()

	if len(ps.pm) == 0 {
		ps.LoadCryptoPrices()
	}

	key := fmt.Sprintf("%s-%d", symbol, ms)
	// log.Printf("Date: %v Key: %s", ndate, key)
	cp, ok := ps.pm[key]

	if !ok {
		th, err := ps.tstorage.GetTickerHistoryByDate(symbol, ndate)
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
	prices, _ := ps.pstorage.GetCryptoPrices()
	for _, price := range prices {
		ps.pm[price.ID] = price
	}

}
