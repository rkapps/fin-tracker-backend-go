package ethereum

import (
	"fmt"

	"github.com/nanmu42/etherscan-api"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
	"github.com/shopspring/decimal"
)

var (
	ETH_BASE_CURRENCY = "ETH"
	TXN_DECIMALS      = decimal.NewFromInt(18)
)

func (s EthereumTransformer) createNormalTxnActivity(
	ps core.PriceService, eaccts []domain.Account, ntxn etherscan.NormalTx,
) *domain.Activity {

	// conver amount
	amount, _ := providers.ConvertStringToBaseDecimal(ntxn.Value.Int().String(), TXN_DECIMALS)
	if amount.IsZero() {
		s.logger.Error("createNormalTxnActivity", "Warning", "Value is zero")
		// return nil
	}
	if s.debug {
		s.logger.Info("createNormalTxnActivity", "Amount", amount)
	}

	faddrAcct := providers.GetAccountFromAddress(eaccts, ntxn.From)
	taddrAcct := providers.GetAccountFromAddress(eaccts, ntxn.To)

	if faddrAcct == nil && taddrAcct == nil {
		if s.debug {
			s.logger.Error("createNormalTxnActivity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = ntxn.Hash
	actv.Hash = ntxn.Hash
	actv.Date = ntxn.TimeStamp.Time()

	if s.debug {
		s.logger.Info("createNormalTxnActivity", "", fmt.Sprintf("From: %s", ntxn.From))
		s.logger.Info("createNormalTxnActivity", "", fmt.Sprintf("To  : %s", ntxn.To))
		s.logger.Info("createNormalTxnActivity", "", fmt.Sprintf("%s-%v", ETH_BASE_CURRENCY, amount))
	}

	if taddrAcct != nil && faddrAcct != nil {
		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeTransfer
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = ETH_BASE_CURRENCY
		actv.SentAmount = amount
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = ETH_BASE_CURRENCY
		actv.RcvAmount = amount

	} else if faddrAcct != nil {
		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeSend
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = ETH_BASE_CURRENCY
		actv.SentAmount = amount

		price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.RcvAmount = actv.SentAmount.Mul(price)
		actv.RcvSymbol = "USD"

	} else if taddrAcct != nil {
		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = ETH_BASE_CURRENCY
		actv.RcvAmount = amount

		price, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		actv.SentAmount = actv.RcvAmount.Mul(price)
		actv.SentSymbol = "USD"

	}

	gasprice := ntxn.GasPrice.Int().Int64()
	gasused := int64(ntxn.GasUsed)

	fee, _ := providers.ConvertInt64ToBaseDecimal(gasprice*gasused, TXN_DECIMALS)
	actv.Fee = fee
	actv.FeeCurrency = ETH_BASE_CURRENCY

	if s.debug {
		s.logger.Info("createNormalTxnActivity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func (s EthereumTransformer) createErc20Activity(
	ps core.PriceService, eaccts []domain.Account, tsfr etherscan.ERC20Transfer,
) *domain.Activity {

	faddrAcct := providers.GetAccountFromAddress(eaccts, tsfr.From)
	taddrAcct := providers.GetAccountFromAddress(eaccts, tsfr.To)

	if faddrAcct == nil && taddrAcct == nil {
		if s.debug {
			s.logger.Error("createErc20Activity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = tsfr.Hash
	actv.Hash = tsfr.Hash
	actv.Date = tsfr.TimeStamp.Time()

	// decExp := decimal.NewFromInt(int64(tsfr.TokenDecimal))
	// amount, _ := providers.ConvertStringToBaseDecimal(tsfr.Value.Int().String(), decExp)
	amount, _ := ConvertERC20Value(tsfr.Value.Int().String(), tsfr.TokenDecimal)

	if s.debug {
		s.logger.Info("createErc20Activity", "", fmt.Sprintf("From: %s", tsfr.From))
		s.logger.Info("createErc20Activity", "", fmt.Sprintf("To  : %s", tsfr.To))
		s.logger.Info("createErc20Activity", "", fmt.Sprintf("%s-%v", tsfr.TokenSymbol, amount))
	}

	if faddrAcct != nil && taddrAcct != nil {

		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeTransfer
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = tsfr.TokenSymbol
		actv.RcvAmount = amount
		actv.SentSymbol = tsfr.TokenSymbol
		actv.SentAmount = amount
	} else if faddrAcct != nil {

		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeSend
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = tsfr.TokenSymbol
		actv.SentAmount = amount
		price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.RcvAmount = actv.SentAmount.Mul(price)
		actv.RcvSymbol = "USD"

	} else if taddrAcct != nil {

		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = tsfr.TokenSymbol
		actv.RcvAmount = amount
		price, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		actv.SentAmount = actv.RcvAmount.Mul(price)
		actv.SentSymbol = "USD"

	}

	if s.debug {
		s.logger.Info("createErc20Activity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}
