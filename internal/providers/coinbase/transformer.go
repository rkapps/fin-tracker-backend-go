package coinbase

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

type CoinbaseAccountTransformer struct {
	logger *logger.Logger
}

func NewCoinbaseAccountTransformer(logConfig *logger.Config) CoinbaseAccountTransformer {
	plog := logConfig.For("refresher.coinbase")
	return CoinbaseAccountTransformer{plog}
}

func (r CoinbaseAccountTransformer) Name() string {
	return "coinbase"
}

func (r CoinbaseAccountTransformer) Transform(ctx context.Context, ps core.PriceService,
	gaccts []*domain.Account,
	acreds []domain.AccountWithCredential,
	globalRaws []domain.RawItem,
	rawsm map[string][]domain.RawItem,
) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}

	// // get grouped raw items by account
	// rawsm := providers.GroupRawItems(raws)

	for _, acred := range acreds {
		raws := rawsm[acred.Account.ID]
		if len(raws) == 0 {
			continue
		}
		r.logger.Debug("Transform", "Items", len(raws))

		// get payments, accounts and transactions
		pms, caccts, txns := r.marshalData(raws)
		r.logger.Debug("Transform", "Payments", len(pms))
		r.logger.Debug("Transform", "Transactions", len(txns))

		// add payments to map by payment names
		// pmsm := make(map[string]CnbPaymentMethod)
		// for _, pm := range pms {
		// 	pmsm[pm.Payment_Method_Name] = pm
		// }

		for _, cacct := range caccts {
			// log.Println(cacct)
			r.logger.Debug("Transform", "Account", acred.Account.ID, "CnbAccount", cacct.Balance.Currency)
			if cacct.Balance.Currency == "ETH" ||
				cacct.Balance.Currency == "BCH" ||
				cacct.Balance.Currency == "USD" ||
				cacct.Balance.Currency == "LTC" {
				// log.Printf("%s-%s", cacct.Id, cacct.Balance.Currency)
			}
		}
		sort.Slice(txns, func(a, b int) bool {
			date1 := utils.DateTimeFromString(txns[a].Created_At)
			date2 := utils.DateTimeFromString(txns[b].Created_At)
			return date1.Before(*date2)
		})

		stkActvsm := make(map[int]*domain.Activity)

		loopBreak := false
		for i, txn := range txns {

			if strings.Compare(txn.Id, "e12f748a-42a9-5a2e-b46e-35753dc5a49f") == 0 {
				r.logger.Debug("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id), "CreatedAt", txn.Created_At)
				r.logger.Debug("Transform", "Status", txn.Status)
			}

			if strings.Compare(txn.Status, "canceled") == 0 || strings.Compare(txn.Status, "cancelled") == 0 {
				continue
			}
			// // possible duplicate transaction or internal coinbase transaction
			if (strings.Compare(txn.Type, "sell") == 0 || strings.Compare(txn.Type, "buy") == 0 || strings.Compare(txn.Type, "advanced_trade_fill") == 0) &&
				strings.Compare(txn.Amount.Currency, txn.Native_Amount.Currency) == 0 {
				continue
			}

			actv := &domain.Activity{}
			actv.UID = acred.Account.UID
			actv.AccountID = acred.Account.ID
			actv.ID = txn.Id
			actv.Date = *utils.DateTimeFromString(txn.Created_At)

			amount, err := utils.ConvertStringToDecimal(txn.Amount.Amount)
			symbol := txn.Amount.Currency
			nsymbol := txn.Native_Amount.Currency
			namount, _ := utils.ConvertStringToDecimal(txn.Native_Amount.Amount)

			if err != nil {
				r.logger.Error("Transform", "Transaction", txn.Id, "Error", err)
				continue
			}
			switch txn.Type {
			case "buy":

				buy := txn.Buy
				total, _ := utils.ConvertStringToDecimal(buy.Total.Amount)
				subtotal, _ := utils.ConvertStringToDecimal(buy.Subtotal.Amount)

				actv.TxnType = domain.ActivityTypeBuy
				actv.RcvAccountID = acred.Account.ID
				actv.RcvAmount = amount
				actv.RcvSymbol = symbol
				actv.SentAmount = subtotal
				actv.SentSymbol = nsymbol
				actv.SentAccountID = acred.Account.ID
				actv.RcvPrice = actv.SentAmount.Div(actv.RcvAmount)
				actv.Fee = total.Sub(subtotal)
				actv.FeeCurrency = buy.Fee.Currency

				if len(txn.Buy.Payment_Method_Name) > 0 {

					bacct := core.GetBankAccount(gaccts, txn.Buy.Payment_Method_Name)
					r.logger.Debug("Transform", "Payment", txn.Buy.Payment_Method_Name, "Account", bacct)
					if bacct != nil {
						actv.SentAccountID = bacct.ID
						// actv.SentAccount = bacct.ID
					}
				}

			case "sell":
				sell := txn.Sell
				total, _ := utils.ConvertStringToDecimal(sell.Total.Amount)
				subtotal, _ := utils.ConvertStringToDecimal(sell.Subtotal.Amount)

				actv.TxnType = domain.ActivityTypeSell
				actv.SentAccountID = acred.Account.ID
				actv.SentAmount = amount.Abs()
				actv.SentSymbol = symbol
				actv.RcvAmount = total
				actv.RcvSymbol = nsymbol
				actv.RcvAccountID = acred.Account.ID
				actv.SentPrice = actv.RcvAmount.Div(actv.SentAmount)

				actv.Fee = subtotal.Sub(total)
				actv.FeeCurrency = sell.Fee.Currency

			case "send":
				// log.Println(txn)
				tamount, _ := utils.ConvertStringToDecimal(txn.Network.Transaction_Fee.Amount)

				if len(txn.To.Address) == 0 {
					actv.TxnType = domain.ActivityTypeReceive
					actv.RcvAccountID = acred.Account.ID
					actv.RcvSymbol = symbol
					actv.RcvAmount = amount
					actv.SentSymbol = nsymbol
					actv.SentAmount = namount
					actv.Notes = txn.From.Name

					if strings.Compare(txn.From.Name, "Coinbase") == 0 ||
						strings.Compare(txn.From.Name, "Coinbase Earn") == 0 ||
						strings.Compare(txn.From.Name, "Coinbase Rewards") == 0 {

						// log.Println(txn.Amount)
						// log.Println(txn.Native_Amount)
						actv.TxnType = domain.ActivityTypeReward
						actv.Value = namount
					}

				} else {

					actv.TxnType = domain.ActivityTypeSend
					actv.SentAccountID = acred.Account.ID
					actv.SentAmount = amount.Abs().Sub(tamount.Abs())
					actv.SentSymbol = symbol
					actv.RcvAccount = txn.To.Address
					actv.RcvAmount = namount.Abs()
					actv.Fee = tamount

					actv.Notes = fmt.Sprintf("To:%s", txn.To.Address)

				}

			case "fiat_withdrawal":
				actv.TxnType = domain.ActivityTypeWithdraw
				actv.SentAccountID = acred.Account.ID
				actv.SentAmount = amount.Abs()
				actv.SentSymbol = symbol
				// log.Println(txn)
				racct := core.GetFirstBankAccount(gaccts)
				if racct != nil {
					actv.RcvAccountID = racct.ID
					actv.RcvSymbol = symbol
					actv.RcvAmount = actv.SentAmount
				} else {
					r.logger.Info("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id), "Error", "Fiat withdrawal--")
					continue
				}

			case "fiat_deposit":
				actv.TxnType = domain.ActivityTypeDeposit
				actv.RcvAccountID = acred.Account.ID
				actv.RcvAmount = amount.Abs()
				actv.RcvSymbol = symbol
				// log.Println(txn)
				sacct := core.GetFirstBankAccount(gaccts)
				if sacct != nil {
					actv.SentAccountID = sacct.ID
					actv.SentSymbol = symbol
					actv.SentAmount = actv.RcvAmount
				} else {
					r.logger.Info("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id), "Error", "Fiat withdrawal--")
					continue
				}

			case "pro_deposit", "exchange_deposit":
				// log.Println(txn)
				actv.TxnType = domain.ActivityTypeTransfer
				actv.SentAccountID = acred.Account.ID
				actv.SentAmount = amount.Abs()
				actv.SentSymbol = symbol

				racct := core.GetCoinbaseProAccount(gaccts)
				if racct == nil {
					r.logger.Error("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id))
				} else {
					actv.RcvAccountID = racct.ID
					actv.RcvAmount = actv.SentAmount
					actv.RcvSymbol = actv.SentSymbol
				}
			case "pro_withdrawal", "exchange_withdrawal":
				// log.Println(amount)
				actv.TxnType = domain.ActivityTypeTransfer
				actv.RcvAccountID = acred.Account.ID
				actv.RcvAmount = amount.Abs()
				actv.RcvSymbol = symbol

				sacct := core.GetCoinbaseProAccount(gaccts)
				if sacct == nil {
					r.logger.Error("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id))
				} else {
					actv.SentAccountID = sacct.ID
					actv.SentAmount = actv.RcvAmount
					actv.SentSymbol = actv.RcvSymbol
					actv.AccountID = sacct.ID
				}
			case "staking_reward", "interest":
				if strings.Compare(txn.Type, "staking_reward") == 0 {
					actv.TxnType = domain.ActivityTypeReward
				} else {
					actv.TxnType = domain.ActivityTypeInterest
				}
				actv.RcvAccountID = acred.Account.ID
				actv.RcvAmount = amount.Abs()
				actv.RcvSymbol = symbol
				actv.SentSymbol = nsymbol
				actv.SentAmount = namount

			case "trade":
				actv.RcvAccountID = actv.AccountID
				actv.SentAccountID = actv.AccountID

				if amount.IsPositive() {
					actv.TxnType = domain.ActivityTypeBuy
					actv.RcvSymbol = symbol
					actv.RcvAmount = amount
					actv.SentSymbol = nsymbol
					actv.SentAmount = namount

				} else {

					actv.TxnType = domain.ActivityTypeSell
					actv.RcvSymbol = nsymbol
					actv.RcvAmount = namount.Neg()
					actv.SentSymbol = symbol
					actv.SentAmount = amount.Neg()
				}
			case "advanced_trade_fill":
				atf := txn.Advanced_trade_fill
				actv.SentAccountID = actv.AccountID
				actv.RcvAccountID = actv.AccountID

				amount = amount.Abs()
				fprice, _ := utils.ConvertStringToDecimal(atf.Fill_Price)
				prds := strings.Split(atf.Product_Id, "-")

				actv.TxnType = domain.ActivityType(atf.Order_Side)
				if strings.Compare(string(actv.TxnType), string(domain.ActivityTypeBuy)) == 0 {
					actv.SentSymbol = prds[1]
					actv.SentAmount = amount.Mul(fprice)
					actv.RcvSymbol = prds[0]
					actv.RcvAmount = amount
				} else {
					actv.SentSymbol = prds[0]
					actv.SentAmount = amount
					actv.RcvSymbol = prds[1]
					actv.RcvAmount = amount.Mul(fprice)
				}
				// log.Println(atf.Commission)
				// fee, err := utils.ConvertStringToDecimal(atf.Commission)
				// if err == nil {
				// log.Println(fee)
				// actv.Fee = fee
				// actv.FeeCurrency = "USD"
				// }
			case "staking_transfer":
				r.logger.Debug("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id), "Date", txn.Created_At)
				r.logger.Debug("Transform", "Amount", txn.Amount, "NativeAmount", txn.Native_Amount)

				duration := actv.Date.Sub(time.Unix(0, 0))
				ms := int(duration.Minutes())
				sactv := stkActvsm[ms]
				if sactv != nil {
					actv.TxnType = domain.ActivityTypeTransfer
					actv.SentAccountID = actv.AccountID
					actv.RcvAccountID = actv.AccountID
					if amount.IsPositive() {
						actv.RcvSymbol = symbol
						actv.RcvAmount = sactv.SentAmount
						actv.SentSymbol = sactv.SentSymbol
						actv.SentAmount = sactv.SentAmount
					} else {
						actv.RcvSymbol = sactv.RcvSymbol
						actv.RcvAmount = sactv.RcvAmount
						actv.SentSymbol = symbol
						actv.SentAmount = sactv.RcvAmount
					}

				} else {

					if amount.IsPositive() {
						actv.RcvSymbol = symbol
						actv.RcvAmount = amount
						actv.SentSymbol = nsymbol
						actv.SentAmount = namount
					} else {
						actv.SentSymbol = symbol
						actv.SentAmount = amount.Neg()
						actv.RcvSymbol = nsymbol
						actv.RcvAmount = namount.Neg()
					}
					stkActvsm[ms] = actv
					continue
				}
			// case "unstaking_transfer":
			// 	// log.Println(txn)
			// 	continue

			case "retail_eth2_deprecation":
				// log.Println(txn.Amount)

				actv.TxnType = domain.ActivityTypeTransfer
				actv.SentAccountID = actv.AccountID
				actv.SentSymbol = symbol
				actv.SentAmount = amount.Neg()
				actv.RcvAccountID = actv.AccountID
				actv.RcvSymbol = "ETH"
				actv.RcvAmount = actv.SentAmount

			// case "tx":
			// 	continue
			default:
				r.logger.Info("Transform", "Transaction", fmt.Sprintf("%s-%s", txn.Type, txn.Id), "Date", txn.Created_At)
				// log.Println(txn)
				loopBreak = true
				continue
			}
			actvs = append(actvs, actv)
			if loopBreak {
				// log.Println(i)
				// break
			}
			if i > 11 {
				// break
			}
		}
	}
	r.logger.Info("Transform", "Provider", r.Name(), "Actvs", len(actvs))
	return actvs, nil
}

func (r CoinbaseAccountTransformer) marshalData(raws []domain.RawItem) (
	[]CnbPaymentMethod,
	[]CnbAccount,
	[]CnbTransaction,
) {

	var pms []CnbPaymentMethod
	var caccts []CnbAccount
	var txns []CnbTransaction

	for _, raw := range raws {
		r.logger.Debug("Refresh", "Id", raw.ExternalID, "Stream", raw.Stream)
		switch raw.Stream {
		case "payments":
			pm, _ := r.getPaymentMethod(raw)
			pms = append(pms, pm)
		case "accounts":
			cacct, _ := r.getAccount(raw)
			caccts = append(caccts, cacct)
		case "transactions":
			txn, _ := r.getTransaction(raw)
			txns = append(txns, txn)
		}
	}

	return pms, caccts, txns
}

func (r CoinbaseAccountTransformer) getPaymentMethod(raw domain.RawItem) (CnbPaymentMethod, error) {
	var pm CnbPaymentMethod
	bytes, err := json.Marshal(raw.Payload)
	if err == nil {
		err = json.Unmarshal(bytes, &pm)
	}
	return pm, err
}

func (r CoinbaseAccountTransformer) getAccount(raw domain.RawItem) (CnbAccount, error) {
	var a CnbAccount
	bytes, err := json.Marshal(raw.Payload)
	// log.Printf("%s", bytes)
	if err == nil {
		err = json.Unmarshal(bytes, &a)
	}
	return a, err
}

func (r CoinbaseAccountTransformer) getTransaction(raw domain.RawItem) (CnbTransaction, error) {
	var t CnbTransaction
	bytes, err := json.Marshal(raw.Payload)

	if err == nil {
		err = json.Unmarshal(bytes, &t)
	}
	return t, err
}
