package ethereum

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nanmu42/etherscan-api"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type EthereumTransformer struct {
	logger *logger.Logger
	debug  bool
}

func NewEthereumTransformer(logConfig *logger.Config) EthereumTransformer {
	plog := logConfig.For("refresher.ethereum")
	return EthereumTransformer{plog, false}
}

func (s EthereumTransformer) Name() string {
	return "ethereum"
}

func (s EthereumTransformer) Transform(ctx context.Context, ps core.PriceService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	// gather the ethereum accounts
	eaccts := []domain.Account{}
	for _, acred := range acreds {
		eaccts = append(eaccts, acred.Account)
	}

	actvs := []*domain.Activity{}
	// s.logger.Info("Transform", "Provider", s.Name(), "Actvs", len(actvs))
	tsfrsm, itxns, ntxns := s.marshalData(rawsm)
	s.logger.Info("Transform", s.Name(), fmt.Sprintf("Transfers: %d InternalTxs: %d NormalTxs: %d", len(tsfrsm), len(itxns), len(ntxns)))
	hashProcessed := map[string]string{}

	sort.Slice(ntxns, func(i, j int) bool {
		return ntxns[i].TimeStamp.Time().Before(ntxns[j].TimeStamp.Time())
	})

	for i, ntxn := range ntxns {
		date := ntxn.TimeStamp.Time()
		tsfrs := tsfrsm[ntxn.Hash]
		s.debug = true

		if s.debug {
			s.logger.Info("Transform")
			s.logger.Info("Transform", "Normal", ntxn.Hash, "Date", date)
		}

		nactvs := s.buildActivityFromNormal(ps, eaccts, ntxn, tsfrs)
		actvs = append(actvs, nactvs...)
		hashProcessed[ntxn.Hash] = ntxn.Hash
		if i > 5 {
			break
		}
	}

	// get orders list of hashes
	hashes := s.sortTransfers(tsfrsm)
	for i, hash := range hashes {
		if _, ok := hashProcessed[hash]; ok {
			continue
		}
		tsfrs := tsfrsm[hash]
		date := tsfrs[0].TimeStamp.Time()

		s.debug = true
		if s.debug {
			s.logger.Info("Transform")
			s.logger.Info("Transform", "erc20", hash, "Date", date)
		}

		tactvs := s.buildActivityFromErc20(ps, eaccts, tsfrs)
		actvs = append(actvs, tactvs...)

		if i > 5 {
			break
		}
	}

	s.logger.Info("Transform", "Provider", s.Name(), "Actvs", len(actvs))

	return actvs, nil
}

func (s EthereumTransformer) buildActivityFromNormal(
	ps core.PriceService, eaccts []domain.Account,
	ntxn etherscan.NormalTx, tsfrs []etherscan.ERC20Transfer,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	method := ""
	if len(ntxn.Input) >= 10 {
		method = ntxn.Input[0:10]
	}

	if strings.Compare(ntxn.Input, "0x") == 0 {
		actv := s.createNormalTxnActivity(ps, eaccts, ntxn)
		if actv != nil {
			actvs = append(actvs, actv)
			return actvs
		}
	}

	//
	sel := knownSelectors[method]

	switch len(tsfrs) {
	case 0:
		actv := s.createNormalTxnActivity(ps, eaccts, ntxn)
		if actv != nil {
			actv.TxnType = sel.TxnType
			actvs = append(actvs, actv)
		}

	case 1:
		switch sel.Category {
		// case CategorySwap:
		// case CategoryApprove:
		case CategoryReward, CategoryDeposit, CategoryWithdrawal, CategoryTransfer:
			actv := s.createErc20Activity(ps, eaccts, tsfrs[0])
			if actv != nil {
				actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}
		}
	default:
		s.logger.Error("buildNormalErc20", "Warning", "Not implemented", "Transfers", len(tsfrs))
	}

	return actvs
}

func (s EthereumTransformer) buildActivityFromErc20(
	ps core.PriceService, eaccts []domain.Account, tsfrs []etherscan.ERC20Transfer,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	switch len(tsfrs) {
	case 1:
		actv := s.createErc20Activity(ps, eaccts, tsfrs[0])
		if actv != nil {
			actvs = append(actvs, actv)
		}
	default:
		s.logger.Error("buildActivityFromErc20", "Warning", "Not implemented")
	}

	return actvs
}

func (s EthereumTransformer) marshalData(rawsm map[string][]domain.RawItem,
) (map[string][]etherscan.ERC20Transfer, []etherscan.InternalTx, []etherscan.NormalTx) {

	var itxns []etherscan.InternalTx
	var ntxns []etherscan.NormalTx

	tsfrsm := make(map[string][]etherscan.ERC20Transfer)
	itxnsm := make(map[string]string)
	ntxnsm := make(map[string]string)

	for _, raws := range rawsm {
		for _, raw := range raws {

			s.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
			switch raw.Stream {
			case "erc20":

				var tsfr etherscan.ERC20Transfer
				bytes, err := json.Marshal(raw.Payload)
				if err == nil {
					err = json.Unmarshal(bytes, &tsfr)
				}
				tsfrsm[tsfr.Hash] = append(tsfrsm[tsfr.Hash], tsfr)

			case "internal":
				var itxn etherscan.InternalTx
				bytes, err := json.Marshal(raw.Payload)
				if err == nil {
					err = json.Unmarshal(bytes, &itxn)
				}
				if _, ok := itxnsm[itxn.Hash]; !ok {
					itxns = append(itxns, itxn)
					itxnsm[itxn.Hash] = itxn.Hash
				}

			case "normal":
				var ntxn etherscan.NormalTx
				bytes, err := json.Marshal(raw.Payload)
				if err == nil {
					err = json.Unmarshal(bytes, &ntxn)
				}
				if _, ok := ntxnsm[ntxn.Hash]; !ok {
					ntxns = append(ntxns, ntxn)
					ntxnsm[ntxn.Hash] = ntxn.Hash
				}

			}
		}
	}
	return tsfrsm, itxns, ntxns
}

func (s EthereumTransformer) sortTransfers(tsfrsm map[string][]etherscan.ERC20Transfer) []string {

	// Now get a date-ordered list of hashes to process
	hashes := make([]string, 0, len(tsfrsm))
	for h := range tsfrsm {
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		// compare using the FIRST transfer's timestamp for each hash — all transfers
		// under one hash share the same transaction, so any one of them works
		return time.Time(tsfrsm[hashes[i]][0].TimeStamp).Before(time.Time(tsfrsm[hashes[j]][0].TimeStamp))
	})

	return hashes
}
