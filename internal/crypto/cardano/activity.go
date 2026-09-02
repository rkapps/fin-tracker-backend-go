package cardano

import (
	"fmt"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

var (
	BASE_CURRENCY = "ADA"
	TXN_DECIMALS  = decimal.NewFromInt(6)
	ASSET_SYMBOL  = map[string]string{
		"lovelace": "ADA",
	}
)

type AccountActivity struct {
	addrAcctm    map[string]string
	withdrawalsm map[string]string
	assetsm      map[string]AccountAsset
	txn          Transaction
	logger       *logger.Logger
	debug        bool
}

type AccountAsset struct {
	unitm map[string]decimal.Decimal
}

func NewAccountActivity(addrAcctm map[string]string, withdrawalsm map[string]string, logger *logger.Logger, debug bool) AccountActivity {
	assetsm := make(map[string]AccountAsset)
	activity := AccountActivity{addrAcctm: addrAcctm, assetsm: assetsm, withdrawalsm: withdrawalsm, logger: logger, debug: debug}
	return activity

}

func (a *AccountActivity) add_transaction(txn Transaction) {

	a.txn = txn
	if a.debug {
		a.logger.Info("add_transaction", "Account", a.txn.AccountID, "Withdrawal", a.txn.WithdrawalAmount)
		a.logger.Info("add_transaction", "StakeCert", a.txn.StakeCertificates, "Metadata", a.txn.Metadata)
	}

	for _, entry := range txn.UTXO.Outputs {
		a.add_entry(true, entry.Address, entry.Amount)
	}
	for _, entry := range txn.UTXO.Inputs {
		a.add_entry(false, entry.Address, entry.Amount)
	}
}

func (a *AccountActivity) add_entry(receive bool, address string, txnAmounts []TransactionAmount) {
	// check if the address is in our address map. skip if it is not.

	acctId := a.addrAcctm[address]
	if len(acctId) == 0 {
		return
	}
	// log.Printf("Address: %s", address)
	if a.debug {
		a.logger.Info("add_entry", "Address", address)
	}
	aAsset := a.assetsm[acctId]
	if aAsset.unitm == nil {
		aAsset.unitm = make(map[string]decimal.Decimal)
	}

	// log.Println("reached")
	for _, txnAmount := range txnAmounts {
		amount, ok := aAsset.unitm[txnAmount.Unit]
		if !ok {
			aAsset.unitm[txnAmount.Unit] = decimal.Zero
		}
		qty, err := a.convertQty(txnAmount.Quantity)
		if !receive {
			qty = qty.Neg()
		}
		if a.debug {
			a.logger.Info("getActivity", "add_entry", txnAmount.Unit, "Qty", qty)
		}

		if err == nil {
			amount = amount.Add(qty)
			aAsset.unitm[txnAmount.Unit] = amount
		}
	}
	a.assetsm[acctId] = aAsset
}

func (a *AccountActivity) convertQty(quantity string) (decimal.Decimal, error) {
	return crypto.ConvertStringToBaseDecimal(quantity, TXN_DECIMALS)
}

func (a *AccountActivity) getActivity() ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}

	if a.debug {
		a.logger.Debug("getActivity", "assets", len(a.assetsm))
	}

	if len(a.assetsm) == 0 {
		return nil, fmt.Errorf("no assets found")
	}
	if len(a.assetsm) == 2 {
		return a.getTransferActivity()
	}
	if len(a.assetsm) != 1 {
		return nil, fmt.Errorf("should never happen")
	}

	// there is only 1 account - send or a receive or trade
	for k, v := range a.assetsm {
		if a.debug {
			a.logger.Info("getActivity", "account", k)
		}

		// loop through the units ( there could be multiple assets )
		for asset, amount := range v.unitm {
			if a.debug {
				a.logger.Info("getActivity", "unit", fmt.Sprintf("%s-%v", asset, amount))
			}
			symbol := ASSET_SYMBOL[asset]
			if len(symbol) == 0 {
				// return nil, fmt.Errorf("Asset '%s' map not found", asset)
				continue
			}

			wamount := decimal.Zero
			// if there is a withdrawal amount, add negative amount
			if strings.Compare(symbol, "ADA") == 0 {
				if a.withdrawalsm != nil {
					wamountstr := a.withdrawalsm[a.txn.TxHash]
					wamount, _ = a.convertQty(wamountstr)

					if a.debug {
						a.logger.Info("getActivity", "withdrawal", wamount)
					}
				}
			}
			if !wamount.IsZero() {
				amount = amount.Sub(wamount)
			}
			if a.debug {
				a.logger.Info("getActivity", "unit", fmt.Sprintf("%s-%v", asset, amount))
			}

			// check txntype from metadata
			mnotes, mtxnType := a.getTxnTypeFromMetadata(a.txn.Metadata)

			actv := &domain.Activity{}
			actv.UID = a.txn.UID
			// actv.AccountID = a.txn.AccountID
			actv.AccountID = k
			if amount.IsPositive() {
				actv.TxnType = domain.ActivityTypeReceive
				actv.RcvAccountID = k
				actv.RcvSymbol = symbol
				actv.RcvAmount = amount
				if len(mtxnType) > 0 {
					actv.TxnType = domain.ActivityType(mtxnType)
					actv.Notes = mnotes
				}

			} else {

				actv.TxnType = domain.ActivityTypeSend
				actv.SentAccountID = k
				actv.SentSymbol = symbol
				actv.SentAmount = amount.Neg()

				// add fees on to the first activity
				if len(actvs) == 0 {
					fee, _ := a.convertQty(a.txn.Fees)
					actv.Fee = fee
					actv.FeeCurrency = BASE_CURRENCY
					actv.SentAmount = actv.SentAmount.Sub(actv.Fee)
				}

				if a.debug {
					a.logger.Info("getActivity", "Fee", actv.Fee, "SentAmount", actv.SentAmount)
				}

				if len(a.txn.Delegations) > 0 {
					actv.TxnType = domain.ActivityTypeDelegation
				}

				// cert should override the delegation
				for _, cert := range a.txn.StakeCertificates {
					if cert.Registration {
						actv.TxnType = domain.ActivityTypeStakeFee
					}
				}

				// default to fee if amount is zero
				if actv.SentAmount.IsZero() {
					actv.TxnType = domain.ActivityTypeFee
				}

				if a.debug {
					a.logger.Info("getActivity", "TxnType", actv.TxnType, "unit", fmt.Sprintf("%s-%v", actv.SentSymbol, actv.SentAmount))
				}

			}

			actvs = append(actvs, actv)
			// return actv, nil
		}
	}

	return actvs, nil
}

