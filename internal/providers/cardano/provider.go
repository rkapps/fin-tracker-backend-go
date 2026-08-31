package cardano

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
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
var _ core.GlobalSyncProvider = (*Provider)(nil)

func New(api API, logConfig *logger.Config) *Provider {
	slog := logConfig.For("syncer.cardano")
	return &Provider{HTTP: api, Limiter: rate.NewLimiter(rate.Limit(10), 10), logger: slog}
}

// Name implements [provider.SyncSourceProvider].
func (p *Provider) Name() string {
	return "cardano"
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) Streams() []string {
	return []string{"rewards", "transactions"}
	// return []string{"withdrawals"}
}

// Streams implements [provider.SyncSourceProvider].
func (p *Provider) GlobalStreams() []string {
	return []string{"epochs"}
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
		return p.fetchRewards(ctx, acred.Account, stream, cursor)
	case "transactions":
		return p.fetchTransactions(ctx, acred.Account, stream, cursor)
		// case "withdrawals":
		// 	return p.fetchWithdrawals(acred.Account, stream, "")
	}
	return core.SyncPage{}, nil
}

// FetchRaw implements [provider.SyncSourceProvider].
func (p *Provider) FetchGlobalRaw(ctx context.Context,
	stream string, cursor string,
) (core.SyncPage, error) {

	if err := p.Limiter.Wait(ctx); err != nil {
		return core.SyncPage{}, err
	}

	switch stream {
	case "epochs":
		return p.fetchEpochs(ctx, stream, cursor)
	}
	return core.SyncPage{}, nil

}

func (p Provider) fetchRewards(ctx context.Context, account domain.Account, stream string, cursor string) (core.SyncPage, error) {

	rCount := 10
	address := account.Address()
	if !strings.HasPrefix(address, "stake") {
		return core.SyncPage{HasMore: false}, nil // not a stake address — nothing to fetch, not an error
	}

	var cur cardanoRewardsCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.Page = 1
	}
	p.logger.Debug("fetchRewards", "Address", address, "Cursor", fmt.Sprintf("%s-%v", cursor, cur))

	rewards, err := p.HTTP.GetAccountRewards(ctx, address, cur.Page, rCount)
	if err != nil {
		p.logger.Error("fetchRewards", "GetAccountRewards", err)
		// log.Println(err)
		return core.SyncPage{}, err
	}
	p.logger.Debug("fetchRewards", "Rewards", len(rewards))

	var items []domain.RawItem
	for _, r := range rewards {
		raw, _ := json.Marshal(r)
		item := providers.GetRawItem(account, p.Name(), stream, fmt.Sprintf("%s-%d", address, r.Epoch), raw, time.Now())
		items = append(items, item)
	}

	// const pageSize = 10 // confirm Blockfrost's actual default for this endpoint
	hasMore := len(rewards) >= rCount
	if hasMore {
		cur.Page++
	}
	p.logger.Debug("fetchRewards", "Items", len(items), "hasMore", hasMore)

	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: hasMore}, nil
}

func (p Provider) fetchTransactions(ctx context.Context, account domain.Account, stream string, cursor string) (core.SyncPage, error) {

	var items []domain.RawItem
	addrs := []string{}
	address := account.Address()
	p.logger.Debug("fetchTransactions", "Account", account.ID, "Address", address)

	var withdrawals []TransactionWithdrawal
	if strings.HasPrefix(address, "stake") {
		withdrawals, _ = p.HTTP.GetTransactionWithdrawals(address)
		saddrs, err := p.HTTP.GetAccountAddresses(ctx, address)
		if err != nil {
			return core.SyncPage{HasMore: false}, nil // not a stake address — nothing to fetch, not an error
		}
		for _, addr := range saddrs {
			addrs = append(addrs, addr.Address)
		}
	} else {
		addrs = append(addrs, address)
	}

	//withdrawals map on txHash
	withdrawalsm := make(map[string]TransactionWithdrawal)
	for _, withdrawal := range withdrawals {
		withdrawalsm[withdrawal.TxHash] = withdrawal
		p.logger.Debug("fetchTransactions", "Withdrawal", withdrawal.TxHash, "Amount", withdrawal.Amount)
	}

	// add addresses
	raw, _ := json.Marshal(addrs)
	// log.Println(string(raw))
	item := providers.GetRawItem(account, p.Name(), "addresses", "addresses", raw, time.Now())
	items = append(items, item)

	var cur cardanoTransactionCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.BlockHeight = 0
	}
	bheight := cur.BlockHeight
	nbheight := cur.BlockHeight
	txHashm := make(map[string]string)

	for _, address := range addrs {

		// fetch transactions
		atxns := p.fetchAllTransactions(ctx, address, bheight)
		p.logger.Debug("fetchTransactions", "Address", address, "Tnxs", len(atxns))
		for _, atxn := range atxns {

			// check for duplicate txHash
			if _, ok := txHashm[atxn.TxHash]; ok {
				continue
			}
			txHashm[atxn.TxHash] = atxn.TxHash

			if atxn.BlockHeight > nbheight {
				nbheight = atxn.BlockHeight
			}

			withdrawal := withdrawalsm[atxn.TxHash]
			// enrich each transaction
			transaction := p.enrichTransaction(atxn)
			transaction.WithdrawalAmount = withdrawal.Amount

			raw, _ := json.Marshal(transaction)
			if strings.Compare(atxn.TxHash, "8e2174ce0da75c2c76e96f909dc1e0efae40e67c5410d9cdbfd9c8161bb07a8f") == 0 ||
				false {
				// log.Println(string(raw))
			}

			// log.Println(string(raw))
			item := providers.GetRawItem(account, p.Name(), stream, atxn.TxHash, raw, time.Now())
			items = append(items, item)
		}
	}
	cur.BlockHeight = nbheight
	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: false}, nil
}

