package core

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type TransformerProvider interface {
	Name() string
	// Transform sees ALL unprocessed raws for the account, across streams —
	// required for cross-stream merging (Coinbase) and leg grouping (Kraken).
	Transform(ctx context.Context, ps PriceService,
		gaccts []*domain.Account,
		acreds []domain.AccountWithCredential,
		globalRaws []domain.RawItem,
		raws map[string][]domain.RawItem,
	) ([]*domain.Activity, error)
}

type TransformerRegistry struct {
	providers map[string]TransformerProvider
}

func NewTransformerRegistry() *TransformerRegistry {
	return &TransformerRegistry{providers: map[string]TransformerProvider{}}
}

func (r *TransformerRegistry) Register(p TransformerProvider) {
	r.providers[p.Name()] = p
}

func (r *TransformerRegistry) Get(name string) (TransformerProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("no provider registered for %q", name)
	}
	return p, nil
}
