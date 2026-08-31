package core

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

// SyncProvider is the sync-side contract: fetch raw payloads, nothing more.
type SyncProvider interface {
	Name() string
	Streams() []string
	ValidateConfig(config json.RawMessage) error
	FetchRaw(ctx context.Context, acred domain.AccountWithCredential, stream, cursor string) (SyncPage, error)
}

// GlobalSyncProvider is an OPTIONAL extension — only providers with
// account-independent data (Cardano epoch info, etc.) implement this.
type GlobalSyncProvider interface {
	GlobalStreams() []string
	FetchGlobalRaw(ctx context.Context, stream, cursor string) (SyncPage, error)
}

type SyncPage struct {
	Items      []domain.RawItem
	NextCursor string // opaque; only the provider understands it
	HasMore    bool
}

type SyncRegistry struct {
	providers map[string]SyncProvider
}

func NewSyncRegistry() *SyncRegistry {
	return &SyncRegistry{providers: map[string]SyncProvider{}}
}

func (r *SyncRegistry) Register(p SyncProvider) {
	r.providers[p.Name()] = p
}

func (r *SyncRegistry) Get(name string) (SyncProvider, error) {
	p, ok := r.providers[name]
	if !ok {
		return nil, fmt.Errorf("no provider registered for %q", name)
	}
	return p, nil
}