func (p Provider) fetchAllTransactions(ctx context.Context, address string, bheight int64) []AddressTransaction {

	atxns := []AddressTransaction{}
	page := 1
	for {
		txns, _ := p.HTTP.GetAccountTransactions(ctx, address, bheight, page)
		atxns = append(atxns, txns...)
		if len(txns) < 100 {
			break
		}
		page++
	}
	return atxns
}

func (p Provider) enrichTransaction(aTxn AddressTransaction) Transaction {

	info, _ := p.HTTP.GetTransactionInfo(aTxn.TxHash)
	metadata, _ := p.HTTP.GetTransactionMetadata(aTxn.TxHash)
	if len(metadata) == 0 {
		metadata = nil
	}
	scerts, _ := p.HTTP.GetTransactionStakeCerticates(aTxn.TxHash)
	if len(scerts) == 0 {
		scerts = nil
	}
	utxo, _ := p.HTTP.GetTransactionUTXOs(aTxn.TxHash)
	delegations, _ := p.HTTP.GetTransactionDelegations(aTxn.TxHash)

	tx := Transaction{}
	tx.TxHash = aTxn.TxHash
	tx.BlockHeight = aTxn.BlockHeight
	date := time.Unix(aTxn.BlockTime, 10)
	tx.BlockTime = &date
	tx.UTXO = utxo
	tx.Metadata = metadata
	tx.StakeCertificates = scerts
	tx.Delegations = delegations
	if strings.Compare(tx.TxHash, "8e2174ce0da75c2c76e96f909dc1e0efae40e67c5410d9cdbfd9c8161bb07a8f") == 0 {
		// log.Println(info.Fees)
	}
	if len(info.TxHash) > 0 {
		tx.Fees = info.Fees
	}
	return tx
}

func (p Provider) fetchEpochs(ctx context.Context, stream string, cursor string) (core.SyncPage, error) {

	var cur cardanoEpochCursor
	if cursor != "" {
		if err := json.Unmarshal([]byte(cursor), &cur); err != nil {
			return core.SyncPage{}, fmt.Errorf("bad cursor: %w", err)
		}
	} else {
		cur.Epoch = 0 // confirm Cardano's actual first epoch number — likely 0 or 208 depending on era
		cur.Page = 1
	}
	p.logger.Debug("fetchEpochs", "Cursor", cursor, "SavedCursor", cur)
	epochs, err := p.HTTP.GetEpochInformation(ctx, cur.Epoch, cur.Page)
	if err != nil {
		return core.SyncPage{}, err
	}

	var items []domain.RawItem
	for _, e := range epochs {
		raw, _ := json.Marshal(e)
		epochStr := strconv.FormatInt(e.Epoch, 10)
		item := providers.GetGlobalRawItem(p.Name(), stream, epochStr, raw, time.Unix(e.StartTime, 0))
		items = append(items, item)
	}

	const pageSize = 100
	if len(epochs) < pageSize {
		// Caught up. Remember the last epoch seen so the NEXT sync run
		// resumes from here instead of restarting at epoch 0.
		if len(epochs) > 0 {
			cur.Epoch = epochs[len(epochs)-1].Epoch
		}
		cur.Page = 1
		next, _ := json.Marshal(cur)
		return core.SyncPage{Items: items, NextCursor: string(next), HasMore: false}, nil
	}

	// Full page — more to fetch within THIS run.
	cur.Page++
	next, _ := json.Marshal(cur)
	return core.SyncPage{Items: items, NextCursor: string(next), HasMore: true}, nil

}
