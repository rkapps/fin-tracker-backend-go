package ethereum

import (
	"fmt"

	"github.com/nanmu42/etherscan-api"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	BASE_CURRENCY         = "USD"
	ETH_BASE_CURRENCY     = "ETH"
	ETH_WRAP_CURRENCY     = "WETH"
	POLYGON_BASE_CURRENCY = "POL"
	POLYGON_WRAP_CURRENCY = "WPOL"
	TXN_DECIMALS          = decimal.NewFromInt(18)
)

func createNormalTxnActivity(
	ps core.PriceService, eaccts []domain.Account, ntxn etherscan.NormalTx, baseSymbol string,
	sel Selector, logger *logger.Logger, debug bool,
) *domain.Activity {

	// conver amount
	amount, _ := crypto.ConvertStringToBaseDecimal(ntxn.Value.Int().String(), TXN_DECIMALS)
	if amount.IsZero() {
		// s.logger.Warn("createNormalTxnActivity", "Warning", "Value is zero")
		// DONT SKIP ANY TRANSACTIONS BECAUSE OF THE FEE
		// return nil
	}

	faddrAcct := crypto.GetAccountFromAddress(eaccts, ntxn.From)
	taddrAcct := crypto.GetAccountFromAddress(eaccts, ntxn.To)

	if faddrAcct == nil && taddrAcct == nil {
		if debug {
			logger.Error("createNormalTxnActivity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = ntxn.Hash
	actv.Hash = ntxn.Hash
	actv.Date = ntxn.TimeStamp.Time()

	if debug {
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("From: %s", ntxn.From))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("To  : %s", ntxn.To))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("%s-%v", baseSymbol, amount))
	}

	if taddrAcct != nil && faddrAcct != nil {
		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeTransfer
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = baseSymbol
		actv.SentAmount = amount
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = baseSymbol
		actv.RcvAmount = amount

	} else if faddrAcct != nil {
		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeSend
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = baseSymbol
		actv.SentAmount = amount

		price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.RcvAmount = actv.SentAmount.Mul(price)
		actv.RcvSymbol = baseSymbol

	} else if taddrAcct != nil {
		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = baseSymbol
		actv.RcvAmount = amount

		price, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		actv.SentAmount = actv.RcvAmount.Mul(price)
		actv.SentSymbol = BASE_CURRENCY

	}

	gasprice := ntxn.GasPrice.Int().Int64()
	gasused := int64(ntxn.GasUsed)

	fee, _ := crypto.ConvertInt64ToBaseDecimal(gasprice*gasused, TXN_DECIMALS)
	actv.Fee = fee
	actv.FeeCurrency = baseSymbol

	// update from selector
	if len(sel.TxnType) > 0 {
		actv.TxnType = sel.TxnType
		actv.Notes = sel.Label
	}

	if debug {
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func createErc20Activity(
	ps core.PriceService, eaccts []domain.Account, tsfr etherscan.ERC20Transfer, baseSymbol string,
	sel Selector, logger *logger.Logger, debug bool,
) *domain.Activity {

	// spam transfers have blank token symbols
	if len(tsfr.TokenSymbol) == 0 {
		return nil
	}

	faddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr.From)
	taddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr.To)

	if faddrAcct == nil && taddrAcct == nil {
		if debug {
			logger.Error("createErc20Activity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = tsfr.Hash
	actv.Hash = tsfr.Hash
	actv.Date = tsfr.TimeStamp.Time()

	fee, _ := crypto.ConvertInt64ToBaseDecimal(int64(tsfr.GasPrice.Int().Int64())*int64(tsfr.GasUsed), TXN_DECIMALS)
	actv.Fee = fee.Round(int32(crypto.MAX_DECIMALS))
	actv.FeeCurrency = baseSymbol
	if debug {
		logger.Info("createErc20Activity", "", fmt.Sprintf("Fee %v-%v", tsfr.GasPrice, tsfr.GasUsed))
		logger.Info("createErc20Activity", "", fmt.Sprintf("Fee %s-%v", actv.FeeCurrency, actv.Fee))
	}
	// decExp := decimal.NewFromInt(int64(tsfr.TokenDecimal))
	// amount, _ := crypto.ConvertStringToBaseDecimal(tsfr.Value.Int().String(), decExp)
	amount, _ := ConvertERC20Value(tsfr.Value.Int().String(), tsfr.TokenDecimal)

	if debug {
		logger.Info("createErc20Activity", "", fmt.Sprintf("From: %s", tsfr.From))
		logger.Info("createErc20Activity", "", fmt.Sprintf("To  : %s", tsfr.To))
		logger.Info("createErc20Activity", "", fmt.Sprintf("%s-%v", tsfr.TokenSymbol, amount))
		logger.Info("createErc20Activity", "", fmt.Sprintf("Fee %s-%v", actv.FeeCurrency, actv.Fee))
	}

	if faddrAcct != nil && taddrAcct != nil {

		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeTransfer
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = tsfr.TokenSymbol
		actv.SentAmount = amount

		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = tsfr.TokenSymbol
		actv.RcvAmount = amount

	} else if faddrAcct != nil {

		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.TxnType = domain.ActivityTypeSend
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = tsfr.TokenSymbol
		actv.SentAmount = amount
		price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		actv.RcvAmount = actv.SentAmount.Mul(price)
		actv.RcvSymbol = BASE_CURRENCY

	} else if taddrAcct != nil {

		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = tsfr.TokenSymbol
		actv.RcvAmount = amount
		price, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		actv.SentAmount = actv.RcvAmount.Mul(price)
		actv.SentSymbol = BASE_CURRENCY

	}

	// update from selector
	if len(sel.TxnType) > 0 {
		actv.TxnType = sel.TxnType
		actv.Notes = sel.Label
	}

	if debug {
		logger.Info("createErc20Activity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func createErc20TradeActivity(
	ps core.PriceService, eaccts []domain.Account, tsfr etherscan.ERC20Transfer, tsfr1 etherscan.ERC20Transfer,
	sel Selector, logger *logger.Logger, debug bool,
) *domain.Activity {

	faddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr.From)
	taddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr1.To)

	// decExp := decimal.NewFromInt(int64(tsfr.TokenDecimal))
	// amount, _ := crypto.ConvertStringToBaseDecimal(tsfr.Value.Int().String(), decExp)
	amount1, _ := ConvertERC20Value(tsfr.Value.Int().String(), tsfr.TokenDecimal)
	amount2, _ := ConvertERC20Value(tsfr1.Value.Int().String(), tsfr1.TokenDecimal)

	if debug {
		logger.Info("createErc20TradeActivity", "", fmt.Sprintf("From: %s", tsfr.From))
		logger.Info("createErc20TradeActivity", "", fmt.Sprintf("To  : %s", tsfr.To))
		logger.Info("createErc20TradeActivity", "", fmt.Sprintf("%s-%v", tsfr.TokenSymbol, amount1))
	}

	if faddrAcct == nil && taddrAcct == nil {
		if debug {
			logger.Warn("createErc20Activity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = tsfr.Hash
	actv.Hash = tsfr.Hash
	actv.Date = tsfr.TimeStamp.Time()

	actv.UID = faddrAcct.UID
	actv.AccountID = faddrAcct.ID
	actv.TxnType = domain.ActivityTypeTrade
	actv.SentSymbol = tsfr.TokenSymbol
	actv.SentAmount = amount1

	price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
	actv.SentPrice = price

	actv.RcvAccountID = taddrAcct.ID
	actv.RcvSymbol = tsfr1.TokenSymbol
	actv.RcvAmount = amount2
	price2, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
	actv.RcvPrice = price2

	actv.Notes = sel.Label

	if debug {
		logger.Info("createErc20Activity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func createErc20WrapActivity(
	ps core.PriceService, eaccts []domain.Account, tsfr etherscan.ERC20Transfer, wrapSymbol string,
	sel Selector, logger *logger.Logger, debug bool,
) *domain.Activity {

	faddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr.From)
	taddrAcct := crypto.GetAccountFromAddress(eaccts, tsfr.To)

	if faddrAcct == nil && taddrAcct == nil {
		if debug {
			logger.Error("createErc20WrapActivity", "Error", "should never happen")
		}
		return nil
	}

	actv := &domain.Activity{}
	actv.ID = tsfr.Hash
	actv.Hash = tsfr.Hash
	actv.Date = tsfr.TimeStamp.Time()

	// decExp := decimal.NewFromInt(int64(tsfr.TokenDecimal))
	// amount, _ := crypto.ConvertStringToBaseDecimal(tsfr.Value.Int().String(), decExp)
	amount, _ := ConvertERC20Value(tsfr.Value.Int().String(), tsfr.TokenDecimal)

	if debug {
		logger.Info("createErc20WrapActivity", "", fmt.Sprintf("From: %s", tsfr.From))
		logger.Info("createErc20WrapActivity", "", fmt.Sprintf("To  : %s", tsfr.To))
		logger.Info("createErc20WrapActivity", "", fmt.Sprintf("%s-%v", tsfr.TokenSymbol, amount))
	}

	if faddrAcct != nil {

		actv.UID = faddrAcct.UID
		actv.AccountID = faddrAcct.ID
		actv.SentAccountID = faddrAcct.ID
		actv.SentSymbol = tsfr.TokenSymbol
		actv.SentAmount = amount

		// get the sent and receive prices
		sprice, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
		rprice, _ := ps.GetCryptoPrice(wrapSymbol, actv.Date)
		actv.SentPrice = sprice

		actv.RcvAccountID = faddrAcct.ID
		actv.RcvSymbol = wrapSymbol
		if rprice.IsPositive() {
			actv.RcvAmount = actv.SentAmount.Mul(sprice).Div(rprice)
		}

	} else if taddrAcct != nil {

		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.SentAccountID = taddrAcct.ID
		actv.SentSymbol = wrapSymbol
		actv.SentAmount = amount

		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = tsfr.TokenSymbol
		sprice, _ := ps.GetCryptoPrice(wrapSymbol, actv.Date)
		actv.SentPrice = sprice

		rprice, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		if rprice.IsPositive() {
			actv.RcvAmount = actv.SentAmount.Mul(sprice).Div(rprice)
		}

	}

	actv.TxnType = domain.ActivityTypeTrade
	if len(sel.Label) > 0 {
		actv.Notes = sel.Label
	}

	if debug {
		logger.Info("createErc20Activity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func createErcMultiActivity(
	ps core.PriceService, eaccts []domain.Account, tsfrs []etherscan.ERC20Transfer, baseSymbol string,
	logger *logger.Logger, debug bool,
) []*domain.Activity {

	actvs := []*domain.Activity{}
	for i, tsfr := range tsfrs {
		actv := createErc20Activity(ps, eaccts, tsfr, baseSymbol, Selector{}, logger, debug)
		actv.ID = fmt.Sprintf("%s-%d", tsfr.Hash, i)
		if actv != nil {
			switch actv.TxnType {
			case domain.ActivityTypeReceive:
				actv.TxnType = domain.ActivityTypeExitLiquidity
			case domain.ActivityTypeSend:
				actv.TxnType = domain.ActivityTypeAddLiquidity
			}
			actvs = append(actvs, actv)
		}
	}
	return actvs
}

func createNormalWrapActivity(
	ps core.PriceService, eaccts []domain.Account, ntxn etherscan.NormalTx, baseSymbol string, wrapSymbol string,
	logger *logger.Logger, debug bool,
) *domain.Activity {

	// conver amount
	amount, _ := crypto.ConvertStringToBaseDecimal(ntxn.Value.Int().String(), TXN_DECIMALS)
	if amount.IsZero() {
		// s.logger.Warn("createNormalTxnActivity", "Warning", "Value is zero")
		// DONT SKIP ANY TRANSACTIONS BECAUSE OF THE FEE
		// return nil
	}

	faddrAcct := crypto.GetAccountFromAddress(eaccts, ntxn.From)

	actv := &domain.Activity{}
	actv.ID = ntxn.Hash
	actv.Hash = ntxn.Hash
	actv.Date = ntxn.TimeStamp.Time()

	if debug {
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("From: %s", ntxn.From))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("To  : %s", ntxn.To))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("%s-%v", baseSymbol, amount))
	}

	actv.UID = faddrAcct.UID
	actv.AccountID = faddrAcct.ID
	actv.TxnType = domain.ActivityTypeTrade
	actv.SentAccountID = faddrAcct.ID
	actv.SentSymbol = baseSymbol
	actv.SentAmount = amount
	price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
	actv.SentPrice = price

	actv.RcvAccountID = faddrAcct.ID
	actv.RcvSymbol = wrapSymbol
	actv.RcvAmount = amount

	gasprice := ntxn.GasPrice.Int().Int64()
	gasused := int64(ntxn.GasUsed)

	fee, _ := crypto.ConvertInt64ToBaseDecimal(gasprice*gasused, TXN_DECIMALS)
	actv.Fee = fee
	actv.FeeCurrency = baseSymbol

	if debug {
		logger.Info("createNormalWrapActivity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}

func createInternalTxnActivity(
	ps core.PriceService, eaccts []domain.Account, itxn etherscan.InternalTx,
	logger *logger.Logger, debug bool,
) *domain.Activity {

	// conver amount
	amount, _ := crypto.ConvertStringToBaseDecimal(itxn.Value.Int().String(), TXN_DECIMALS)
	if amount.IsZero() {
		return nil
	}

	taddrAcct := crypto.GetAccountFromAddress(eaccts, itxn.To)

	actv := &domain.Activity{}
	actv.ID = itxn.Hash
	actv.Hash = itxn.Hash
	actv.Date = itxn.TimeStamp.Time()

	if debug {
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("From: %s", itxn.From))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("To  : %s", itxn.To))
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("%s-%v", ETH_BASE_CURRENCY, amount))
	}

	if taddrAcct != nil {
		actv.UID = taddrAcct.UID
		actv.AccountID = taddrAcct.ID
		actv.TxnType = domain.ActivityTypeReceive
		actv.RcvAccountID = taddrAcct.ID
		actv.RcvSymbol = ETH_BASE_CURRENCY
		actv.RcvAmount = amount

		price, _ := ps.GetCryptoPrice(actv.RcvSymbol, actv.Date)
		actv.SentAmount = actv.RcvAmount.Mul(price)
		actv.SentSymbol = BASE_CURRENCY

	} else {
		return nil
	}

	gasprice := int64(itxn.Gas)
	gasused := int64(itxn.GasUsed)

	fee, _ := crypto.ConvertInt64ToBaseDecimal(gasprice*gasused, TXN_DECIMALS)
	actv.Fee = fee
	actv.FeeCurrency = ETH_BASE_CURRENCY

	if debug {
		logger.Info("createNormalTxnActivity", "", fmt.Sprintf("---%s---", actv.TxnType))
	}

	return actv
}
