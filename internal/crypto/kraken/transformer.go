package kraken

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

type KrakenAccountTransformer struct {
	logger *logger.Logger
}

func NewKrakenAccountTransformer(logConfig *logger.Config) KrakenAccountTransformer {
	plog := logConfig.For("refresher.kraken")
	return KrakenAccountTransformer{plog}
}

func (k KrakenAccountTransformer) Name() string {
	return "kraken"
}

func (k KrakenAccountTransformer) Transform(ctx context.Context, ps core.PriceService,
	spamService core.CryptoSpamService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}

	// get assets and pairs from global raw
	assetsm, pairsm := k.marshalGlobalData(globalRaws)

	for _, acred := range acreds {
		raws := rawsm[acred.Account.ID]
		if len(raws) == 0 {
			continue
		}
		k.logger.Debug("Transform", "Items", len(raws))

		trades, ledgers := k.marshalData(raws)
		k.logger.Info("Transform", "Trades", len(trades), "Ledgers", len(ledgers))
		//trades
		for _, trade := range trades {
			// k.logger.Info("Transform", "Trade", trade.OrderTxID, "AssetPair", trade.AssetPair)
			ts, _ := crypto.KrakenTime(trade.Time)
			tpair := pairsm[trade.AssetPair]
			baseSymbol := assetsm[tpair.Base]
			quoteSymbol := assetsm[tpair.Quote]

			actv := &domain.Activity{}
			actv.UID = acred.Account.UID
			actv.AccountID = acred.Account.ID
			actv.ID = trade.OrderTxID
			actv.Date = *ts

			switch trade.Type {
			case string(domain.ActivityTypeBuy):
				actv.SentAccountID = acred.Account.ID
				actv.SentSymbol = quoteSymbol.Altname
				actv.SentAmount = utils.ConvertFloatToDecimal(trade.Cost)
				actv.RcvAccountID = acred.Account.ID
				actv.RcvSymbol = baseSymbol.Altname
				actv.RcvAmount = utils.ConvertFloatToDecimal(trade.Volume)
				if crypto.IsCurrency(actv.SentSymbol) {
					actv.TxnType = domain.ActivityTypeBuy
				} else {
					actv.TxnType = domain.ActivityTypeTrade
					price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
					actv.SentPrice = price
					price, _ = ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
					actv.RcvPrice = price
				}
			case string(domain.ActivityTypeSell):

				actv.SentAccountID = acred.Account.ID
				actv.SentSymbol = baseSymbol.Altname
				actv.SentAmount = utils.ConvertFloatToDecimal(trade.Volume)
				actv.RcvAccountID = acred.Account.ID
				actv.RcvSymbol = quoteSymbol.Altname
				actv.RcvAmount = utils.ConvertFloatToDecimal(trade.Cost)

				if crypto.IsCurrency(actv.RcvSymbol) {
					actv.TxnType = domain.ActivityTypeSell
				} else {
					actv.TxnType = domain.ActivityTypeTrade
					price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
					actv.SentPrice = price
					price, _ = ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
					actv.RcvPrice = price
				}
			default:
				k.logger.Error("Transform", "Trade", trade.OrderTxID, "ErrorType", trade.Type)
			}
			actvs = append(actvs, actv)
		}

		// ledgers
		for i, ledger := range ledgers {

			ts, err := crypto.KrakenTime(ledger.Time)
			if err != nil {
				k.logger.Error("Transform", "Ledger", ledger.InfoId, "Error", err)
			}

			actv := &domain.Activity{}
			actv.UID = acred.Account.UID
			actv.AccountID = acred.Account.ID
			actv.ID = ledger.InfoId
			actv.Date = *ts

			amount, _ := utils.ConvertStringToDecimal(ledger.Amount)
			symbol := normalizeKrakenAsset(ledger.Asset)
			if asymbol, ok := assetsm[symbol]; ok {
				symbol = asymbol.Altname
			}

			if len(symbol) == 0 {
				k.logger.Error("Transform", "Ledger", ledger.InfoId, "Error", ledger.Asset)
				continue
			}
			fee, _ := utils.ConvertStringToDecimal(ledger.Fee)
			actv.Fee = fee
			actv.FeeCurrency = symbol

			k.logger.Debug("Transform", "Ledger", ledger.Debug())

			switch ledger.Type {
			case "withdrawal":
				actv.TxnType = domain.ActivityTypeSend
				actv.SentAccountID = actv.AccountID
				actv.SentSymbol = symbol
				actv.SentAmount = amount.Neg()

				if strings.Compare(actv.ID, "LC5PIG-5QLXI-J25DVS") == 0 || strings.Compare(actv.ID, "LRH5YV-5HJKI-GLPD2Q") == 0 {
					actv.Date = actv.Date.Add(time.Minute * -2)
				}

			case "deposit":
				actv.TxnType = domain.ActivityTypeReceive
				actv.RcvAccountID = actv.AccountID
				actv.RcvSymbol = symbol
				actv.RcvAmount = amount
			case "staking":
				actv.TxnType = domain.ActivityTypeReward
				actv.RcvAccountID = actv.AccountID
				actv.RcvSymbol = symbol
				actv.RcvAmount = amount
				price, err := ps.GetCryptoPrice(symbol, actv.Date)
				if err != nil {
					k.logger.Error("Transform", "Ledger", ledger.Debug(), "Error", err)
					// continue
				}
				actv.SentAmount = amount.Mul(price)
				actv.SentSymbol = "USD"

			default:
				// log.Printf("Default: %v", )
				continue
			}
			actvs = append(actvs, actv)

			if i > 10 {
				// break
			}
		}
	}
	k.logger.Info("Transform", "Provider", k.Name(), "Actvs", len(actvs))

	return actvs, nil
}

