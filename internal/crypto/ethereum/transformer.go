package ethereum

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/nanmu42/etherscan-api"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
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
	spamService core.CryptoSpamService,
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

		if ntxn.IsError == 1 {
			continue
		}

		if spamService.IsSpamEthereumFromAddress(ntxn.From, crypto.BLOCKCHAIN_ETHEREUM) {
			s.logger.Debug("Transform", "Spam", fmt.Sprintf("FAddress: %s", ntxn.From))
			continue
		}

		date := ntxn.TimeStamp.Time()
		tsfrs := tsfrsm[ntxn.Hash]
		s.debug = false

		if strings.Compare(ntxn.Hash, "0xe4a72d11a8736326ce4940b3facb5a00a07c11cf700886e26a10f28bec999134") == 0 {
			// s.debug = true
		}
		if i > 100 {
			// s.debug = true
		}

		if s.debug {
			s.logger.Info("Transform")
			s.logger.Info("Transform", "Normal", ntxn.Hash, "Date", date)
			s.logger.Info("Transform", "Transfers", len(tsfrs))
			s.logger.Info("Transform", "Error", ntxn.IsError)
		}

		nactvs := s.buildActivityFromNormal(ps, eaccts, ntxn, tsfrs)
		actvs = append(actvs, nactvs...)
		hashProcessed[ntxn.Hash] = ntxn.Hash
		if i > 120 {
			// break
		}
	}

	// get orders list of hashes
	hashes := sortTransfers(tsfrsm)
	for i, hash := range hashes {
		if _, ok := hashProcessed[hash]; ok {
			continue
		}
		tsfrs := tsfrsm[hash]
		tsfr := tsfrs[0]

		date := tsfr.TimeStamp.Time()
		s.debug = false

		if spamService.IsSpamEthereumContractAddress(tsfr.ContractAddress, crypto.BLOCKCHAIN_ETHEREUM) {
			s.logger.Debug("Transform", "Spam", fmt.Sprintf("CAddress: %s", tsfr.ContractAddress))
			continue
		}
		if spamService.IsSpamEthereumSymbol(tsfr.TokenSymbol, crypto.BLOCKCHAIN_ETHEREUM) {
			s.logger.Debug("Transform", "Spam", fmt.Sprintf("Symbol: %s", tsfr.TokenSymbol))
			continue
		}

		if strings.Compare(tsfr.Hash, "0xe4a72d11a8736326ce4940b3facb5a00a07c11cf700886e26a10f28bec999134") == 0 {
			continue
		}
		if i > 55 {
			// s.debug = true
		}
		if s.debug {
			s.logger.Info("Transform")
			s.logger.Info("Transform", "erc20", hash, "Date", date)
			s.logger.Info("Transform", "Transfers--", len(tsfrs))
		}

		tactvs := s.buildActivityFromErc20(ps, eaccts, tsfrs)
		actvs = append(actvs, tactvs...)

		if i > 60 {
			// break
		}
	}

	for _, itxn := range itxns {
		if _, ok := hashProcessed[itxn.Hash]; ok {
			continue
		}
		actv := createInternalTxnActivity(ps, eaccts, itxn, s.logger, s.debug)
		if actv != nil {
			actvs = append(actvs, actv)
		}
	}

	s.logger.Info("Transform", "Provider", s.Name(), "Actvs", len(actvs))
	for _, actv := range actvs {
		if strings.Compare(actv.RcvSymbol, "MATIC") == 0 {
			actv.RcvSymbol = "POL"
		}
		if strings.Compare(actv.SentSymbol, "MATIC") == 0 {
			actv.SentSymbol = "POL"
		}
	}

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
		actv := createNormalTxnActivity(ps, eaccts, ntxn, ETH_BASE_CURRENCY, Selector{}, s.logger, s.debug)
		if actv != nil {
			actvs = append(actvs, actv)
			return actvs
		}
	}

	//
	sel := knownSelectors[method]
	if sel.Category == CategoryUnknown {
		// s.logger.Warn("buildNormalActivity", "Method", method, "Warning", "Selector Unknow")
		return nil
	}

	if len(sel.Category) == 0 {
		s.logger.Error("buildNormalActivity", "Txn", ntxn.Hash, "Error", "Selector is blank")
		return nil
	}
	if s.debug {
		s.logger.Info("Transform", "Selector", sel.Category)
	}

	switch len(tsfrs) {
	case 0:
		switch sel.Category {
		case CategoryWrap:
			actv := createNormalWrapActivity(ps, eaccts, ntxn, ETH_BASE_CURRENCY, ETH_WRAP_CURRENCY, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}
		default:
			actv := createNormalTxnActivity(ps, eaccts, ntxn, ETH_BASE_CURRENCY, sel, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}
		}

	case 1:
		switch sel.Category {
		case CategoryWrap:
			actv := createErc20WrapActivity(ps, eaccts, tsfrs[0], ETH_WRAP_CURRENCY, sel, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}
		case CategorySwap:
			actv := createErc20WrapActivity(ps, eaccts, tsfrs[0], ETH_BASE_CURRENCY, sel, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}

		// case CategoryApprove:
		case CategoryReward, CategoryDeposit, CategoryWithdrawal, CategoryTransfer:
			actv := createErc20Activity(ps, eaccts, tsfrs[0], ETH_BASE_CURRENCY, sel, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}
		}
	case 2:
		switch sel.Category {
		case CategoryTransfer:
			actv := createErc20Activity(ps, eaccts, tsfrs[0], ETH_BASE_CURRENCY, sel, s.logger, s.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}

		case CategorySwap, CategoryWrap:
			actv := createErc20TradeActivity(ps, eaccts, tsfrs[0], tsfrs[1], sel, s.logger, s.debug)
			if actv == nil {
				actv = createErc20TradeActivity(ps, eaccts, tsfrs[1], tsfrs[0], sel, s.logger, s.debug)
			}
			if actv != nil {
				actvs = append(actvs, actv)
			}
		case CategoryMultiToken:
			mactvs := createErcMultiActivity(ps, eaccts, tsfrs, ETH_BASE_CURRENCY, s.logger, s.debug)
			actvs = append(actvs, mactvs...)

			// add normal actvitity only if positive amount
			amount, _ := crypto.ConvertStringToBaseDecimal(ntxn.Value.Int().String(), TXN_DECIMALS)
			if amount.IsPositive() {
				actv := createNormalTxnActivity(ps, eaccts, ntxn, ETH_BASE_CURRENCY, sel, s.logger, s.debug)
				if actv != nil {
					actv.TxnType = domain.ActivityTypeAddLiquidity
					actvs = append(actvs, actv)
				}
			}

		case CategoryWithdrawal:
			// 0x2e1a7d4d
			// withdraw
			// hash - 0x3cd0719a3605fcdec3e78d557609eb0c58534b2e72983aeaf908319ef1de63a6
		}
	case 3:
		mactvs := createErcMultiActivity(ps, eaccts, tsfrs, ETH_BASE_CURRENCY, s.logger, s.debug)
		actvs = append(actvs, mactvs...)

	default:
		s.logger.Error("buildNormalErc20", "Warning", fmt.Sprintf("%s Not implemented", sel.TxnType), "Transfers", len(tsfrs))
	}

	return actvs
}

func (s EthereumTransformer) buildActivityFromErc20(
	ps core.PriceService, eaccts []domain.Account, tsfrs []etherscan.ERC20Transfer,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	switch len(tsfrs) {
	case 1:
		actv := createErc20Activity(ps, eaccts, tsfrs[0], ETH_BASE_CURRENCY, Selector{}, s.logger, s.debug)
		if actv != nil {
			actvs = append(actvs, actv)
		}
	case 2:
		actv := createErc20Activity(ps, eaccts, tsfrs[0], ETH_BASE_CURRENCY, Selector{}, s.logger, s.debug)
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
