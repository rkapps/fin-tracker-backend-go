package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
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
	slog := logConfig.For("syncer.coinbase")
	return &Provider{HTTP: api, Limiter: rate.NewLimiter(rate.Limit(10), 10), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	return "coinbase"
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	// return []string{"payments", "accounts", "fills"}
	// return []string{"payments", "accounts"}
	return []string{"accounts"}
}

// ValidateConfig implements [provider.SyncSourceProvider].
func (p *Provider) ValidateConfig(config json.RawMessage) error {
	var cfg Config
	if err := json.Unmarshal(config, &cfg); err != nil {
		return err
	}
	if cfg.KeyName == "" || cfg.PrivateKey == "" {
		return fmt.Errorf("coinbase: key_name and private_key required")
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
	case "fills":
		// slog.Info("FetchRaw", "AccountsNextUrl", "reached")
		err := p.HTTP.ListFills(ctx, cfg, cursor)
		if err != nil {
			slog.Error("FetchRaw", "stream", err)
		}
	case "payments":
		return p.FetchPayments(ctx, acred.Account, cfg, stream, cursor)
	case "accounts":
		return p.FetchAccounts(ctx, acred.Account, cfg, stream, cursor)
	}

	return core.SyncPage{}, nil
}

// FetchAccounts
func (p *Provider) FetchPayments(ctx context.Context,
	account domain.Account,
	cfg Config,
	stream string, cursor string,
) (core.SyncPage, error) {

	items := []domain.RawItem{}
	hasMore := false
	data, err := p.HTTP.ListPaymentMethods(ctx, cfg, cursor)
	for _, payment_method := range data.Payment_methods {
		date := utils.DateTimeFromString(payment_method.Created_At)
		bytes, err := json.Marshal(payment_method)
		if err != nil {
		}
		item := core.GetRawItem(account, p.Name(), stream, payment_method.Id, json.RawMessage(bytes), *date)
		items = append(items, item)
	}
	return core.SyncPage{
		Items:   items,
		HasMore: hasMore,
	}, err

}

// FetchAccounts
func (p *Provider) FetchAccounts(ctx context.Context,
	account domain.Account,
	cfg Config, stream string, cursor string,
) (core.SyncPage, error) {

	var aCursor AccountsCursor
	if len(cursor) == 0 {
		aCursor = AccountsCursor{
			Txns_next_url: make(map[string]string),
		}
	} else {
		json.Unmarshal([]byte(cursor), &aCursor)
	}

	aitems := []domain.RawItem{}
	items := p.FetchAllAccounts(ctx, account, cfg, stream, "")
	p.logger.Info("FetchAccounts", "Accounts", len(items))
	aitems = append(aitems, items...)

	for _, item := range items {

		nexturi := aCursor.Txns_next_url[item.ExternalID]
		// if strings.Compare(item.ExternalID, "a7c4e4db-c50f-5da1-90d9-09ee39790557") == 0 {
		// 	nexturi = ""
		// }
		titems, nexturi := p.FetchAccountTransactions(ctx, account, cfg, stream, item.ExternalID, nexturi)
		// if i%10 == 0 {
		// }
		if strings.Compare(item.ExternalID, "a7c4e4db-c50f-5da1-90d9-09ee39790557") == 0 {
			// for _, item := range titems {
			// 	p.logger.Info("FetchAccounts", "Date", item.Timestamp, "ExternalID", item.ExternalID)
			// }
			// p.logger.Info("FetchAccounts", "External ID", item.ExternalID, "Count", fmt.Sprintf("%d of %d", i, len(items)))
		}

		aCursor.Txns_next_url[item.ExternalID] = nexturi
		aitems = append(aitems, titems...)
	}

	nCursor, err := json.Marshal(aCursor)
	if err != nil {
		slog.Error("FetchRaw", "Marshal nCursor", err)
	}
	p.logger.Info("FetchRaw", "Items", len(aitems))

	return core.SyncPage{
		Items:      aitems,
		NextCursor: string(nCursor),
		HasMore:    false,
	}, err

}

// FetchAccounts
func (p *Provider) FetchAllAccounts(ctx context.Context,
	account domain.Account,
	cfg Config, stream string,
	nextUri string,
) []domain.RawItem {

	aitems := []domain.RawItem{}
	cNextUri := nextUri

	for {

		aPageData, err := p.HTTP.ListAccounts(ctx, cfg, cNextUri, 300)
		if err != nil {
			break
		}

		for _, raw := range aPageData.Data {

			var peek CnbPeek
			if err := json.Unmarshal(raw, &peek); err != nil {
				slog.Error("FetchRaw", "UnMarshallError", err)
				continue
			}
			if !peek.CreatedAt.IsZero() && peek.CreatedAt.Equal(peek.UpdatedAt) {
				// continue
			}

			// add raw account item
			item := core.GetRawItem(account, p.Name(), stream, peek.ID, raw, peek.CreatedAt)
			aitems = append(aitems, item)

		}
		if len(aPageData.Pagination.Next_uri) == 0 {
			break
		} else {
			cNextUri = aPageData.Pagination.Next_uri
		}
	}

	return aitems
}

// FetchAccounts
func (p *Provider) FetchAccountTransactions(ctx context.Context,
	account domain.Account,
	cfg Config, stream string,
	peekId string,
	nextUri string,
) ([]domain.RawItem, string) {

	aitems := []domain.RawItem{}
	cNextUri := nextUri
	for {

		tPageData, err := p.HTTP.ListTransactions(ctx, cfg, peekId, "transactions", cNextUri)
		if err != nil {
			break
		}
		for _, raw := range tPageData.Data {
			var peek CnbPeek
			if err := json.Unmarshal(raw, &peek); err != nil {
				slog.Error("FetchRaw", "UnMarshallError", err)
				continue
			}
			item := core.GetRawItem(account, p.Name(), "transactions", peek.ID, raw, peek.CreatedAt)
			aitems = append(aitems, item)
		}

		if len(tPageData.Pagination.Next_uri) == 0 {
			break
		} else {
			cNextUri = tPageData.Pagination.Next_uri
		}
	}

	return aitems, cNextUri
}
