package ethereum

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
	"golang.org/x/time/rate"
)

const (
	PAGE_OFFSET = 1000
)

type Provider struct {
	HTTP    API
	name    string
	Limiter *rate.Limiter
	logger  *logger.Logger
}

// compile-time check this satisfies the interface — also self-documenting
var _ core.SyncProvider = (*Provider)(nil)

func NewEthereum(api API, logConfig *logger.Config) *Provider {
	return new(api, "ethereum", logConfig)
}

func new(api API, name string, logConfig *logger.Config) *Provider {
	slog := logConfig.For(fmt.Sprintf("syncer.%s", name))
	return &Provider{HTTP: api, name: name, Limiter: rate.NewLimiter(rate.Every(10*time.Second), 1), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	// ethereum, polygon, optimism
	return p.name
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	return []string{"transactions"}
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
	var cur ethereumCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.BlockNumber = 1
	}

	aitems := []domain.RawItem{}
	mbNumber := cur.BlockNumber

	//get erc20 transfer
	nbNumber, items := p.fetchERC20Transfers(acred, cur.BlockNumber)
	aitems = append(aitems, items...)
	if nbNumber > mbNumber {
		mbNumber = nbNumber
	}

	// //get erc721 transfer
	// nbNumber, items = p.fetchERC721Transfers(acred, cur.BlockNumber)
	// aitems = append(aitems, items...)
	// if nbNumber > mbNumber {
	// 	mbNumber = nbNumber
	// }

	//get internal transactions
	nbNumber, items = p.fetchInternalTransactions(acred, cur.BlockNumber)
	aitems = append(aitems, items...)
	if nbNumber > mbNumber {
		mbNumber = nbNumber
	}

	//get normal transactions
	nbNumber, items = p.fetchNormalTransactions(acred, cur.BlockNumber)
	aitems = append(aitems, items...)
	if nbNumber > mbNumber {
		mbNumber = nbNumber
	}
	// set the max block number
	cur.BlockNumber = mbNumber
	next, _ := json.Marshal(cur)
	hasMore := false

	return core.SyncPage{Items: aitems, NextCursor: string(next), HasMore: hasMore}, nil
}

func (p *Provider) fetchERC20Transfers(acred domain.AccountWithCredential, blockNumber int) (int, []domain.RawItem) {

	var aitems []domain.RawItem
	cblockNumber := blockNumber

	page := 1
	hashm := make(map[string]int)
	for {
		resp, err := p.HTTP.GetERC20Transfers(acred.Account.Address(), &blockNumber, page, PAGE_OFFSET)
		if err != nil {
			p.logger.Error("fetchERC20Transfers", "Error", err)
		}
		p.logger.Info("fetchERC20", "erc20", fmt.Sprintf("Account: %s", acred.Account.ID), "count", len(resp))
		var items []domain.RawItem
		for _, transfer := range resp {
			raw, err := json.Marshal(transfer)
			if err != nil {
				p.logger.Error("fetchERC20Transfers", "error", err)
			}
			idx := hashm[transfer.Hash]
			hashm[transfer.Hash] = idx + 1
			externalID := transfer.Hash // first transfer: bare hash, no suffix
			if idx > 0 {
				externalID = fmt.Sprintf("%s-%d", transfer.Hash, idx) // second (idx=1) → "-1", third (idx=2) → "-2", etc.
			}
			items = append(items, providers.GetRawItem(acred.Account, p.Name(), "erc20", externalID, raw, time.Now()))
			if transfer.BlockNumber > cblockNumber {
				cblockNumber = transfer.BlockNumber
			}
		}
		aitems = append(aitems, items...)
		if len(items) < 1000 {
			break
		}
		page++
	}

	return cblockNumber, aitems
}

func (p *Provider) fetchERC721Transfers(acred domain.AccountWithCredential, blockNumber int) (int, []domain.RawItem) {

	var aitems []domain.RawItem
	cblockNumber := blockNumber

	page := 1
	for {
		resp, err := p.HTTP.GetERC721Transfers(acred.Account.Address(), &blockNumber, page, PAGE_OFFSET)
		if err != nil {
			p.logger.Error("fetchERC721Transfers", "Error", err)
		}
		p.logger.Info("fetchERC721", "erc721", fmt.Sprintf("Account: %s", acred.Account.ID), "count", len(resp))
		var items []domain.RawItem
		for _, transfer := range resp {
			raw, _ := json.Marshal(transfer)
			items = append(items, providers.GetRawItem(acred.Account, p.Name(), "erc721", transfer.Hash, raw, time.Now()))
			if transfer.BlockNumber > cblockNumber {
				cblockNumber = transfer.BlockNumber
			}
		}
		aitems = append(aitems, items...)
		if len(items) < 1000 {
			break
		}
		page++
	}

	return cblockNumber, aitems
}

func (p *Provider) fetchInternalTransactions(acred domain.AccountWithCredential, blockNumber int) (int, []domain.RawItem) {

	var aitems []domain.RawItem
	cblockNumber := blockNumber
	page := 1

	for {

		resp, err := p.HTTP.GetInternalTransactions(acred.Account.Address(), &blockNumber, page, PAGE_OFFSET)
		if err != nil {
			p.logger.Error("fetchInternal", "Error", err)
		}
		p.logger.Info("fetchInternal", "Internal", fmt.Sprintf("Account: %s", acred.Account.ID), "count", len(resp))
		var items []domain.RawItem
		for _, txn := range resp {
			raw, _ := json.Marshal(txn)
			items = append(items, providers.GetRawItem(acred.Account, p.Name(), "internal", txn.Hash, raw, time.Now()))
			if txn.BlockNumber > cblockNumber {
				cblockNumber = txn.BlockNumber
			}
		}

		aitems = append(aitems, items...)
		if len(items) < 1000 {
			break
		}
		page++
	}

	return cblockNumber, aitems
}

func (p *Provider) fetchNormalTransactions(acred domain.AccountWithCredential, blockNumber int) (int, []domain.RawItem) {

	var aitems []domain.RawItem
	cblockNumber := blockNumber
	page := 1

	for {

		resp, err := p.HTTP.GetNormalTransactions(acred.Account.Address(), &blockNumber, page, PAGE_OFFSET)
		if err != nil {
			p.logger.Error("fetchNormal", "Error", err)
		}
		p.logger.Info("fetchNormal", "normal", fmt.Sprintf("Account: %s", acred.Account.ID), "count", len(resp))
		var items []domain.RawItem
		for _, txn := range resp {
			raw, _ := json.Marshal(txn)
			items = append(items, providers.GetRawItem(acred.Account, p.Name(), "normal", txn.Hash, raw, time.Now()))
			if txn.BlockNumber > cblockNumber {
				cblockNumber = txn.BlockNumber
			}
		}

		aitems = append(aitems, items...)
		if len(items) < 1000 {
			break
		}
		page++
	}

	return cblockNumber, aitems
}