func (r KrakenAccountTransformer) marshalGlobalData(raws []domain.RawItem) (
	map[string]AssetInfo,
	map[string]AssetPairInfo,
) {
	assetsm := make(map[string]AssetInfo)
	pairsm := make(map[string]AssetPairInfo)
	for _, raw := range raws {
		r.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
		switch raw.Stream {
		case "assets":

			bytes, err := json.Marshal(raw.Payload)
			// log.Println(string(bytes))
			if err == nil {
				err = json.Unmarshal(bytes, &assetsm)
			}

		case "assetpairs":

			bytes, err := json.Marshal(raw.Payload)
			if err == nil {
				err = json.Unmarshal(bytes, &pairsm)
			}
		}
	}
	return assetsm, pairsm

}

func (r KrakenAccountTransformer) marshalData(raws []domain.RawItem) (
	[]TradeHistoryInfo,
	[]LedgerInfo,
) {

	var trades []TradeHistoryInfo
	var ledgers []LedgerInfo

	for _, raw := range raws {
		r.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
		switch raw.Stream {
		case "trades":

			var trade TradeHistoryInfo
			bytes, err := json.Marshal(raw.Payload)
			if err == nil {
				err = json.Unmarshal(bytes, &trade)
			}
			trades = append(trades, trade)

		case "ledgers":

			var ledger LedgerInfo
			bytes, err := json.Marshal(raw.Payload)
			if err == nil {
				err = json.Unmarshal(bytes, &ledger)
			}
			ledgers = append(ledgers, ledger)
		}
	}
	return trades, ledgers
}

func normalizeKrakenAsset(code string) string {
	// Strip Earn/staking suffix: "SOL03.S" -> "SOL03" -> "SOL"
	if i := strings.Index(code, "."); i != -1 {
		code = code[:i]
	}
	// Strip trailing digits some staking variants carry (e.g. "SOL03" -> "SOL")
	code = strings.TrimRight(code, "0123456789")
	return code
}
