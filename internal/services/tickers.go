package services

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type TickersService struct {
	storage storage.StorageService
}

func NewStocksService(storage storage.StorageService) TickersService {
	return TickersService{
		storage: storage,
	}
}

// DeleteTicker returns the ticker for the exchange:symbol
func (t TickersService) DeleteTicker(ctx context.Context, id string) error {
	return t.storage.DeleteTicker(id)
}

// // getTickersByFilter returns all the tickers
// func (s StocksService) getTickersByFilter(ctx context.Context, filter any, sort bson.D) (Tickers, error) {
// 	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
// 	return tr.Find(ctx, filter, sort, 0, 0)
// }

// GetTicker returns the ticker for the exchange:symbol
func (t TickersService) GetTicker(ctx context.Context, id string) (*domain.Ticker, error) {
	return t.storage.GetTicker(id)
}

func (t TickersService) GetTickerGroups(ctx context.Context) (domain.TickerGroups, error) {
	return t.storage.GetTickerGroups()
}

// GetTickerHistory returns the ticker history for the symbol
func (t TickersService) GetTickerHistory(ctx context.Context, symbol string) ([]*domain.TickerHistory, error) {
	return t.storage.GetTickerHistory(symbol)
}

// GetTickerSentiments returns the ticker sentiments for the symbol
func (t TickersService) GetTickerSentiments(ctx context.Context, symbol string) ([]*domain.TickerSentiment, error) {
	return t.storage.GetTickerSentiments(symbol)
}

// GetTickerEmbeddings returns the ticker embeddings for the symbol
func (t TickersService) GetTickerEmbeddings(ctx context.Context, symbol string) ([]*domain.TickerEmbedding, error) {
	return t.storage.GetTickerEmbeddings(symbol)
}

// GetTickers returns the tickers for the symbols
func (t TickersService) GetTickers(ctx context.Context, symbols []string) (domain.Tickers, error) {
	return t.storage.GetTickers(symbols)
}

// SearchTicker search tickers based on input fields
func (t TickersService) SearchTicker(ctx context.Context, ts domain.TickerSearch) (domain.Tickers, error) {
	return t.storage.SearchTicker(ts)
}
