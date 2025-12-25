package stocks

import (
	"context"
)

type Service interface {
	DeleteTicker(ctx context.Context, id string) error
	// GetAllTickers(ctx context.Context) (Tickers, error)
	GetTicker(ctx context.Context, id string) (*Ticker, error)
	GetTickerHistory(ctx context.Context, id string) ([]*TickerHistory, error)
	GetTickers(ctx context.Context, symbols []string) (Tickers, error)
	LoadTickers(ctx context.Context, ts Tickers) error
	SearchTicker(ctx context.Context, ts TickerSearch) (Tickers, error)
	UpdateEOD(ctx context.Context) error
	UpdateRealtime(ctx context.Context) error
}
