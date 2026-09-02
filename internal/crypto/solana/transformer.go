package solana

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

type SolanaAccountTransformer struct {
	logger *logger.Logger
}

func NewSolanaAccountTransformer(logConfig *logger.Config) SolanaAccountTransformer {
	plog := logConfig.For("refresher.solana")
	return SolanaAccountTransformer{plog}
}

func (s SolanaAccountTransformer) Name() string {
	return "solana"
}

func (s SolanaAccountTransformer) Transform(ctx context.Context, ps core.PriceService,
	spamService core.CryptoSpamService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}
	s.logger.Info("Transform", "Provider", s.Name(), "Actvs", len(actvs))

	txnsm := make(map[string]SolanaParsedTransaction)
	tokenAccountsm := make(map[string]SolanaTokenAccount)
	stakeAmountm := make(map[string]decimal.Decimal)

	for _, acred := range acreds {
		raws := rawsm[acred.Account.ID]
		if len(raws) == 0 {
			continue
		}
		atxns, tokenAccounts := s.marshalData(raws)
		for _, txn := range atxns {
			txnsm[txn.Signature] = txn
		}
		for _, tokenAccount := range tokenAccounts {
			tokenAccountsm[tokenAccount.Pubkey] = tokenAccount
		}

	}

	txns := []SolanaParsedTransaction{}
	for _, txn := range txnsm {
		txns = append(txns, txn)
	}
	sort.Slice(txns, func(i, j int) bool {
		return txns[i].Date.Before(*txns[j].Date)
	})

	// txns := s.getTransactionsFromRaw(acreds, rawsm)
	debug := false

	// gather the solana accounts
	saccts := []domain.Account{}
	for _, acred := range acreds {
		saccts = append(saccts, acred.Account)
	}
	for i, txn := range txns {
		debug = false
		if i > 160 {
			// debug = true
		}
		if strings.Compare(txn.Signature, "4r375g95Zy4B8DtmE37rdfcoXX8Tp6C6pu3rwSQ4foy8A6HvoiKEyTnbRBUFrMSzcBhAzF4sFDAHMbDEVv3yns9n") == 0 {
			// debug = true
		}
		if debug {
			s.logger.Info("")
			s.logger.Info("Transform", "Transaction", txn.Signature, "Date", txn.Date)
			s.logger.Info("Transform", "Account", txn.Acct_Id)
			s.logger.Info("Transform", "Address", txn.Address)
		}
		// debug = true
		sActivity := NewSolanaActivity(saccts, ps, spamService, tokenAccountsm, stakeAmountm, txn, s.logger, debug)
		tactvs := sActivity.ProcessTransaction()
		if debug {
			s.logger.Debug("Transform", "Transaction", txn.Signature, "Activities", len(tactvs))
		}
		actvs = append(actvs, tactvs...)
		if i > 180 {
			// break
		}
	}
	s.logger.Info("Transform", "Provider", s.Name(), "Actvs", len(actvs))

	return actvs, nil
}

func (s SolanaAccountTransformer) marshalData(raws []domain.RawItem) ([]SolanaParsedTransaction, []SolanaTokenAccount) {

	var txns []SolanaParsedTransaction
	var tokenAccounts []SolanaTokenAccount
	for _, raw := range raws {
		s.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
		switch raw.Stream {
		case "transactions":

			var txn SolanaParsedTransaction
			bytes, err := json.Marshal(raw.Payload)
			if err == nil {
				err = json.Unmarshal(bytes, &txn)
			}
			txns = append(txns, txn)
		case "tokens":
			var tokenAccount SolanaTokenAccount
			bytes, err := json.Marshal(raw.Payload)
			if err == nil {
				err = json.Unmarshal(bytes, &tokenAccount)
			}
			tokenAccounts = append(tokenAccounts, tokenAccount)
		}
	}
	return txns, tokenAccounts
}
