package solana

import (
	"context"
	"encoding/json"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
	"golang.org/x/time/rate"
)

type Provider struct {
	HTTP    API
	Limiter *rate.Limiter
	logger  *logger.Logger
}

// compile-time check this satisfies the interface — also self-documenting
var _ core.SyncProvider = (*Provider)(nil)

func New(api API, logConfig *logger.Config) *Provider {
	slog := logConfig.For("syncer.solana")
	return &Provider{HTTP: api, Limiter: rate.NewLimiter(rate.Every(6*time.Second), 1), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	return "solana"
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	// return []string{"payments", "accounts", "fills", "orders"}
	return []string{"tokens", "transactions"}
}

// GlobalStreams implements [provider.SyncSourceProvider].
func (p *Provider) GlobalStreams() []string {
	return []string{}
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
	case "tokens":
		return p.fetchTokens(acred.Account, stream)
	case "transactions":
		return p.fetchTransactions(acred.Account, cursor)
	}
	return core.SyncPage{}, nil
}

// FetchRaw implements [provider.SyncSourceProvider].
func (p *Provider) FetchGlobalRaw(ctx context.Context,
	stream string, cursor string,
) (core.SyncPage, error) {
	return core.SyncPage{}, nil
}

func (p *Provider) fetchTokens(account domain.Account, stream string) (core.SyncPage, error) {

	resp, err := p.HTTP.GetSolanaTokenAccounts(account.Address())
	if err != nil {
		return core.SyncPage{}, err
	}

	var items []domain.RawItem
	for _, t := range resp.Result.Value {
		raw, _ := json.Marshal(t)
		items = append(items, providers.GetRawItem(account, p.Name(), stream, t.Pubkey, raw, time.Now()))
	}

	return core.SyncPage{Items: items, NextCursor: "", HasMore: false}, nil
}

func (p *Provider) fetchTransactions(account domain.Account, cursor string) (core.SyncPage, error) {

	var cur SolanaTransactionCursor
	if len(cursor) == 0 {
		cur = SolanaTransactionCursor{UntilSig: ""}
	} else {
		json.Unmarshal([]byte(cursor), &cur)
	}

	resp, err := p.HTTP.GetSolanaSignaturesForAddress(account.Address(), cur.UntilSig)
	if err != nil {
		return core.SyncPage{}, err
	}

	p.logger.Debug("fetchTransactions", "Signatures", len(resp.Result))
	var items []domain.RawItem

	for _, v := range resp.Result {

		date, _ := providers.ConvertInt64ToTime(int64(v.BlockTime))
		id := v.Signature
		v.Address = account.Address()
		v.Date = date
		raw, _ := json.Marshal(v)
		items = append(items, providers.GetRawItem(account, p.Name(), "signatures", id, raw, *date))

		// Get transaction for the signature
		result, err := p.HTTP.GetSolanaTransaction(v.Signature)
		if err != nil {
			p.logger.Error("fetchTransactions", "Signature", v.Signature, "Error", err)
			continue
		}
		txn := result.Result
		txn.UID = account.UID
		txn.Acct_Id = account.ID
		txn.Signature = v.Signature
		txn.Address = v.Address
		txn.Date = v.Date

		raw, _ = json.Marshal(txn)
		items = append(items, providers.GetRawItem(account, p.Name(), "transactions", id, raw, *date))
	}

	if len(resp.Result) > 0 {
		cur.UntilSig = resp.Result[0].Signature
	}

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: false}, nil
}
