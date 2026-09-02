package kraken

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
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
	slog := logConfig.For("syncer.kraken")
	return &Provider{HTTP: api, Limiter: rate.NewLimiter(rate.Every(6*time.Second), 1), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	return "kraken"
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	// return []string{"payments", "accounts", "fills", "orders"}
	return []string{"trades", "ledgers"}
}

// GlobalStreams implements [provider.SyncSourceProvider].
func (p *Provider) GlobalStreams() []string {
	// return []string{"payments", "accounts", "fills", "orders"}
	return []string{"assets", "assetpairs"}
}

// ValidateConfig implements [provider.SyncSourceProvider].
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	if cfg.Api_Key == "" || cfg.Api_Secret == "" {
		return fmt.Errorf("coinbase_pro: key_name and private_key required")
	}
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

	var cfg Config
	json.Unmarshal(acred.Config, &cfg)

	switch stream {
	case "trades":
		return p.fetchTrades(ctx, acred, cfg, stream, cursor)
	case "ledgers":
		return p.fetchLedgers(ctx, acred, cfg, stream, cursor)
	}
	return core.SyncPage{}, nil
}

// FetchRaw implements [provider.SyncSourceProvider].
func (p *Provider) FetchGlobalRaw(ctx context.Context,
	stream string, cursor string,
) (core.SyncPage, error) {

	switch stream {
	case "assets":
		return p.fetchAssets(ctx, stream)
	case "assetpairs":
		return p.fetchAssetPairs(ctx, stream)
	}
	return core.SyncPage{}, nil
}

func (p *Provider) fetchAssets(ctx context.Context, stream string) (core.SyncPage, error) {
	var err error

	resp, err := p.HTTP.GetAssets(ctx)
	if err != nil {
		return core.SyncPage{}, err
	}
	p.logger.Debug("fetchAssets", "assets", len(resp))
	var items []domain.RawItem
	raw, _ := json.Marshal(resp)
	items = append(items, core.GetGlobalRawItem(p.Name(), stream, "all", raw, time.Now()))

	return core.SyncPage{Items: items, HasMore: false}, nil

}

func (p *Provider) fetchAssetPairs(ctx context.Context, stream string) (core.SyncPage, error) {
	var err error

	resp, err := p.HTTP.GetAssetPairs(ctx)
	if err != nil {
		return core.SyncPage{}, err
	}
	var items []domain.RawItem

	p.logger.Debug("fetchAssetPairs", "pairs", len(resp))
	raw, _ := json.Marshal(resp)
	items = append(items, core.GetGlobalRawItem(p.Name(), stream, "all", raw, time.Now()))

	return core.SyncPage{Items: items, HasMore: false}, nil

}

func (p *Provider) fetchTrades(ctx context.Context, acred domain.AccountWithCredential, cfg Config, stream string, cursor string) (core.SyncPage, error) {
	var cur krTradesCursor
	if cursor != "" {
		json.Unmarshal([]byte(cursor), &cur)
	}

	var resp *TradesHistoryResponse
	var err error
	if cur.Start == "" {
		resp, err = p.HTTP.GetTradeHistory(ctx, cfg, "", strconv.Itoa(cur.Offset))
	} else {
		resp, err = p.HTTP.GetTradeHistory(ctx, cfg, cur.Start, "")
	}
	if err != nil {
		return core.SyncPage{}, err
	}

	var items []domain.RawItem
	for id, t := range resp.Trades {
		t.TradeId = id
		raw, _ := json.Marshal(t)
		ts, _ := crypto.KrakenTime(t.Time)
		items = append(items, core.GetRawItem(acred.Account, p.Name(), stream, id, raw, *ts))
		if t.Time > cur.MaxFTime {
			cur.MaxFTime = t.Time
		}
	}

	hasMore := len(resp.Trades) >= 50

	if cur.Start == "" {
		cur.Offset += 50
	}
	if cur.MaxFTime > 0 {
		cur.Start = fmt.Sprintf("%f", cur.MaxFTime)
	}

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: hasMore}, nil
}

func (p *Provider) fetchLedgers(ctx context.Context, acred domain.AccountWithCredential, cfg Config, stream string, cursor string) (core.SyncPage, error) {
	var cur krLedgerCursor
	if cursor != "" {
		json.Unmarshal([]byte(cursor), &cur)
	}

	var resp *LedgersResponse
	var err error
	if cur.Start == "" {
		resp, err = p.HTTP.GetLedgers(ctx, cfg, "", strconv.Itoa(cur.Offset))
	} else {
		resp, err = p.HTTP.GetLedgers(ctx, cfg, cur.Start, "0")
	}
	if err != nil {
		return core.SyncPage{}, err
	}

	p.logger.Debug("fetchLedger", "Ledgers", len(resp.Ledger))
	var items []domain.RawItem
	for id, v := range resp.Ledger {
		// add the infoid
		v.InfoId = id
		raw, _ := json.Marshal(v)
		t, _ := crypto.KrakenTime(v.Time)
		items = append(items, core.GetRawItem(acred.Account, p.Name(), stream, id, raw, *t))
		if v.Time > cur.MaxFTime {
			cur.MaxFTime = v.Time
		}
	}

	hasMore := len(resp.Ledger) >= 50

	if cur.Start == "" {
		cur.Offset += 50
	}
	if !hasMore && cur.MaxFTime > 0 {
		cur.Start = fmt.Sprintf("%f", cur.MaxFTime) // only set once backfill is truly done
	}

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: hasMore}, nil
}
