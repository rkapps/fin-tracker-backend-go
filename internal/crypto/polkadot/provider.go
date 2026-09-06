package polkadot

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"golang.org/x/time/rate"
)

type Provider struct {
	HTTP    API
	Limiter *rate.Limiter
	logger  *logger.Logger
}

// compile-time check this satisfies the interface — also self-documenting
var _ core.SyncProvider = (*Provider)(nil)
var _ core.GlobalSyncProvider = (*Provider)(nil)

func New(api API, logConfig *logger.Config) *Provider {
	slog := logConfig.For("syncer.polkadot")
	return &Provider{HTTP: api, Limiter: rate.NewLimiter(rate.Limit(10), 10), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	return "polkadot"
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	return []string{"rewards", "transfers"}
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) GlobalStreams() []string {
	return nil
}

// ValidateConfig implements [provider.SyncSourceProvider].
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	return nil
}

// FetchRaw implements [provider.SyncSourceProvider].
func (p *Provider) FetchRaw(ctx context.Context,
	acred domain.AccountWithCredential,
	stream string, cursor string,
) (core.SyncPage, error) {

	if err := p.Limiter.Wait(ctx); err != nil {
		return core.SyncPage{}, err
	}

	switch stream {
	case "rewards":
		return p.fetchRewards(acred.Account, stream, cursor)
	case "transfers":
		// return p.fetchTransfers(acred.Account, stream, cursor)
	}
	return core.SyncPage{}, nil
}

// FetchRaw implements [provider.SyncSourceProvider].
func (p *Provider) FetchGlobalRaw(ctx context.Context,
	stream string, cursor string,
) (core.SyncPage, error) {
	return core.SyncPage{}, nil

}

func (p Provider) fetchRewards(account domain.Account, stream string, cursor string) (core.SyncPage, error) {

	rCount := 10
	address := account.Address()
	var cur polkadotRewardsCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.Page = 0
	}
	p.logger.Debug("fetchRewards", "Address", address, "Cursor", fmt.Sprintf("%s-%v", cursor, cur))

	rewardData, err := p.HTTP.GetRewards(address, cur.Page, rCount)
	if err != nil {
		log.Println(err)
		p.logger.Error("fetchRewards", "GetAccountRewards", err)
		// log.Println(err)
		return core.SyncPage{}, err
	}
	p.logger.Info("fetchRewards", "Rewards", len(rewardData.Data.List))

	var items []domain.RawItem
	for _, r := range rewardData.Data.List {
		raw, _ := json.Marshal(r)
		item := core.GetRawItem(account, p.Name(), stream, r.Event_Index, raw, time.Now())
		items = append(items, item)
	}

	hasMore := len(rewardData.Data.List) >= rCount
	if hasMore {
		cur.Page++
	}
	p.logger.Debug("fetchRewards", "Items", len(items), "hasMore", hasMore)

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: hasMore}, nil
}

func (p Provider) fetchTransfers(account domain.Account, stream string, cursor string) (core.SyncPage, error) {

	rCount := 10
	address := account.Address()
	var cur polkadotTransfersCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.Page = 0
	}
	p.logger.Debug("fetcTransfers", "Address", address, "Cursor", fmt.Sprintf("%s-%v", cursor, cur))

	tData, err := p.HTTP.GetTransfers(address, cur.Page, rCount)
	if err != nil {
		p.logger.Error("fetchTransfers", "Transfer", err)
		log.Println(err)
		return core.SyncPage{}, err
	}
	p.logger.Info("fetchTransfers", "Transfers", len(tData.Data.Transfers))

	var items []domain.RawItem
	for _, r := range tData.Data.Transfers {
		raw, _ := json.Marshal(r)
		item := core.GetRawItem(account, p.Name(), stream, r.Hash, raw, time.Now())
		items = append(items, item)
	}

	hasMore := len(tData.Data.Transfers) >= rCount
	if hasMore {
		cur.Page++
	}

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: false}, nil
}
