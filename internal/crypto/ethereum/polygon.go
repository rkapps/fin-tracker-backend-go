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

type PolygonTransformer struct {
	logger *logger.Logger
	debug  bool
}

func NewPolygonTransformer(logConfig *logger.Config) PolygonTransformer {
	plog := logConfig.For("refresher.polygon")
	return PolygonTransformer{plog, false}
}

func (p PolygonTransformer) Name() string {
	return "polygon"
}

func (p PolygonTransformer) Transform(ctx context.Context, ps core.PriceService,
	spamService core.CryptoSpamService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	// gather the polygon accounts
	eaccts := []domain.Account{}
	for _, acred := range acreds {
		eaccts = append(eaccts, acred.Account)
	}

	actvs := []*domain.Activity{}

	tsfrsm, itxns, ntxns := p.marshalData(rawsm)
	p.logger.Info("Transform", p.Name(), fmt.Sprintf("Transfers: %d InternalTxs: %d NormalTxs: %d", len(tsfrsm), len(itxns), len(ntxns)))
	hashProcessed := map[string]string{}

	sort.Slice(ntxns, func(i, j int) bool {
		return ntxns[i].TimeStamp.Time().Before(ntxns[j].TimeStamp.Time())
	})

	for i, ntxn := range ntxns {

		if ntxn.IsError == 1 {
			continue
		}

		// if spamService.IsSpamEthereumFromAddress(ntxn.From, crypto.BLOCKCHAIN_ETHEREUM) {
		// 	p.logger.Debug("Transform", "Spam", fmt.Sprintf("FAddress: %s", ntxn.From))
		// 	continue
		// }

		date := ntxn.TimeStamp.Time()
		tsfrs := tsfrsm[ntxn.Hash]
		p.debug = false
		if strings.Compare(ntxn.Hash, "0xbd81a4c587115fd2f99703a75dec7fd678659a960277f77ebd266fac94fa345c") == 0 {
			// p.debug = true
		}
		if i > 100 {
			// p.debug = true
		}

		if p.debug {
			p.logger.Info("Transform")
			p.logger.Info("Transform", "Normal", ntxn.Hash, "Date", date)
			p.logger.Info("Transform", "Transfers", len(tsfrs))
			p.logger.Info("Transform", "Input", ntxn.Input)
		}

		nactvs := p.buildActivityFromNormal(ps, eaccts, ntxn, tsfrs)
		actvs = append(actvs, nactvs...)
		hashProcessed[ntxn.Hash] = ntxn.Hash
		if i >= 2 {
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

		p.debug = false
		if strings.Compare(hash, "0xc93f352ce2e95d7ca0a405cd82d2391c9cd90b7e5d7fee41cf194d5eaa7186c2") == 0 {
			// p.debug = true
		}
		if p.debug {
			p.logger.Info("Transform")
			p.logger.Info("Transform", "erc20", hash, "Date", date)
			p.logger.Info("Transform", "Caddress", tsfr.ContractAddress)
			p.logger.Info("Transform", "Transfers--", len(tsfrs))
		}

		if spamService.IsSpamEthereumContractAddress(tsfr.ContractAddress, crypto.BLOCKCHAIN_POLYGON) {
			p.logger.Debug("Transform", "Spam", fmt.Sprintf("CAddress: %s", tsfr.ContractAddress))
			continue
		}

		spam := false
		for _, tsfr1 := range tsfrs {
			if len(tsfr1.TokenSymbol) == 0 {
				continue
			}

			if spamService.IsSpamEthereumSymbol(tsfr1.TokenSymbol, crypto.BLOCKCHAIN_POLYGON) {
				// p.logger.Info("Transform", "Spam", fmt.Sprintf("Symbol: %s", tsfr1.TokenSymbol))
				spam = true
				break
			}
		}
		if spam {
			continue
		}

		tactvs := p.buildActivityFromErc20(ps, eaccts, tsfrs, POLYGON_BASE_CURRENCY)
		actvs = append(actvs, tactvs...)

		if i >= 85 {
			// break
		}
	}

	// for _, itxn := range itxns {
	// 	if _, ok := hashProcessed[itxn.Hash]; ok {
	// 		continue
	// 	}
	// 	p.logger.Info("Transform", "internal", itxn.Hash, "Date", itxn.TimeStamp.Time())
	// actv := createInternalTxnActivity(ps, eaccts, itxn, p.logger)
	// if actv != nil {
	// 	actvs = append(actvs, actv)
	// }
	// }

	p.logger.Info("Transform", "Provider", p.Name(), "Actvs", len(actvs))

	return actvs, nil
}

func (p PolygonTransformer) buildActivityFromNormal(
	ps core.PriceService, eaccts []domain.Account,
	ntxn etherscan.NormalTx, tsfrs []etherscan.ERC20Transfer,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	method := ""
	if len(ntxn.Input) >= 10 {
		method = ntxn.Input[0:10]
	}

	if strings.Compare(ntxn.Input, "0x") == 0 {
		actv := createNormalTxnActivity(ps, eaccts, ntxn, POLYGON_BASE_CURRENCY, Selector{}, p.logger, p.debug)
		if actv != nil {
			actvs = append(actvs, actv)
			return actvs
		}
		return nil
	}

	sel := Selector{}
	switch method {

	case "0xd0e30db0":
		//wrap
		actv := createNormalWrapActivity(ps, eaccts, ntxn, POLYGON_BASE_CURRENCY, POLYGON_WRAP_CURRENCY, p.logger, p.debug)
		if actv != nil {
			// actv.TxnType = sel.TxnType
			actvs = append(actvs, actv)
		}

	case "0x095ea7b3":
		// approve
	case "0xb95cac28":
		//liquid activities
		// p.logger.Info("buildActivityFromNormal", method, ntxn.Hash)
		mactvs := createErcMultiActivity(ps, eaccts, tsfrs, POLYGON_BASE_CURRENCY, p.logger, p.debug)
		actvs = append(actvs, mactvs...)

	// case "0x3ce33bff":
	// 	// wrap
	// 	//wrap
	// 	actv := createNormalWrapActivity(ps, eaccts, ntxn, POLYGON_BASE_CURRENCY, POLYGON_WRAP_CURRENCY, p.logger, p.debug)
	// 	if actv != nil {
	// 		// actv.TxnType = sel.TxnType
	// 		actvs = append(actvs, actv)
	// 	}
	case "0x2e1a7d4d", "0x3ce33bff":
		// bridge send
		switch len(tsfrs) {
		case 1:

			actv := createErc20Activity(ps, eaccts, tsfrs[0], POLYGON_BASE_CURRENCY, sel, p.logger, p.debug)
			if actv != nil {
				actvs = append(actvs, actv)
			}
			// log.Println(ntxn.Hash)
		default:

			actv := createNormalTxnActivity(ps, eaccts, ntxn, POLYGON_BASE_CURRENCY, sel, p.logger, p.debug)
			if actv != nil {
				// actv.TxnType = sel.TxnType
				actvs = append(actvs, actv)
			}

		}

	case "0xa9059cbb":
		// transfer
		if len(tsfrs) == 1 {
			actv := createErc20Activity(ps, eaccts, tsfrs[0], POLYGON_BASE_CURRENCY, sel, p.logger, p.debug)
			if actv != nil {
				actvs = append(actvs, actv)
			}
		}

	case "0x34fcd5be":

	case "0xa1798512":

	}

	return actvs
}

func (p PolygonTransformer) buildActivityFromErc20(
	ps core.PriceService, eaccts []domain.Account, tsfrs []etherscan.ERC20Transfer, baseSymbol string,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	switch len(tsfrs) {
	case 1:
		actv := createErc20Activity(ps, eaccts, tsfrs[0], baseSymbol, Selector{}, p.logger, p.debug)
		if actv != nil {
			actvs = append(actvs, actv)
		}
	// case 5:
	// 	return createErcMultiActivity(ps, eaccts, tsfrs, POLYGON_BASE_CURRENCY, p.logger, p.debug)
	default:
		p.logger.Error("buildActivityFromErc20", "Warning", "Not implemented", "Transfers", len(tsfrs))
	}

	return actvs
}

func (p PolygonTransformer) marshalData(rawsm map[string][]domain.RawItem,
) (map[string][]etherscan.ERC20Transfer, []etherscan.InternalTx, []etherscan.NormalTx) {

	var itxns []etherscan.InternalTx
	var ntxns []etherscan.NormalTx

	tsfrsm := make(map[string][]etherscan.ERC20Transfer)
	itxnsm := make(map[string]string)
	ntxnsm := make(map[string]string)

	for _, raws := range rawsm {
		for _, raw := range raws {

			p.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
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
