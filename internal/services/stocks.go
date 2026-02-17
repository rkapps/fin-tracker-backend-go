package services

import (
	"context"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

var (
	TIINGO_API_TOKEN = os.Getenv("TIINGO_API_TOKEN")
	ALPHA_API_KEY    = os.Getenv("ALPHA_KEY")
)

type StocksService struct {
	storage storage.StorageService
}

func NewStocksService(storage storage.StorageService) StocksService {
	return StocksService{
		storage: storage,
	}
}

// DeleteTicker returns the ticker for the exchange:symbol
func (s StocksService) DeleteTicker(ctx context.Context, id string) error {
	return s.storage.DeleteTicker(id)
}

// // getTickersByFilter returns all the tickers
// func (s StocksService) getTickersByFilter(ctx context.Context, filter any, sort bson.D) (Tickers, error) {
// 	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
// 	return tr.Find(ctx, filter, sort, 0, 0)
// }

// GetTicker returns the ticker for the exchange:symbol
func (s StocksService) GetTicker(ctx context.Context, id string) (*domain.Ticker, error) {
	return s.storage.GetTicker(id)
}

func (s StocksService) GetTickerGroups(ctx context.Context) (domain.TickerGroups, error) {
	return s.storage.GetTickerGroups()
}

// GetTickerHistory returns the ticker history for the symbol
func (s StocksService) GetTickerHistory(ctx context.Context, symbol string) ([]*domain.TickerHistory, error) {
	return s.storage.GetTickerHistory(symbol)
}

// GetTickers returns the tickers for the symbols
func (s StocksService) GetTickers(ctx context.Context, symbols []string) (domain.Tickers, error) {
	return s.storage.GetTickers(symbols)
}

// LoadTickers returns the tickers for the symbols
func (s StocksService) LoadTickers(ctx context.Context, ts domain.Tickers) error {
	zero := decimal.NewFromFloat(0.0)
	for _, t := range ts {
		t.Pr52WkHigh = zero
		t.Pr52WkLow = zero
		t.PrClose = zero
		t.PrDiffAmt = zero
		t.PrDiffPerc = zero
		t.PrHigh = zero
		t.PrLow = zero
		t.PrOpen = zero
		t.PrPrev = zero
		t.Performance = make(map[string]map[string]decimal.Decimal)
		t.Technicals = make(map[string]map[string]decimal.Decimal)
		t.PerformanceSearch = make(map[string]map[string]float64)
		t.TechnicalsSearch = make(map[string]map[string]float64)
		s.storage.SaveTicker(t)
		break
	}
	// return s.updateTickersEOD(ctx, ts, true)
	return nil
}

// SearchTicker search tickers based on input fields
func (s StocksService) SearchTicker(ctx context.Context, ts domain.TickerSearch) (domain.Tickers, error) {
	return s.storage.SearchTicker(ts)
}

// SaveTicker saves ticker
func (s StocksService) SaveTicker(ctx context.Context, t *domain.Ticker) error {
	return s.storage.SaveTicker(t)
}
