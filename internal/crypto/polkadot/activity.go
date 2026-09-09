package polkadot

import (
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/shopspring/decimal"
)

var (
	BASE_CURRENCY = "DOT"
	TXN_DECIMALS  = decimal.NewFromInt(10)
)

func (p PolkadotTransformer) createRewardActivity(ps core.PriceService, reward PolkadotReward) *domain.Activity {

	var actv *domain.Activity

	tm := time.Unix(reward.Block_Timestamp, 10)
	amount, _ := crypto.ConvertStringToBaseDecimal(reward.Amount, TXN_DECIMALS)
	value, skip := crypto.EvaluateRewardValue(ps, BASE_CURRENCY, amount, tm)
	if skip {
		return nil
	}

	actv = &domain.Activity{}
	actv.UID = reward.UID
	actv.AccountID = reward.AccountId
	actv.ID = reward.Event_Index
	actv.Date = tm
	actv.TxnType = domain.ActivityTypeReward
	actv.RcvAccountID = actv.AccountID
	actv.RcvSymbol = BASE_CURRENCY
	actv.RcvAmount = amount
	actv.SentAmount = value
	actv.SentSymbol = "USD"

	return actv

}

func (p PolkadotTransformer) createTransferActivity(ps core.PriceService, paccts []domain.Account, tsfr PolkadotTransfer) *domain.Activity {

	var actv *domain.Activity

	sentAcct := crypto.GetAccountFromAddress(paccts, tsfr.From)
	rcvAcct := crypto.GetAccountFromAddress(paccts, tsfr.To)
	amount, _ := utils.ConvertStringToDecimal(tsfr.Amount)
	tm := time.Unix(tsfr.Block_Timestamp, 10)

	actv = &domain.Activity{}
	actv.UID = tsfr.UID
	actv.AccountID = tsfr.AccountId
	actv.Date = tm
	actv.ID = tsfr.Hash
	actv.Hash = tsfr.Hash

	if sentAcct != nil && rcvAcct != nil {

		actv.TxnType = domain.ActivityTypeTransfer
		actv.SentAmount = amount
		actv.SentAccountID = sentAcct.ID
		actv.SentSymbol = tsfr.Asset_Symbol
		actv.RcvAccountID = rcvAcct.ID
		actv.RcvAmount = amount
		actv.RcvSymbol = tsfr.Asset_Symbol

	} else if sentAcct != nil {

		actv.TxnType = domain.ActivityTypeSend
		actv.SentAmount = amount
		actv.SentAccountID = sentAcct.ID
		actv.SentSymbol = tsfr.Asset_Symbol

	} else if rcvAcct != nil {

		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = rcvAcct.ID
		actv.RcvAmount = amount
		actv.RcvSymbol = tsfr.Asset_Symbol

	} else {
		return nil
	}

	return actv

}