func (a *AccountActivity) getTxnTypeFromMetadata(metadata []TransactionMetadata) (string, string) {
	for _, entry := range metadata {
		if a.debug {
			// log.Printf("%v", entry.JsonMetadata)
		}

		metam := entry.JsonMetadata
		for _, v := range metam {
			switch entry := v.(type) {
			// case primitive.A:
			// 	{
			// notes := fmt.Sprintf("%s", entry)
			// if a.debug {
			// 	a.logger.Info("getTxnFromMetadata", "notes", notes)
			// }
			// }
			default:
				notes := fmt.Sprintf("%s", entry)
				txnType := ""
				if strings.Contains(notes, "Voter rewards") {
					txnType = string(domain.ActivityTypeReward)
				}
				if a.debug {
					a.logger.Info("getTxnFromMetadata", "notes", notes, "TxnType", txnType)
				}
				return notes, txnType
			}
		}
	}
	return "", ""
}

func (a *AccountActivity) getTransferActivity() ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}

	var index = 0
	var firstAccountId, secondAccountId string
	var firstAccountAsset, secondAccountAsset AccountAsset
	for k, v := range a.assetsm {
		// a.logger.Info("getActivity", "account", k, "v", v)
		if index == 0 {
			firstAccountId = k
			firstAccountAsset = v
		} else {
			secondAccountId = k
			secondAccountAsset = v
		}
		index++
	}

	for asset, amount := range firstAccountAsset.unitm {
		tamount := secondAccountAsset.unitm[asset]

		symbol := ASSET_SYMBOL[asset]
		if len(symbol) == 0 {
			// return nil, fmt.Errorf("Asset '%s' map not found", asset)
			continue
		}
		if a.debug {
			a.logger.Info("getTransferActivity", "Account", fmt.Sprintf("%s-%s", firstAccountId, secondAccountId), symbol, fmt.Sprintf("%v-%v", amount, tamount))
		}

		// if there is a withdrawal amount, add negative amount
		if strings.Compare(symbol, "ADA") == 0 {
			wamount := decimal.Zero
			if a.withdrawalsm != nil {
				wamountstr := a.withdrawalsm[a.txn.TxHash]
				wamount, _ = a.convertQty(wamountstr)
			}
			if !wamount.IsZero() {
				// if tamount.IsNegative() {
				// 	tamount = tamount.Sub(wamount)
				// } else if amount.IsNegative() {
				// }
				tamount = tamount.Sub(wamount)

			}
		}

		actv := &domain.Activity{}
		actv.UID = a.txn.UID
		actv.AccountID = a.txn.AccountID
		actv.TxnType = domain.ActivityTypeTransfer
		fee, _ := a.convertQty(a.txn.Fees)
		actv.Fee = fee
		actv.FeeCurrency = BASE_CURRENCY

		if amount.IsNegative() {
			actv.AccountID = firstAccountId
			actv.SentAccountID = firstAccountId
			actv.SentSymbol = symbol
			actv.SentAmount = amount.Abs()
			actv.SentAmount = actv.SentAmount.Sub(actv.Fee)
			actv.RcvAccountID = secondAccountId
			actv.RcvSymbol = symbol
			actv.RcvAmount = tamount
		} else {
			actv.AccountID = secondAccountId
			actv.SentAccountID = secondAccountId
			actv.RcvAccountID = firstAccountId
			actv.SentSymbol = symbol
			actv.SentAmount = tamount.Abs()
			actv.SentAmount = actv.SentAmount.Sub(actv.Fee)
			actv.RcvSymbol = symbol
			actv.RcvAmount = amount
		}
		if a.debug {
			a.logger.Info("getTransferActivity", "Account", a.txn.AccountID)
			a.logger.Info("getTransferActivity", "SentAccount", fmt.Sprintf("%s-%s-%v", actv.SentAccountID, actv.SentSymbol, actv.SentAmount))
			a.logger.Info("getTransferActivity", "RcvAccount", fmt.Sprintf("%v-%v", actv.RcvAccountID, actv.RcvAmount))
		}
		actvs = append(actvs, actv)

	}

	return actvs, nil
}
