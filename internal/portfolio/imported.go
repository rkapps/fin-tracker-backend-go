package portfolio

import (
	"context"
	"fmt"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type ImportedAccountRefresher struct {
	storage storage.AccountStorageService
	logger  *logger.Logger
}

func NewImportAccountRefresher(storage storage.AccountStorageService, logConfig *logger.Config) ImportedAccountRefresher {
	plog := logConfig.For("refresher.imported")
	return ImportedAccountRefresher{storage, plog}
}

func (r ImportedAccountRefresher) Refresh(ctx context.Context, ps core.PriceService, accts []*domain.Account, acreds []domain.AccountWithCredential, logConfig *logger.Config) ([]*domain.Activity, error) {

	// get all accounts
	actvs := []*domain.Activity{}
	acctsm := make(map[string]domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = *acct
	}

	for _, acred := range acreds {

		account := acred.Account
		iactvs, err := r.storage.GetImportedActivities(account.UID, account.ID)
		if err != nil {
			return actvs, nil
		}
		for _, iactv := range iactvs {

			r.logger.Debug("Refresh", "Activity", iactv.ID)

			actv := &domain.Activity{}
			actv.UID = iactv.UID
			actv.ID = iactv.ID
			actv.AccountID = iactv.AccountID
			actv.Date = *iactv.Date
			actv.TxnType = domain.ActivityType(strings.ToLower(iactv.TxnType))
			actv.Notes = iactv.Notes

			if actv.TxnType == domain.ActivityTypeDeposit {
				actv.RcvAccountID = resolveAccount(acctsm, actv.AccountID, iactv.RcvAccount)
				actv.SentAccountID = resolveAccount(acctsm, actv.SentAccountID, iactv.SentAccount)
				if len(actv.SentAccountID) == 0 {
					r.logger.Error("Refresh", "Activity", iactv.ID)
					r.logger.Error("Sent Bank error", "SentAccount", actv.SentAccount)
					r.logger.Error("Sent Bank error", "SentAccountID", actv.SentAccountID)
					// r.logger.Error("Sent Bank error", "Id", actv.ID, "SentAccount", actv.SentAccount)
					return nil, fmt.Errorf("bank error: %s", actv.SentAccountID)
				}
			}

			if actv.TxnType == domain.ActivityTypeWithdraw {
				actv.RcvAccountID = resolveAccount(acctsm, actv.RcvAccountID, iactv.RcvAccount)
				actv.SentAccountID = resolveAccount(acctsm, actv.AccountID, iactv.SentAccount)
				if len(actv.RcvAccountID) == 0 {
					r.logger.Error("Refresh", "Activity", iactv.ID)
					r.logger.Error("Rcv Bank error", "RcvAccount", actv.RcvAccount)
					return nil, fmt.Errorf("bank error: %s", actv.RcvAccountID)
				}
			}

			switch iactv.TxnType {
			case string(domain.ActivityTypeRollover), string(domain.ActivityTypeInterest), string(domain.ActivityTypeDividend):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvPrice = decimal.NewFromFloat(1.0)
				actv.RcvSymbol = iactv.RcvCurrency
				actv.SentSymbol = iactv.SentCurrency
				actv.RcvAccountID = account.ID

				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeReward):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				actv.RcvAccountID = account.ID
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAmount = iactv.SentAmount
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeTrade):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				actv.RcvAccountID = account.ID
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAmount = iactv.SentAmount
				actv.SentAccountID = account.ID
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeBuy):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				actv.RcvPrice = actv.RcvAmount.Div(actv.RcvAmount)
				actv.RcvAccountID = account.ID

				actv.SentAmount = iactv.SentAmount
				actv.SentSymbol = iactv.SentCurrency
				actv.SentPrice = decimal.NewFromFloat(1.0)
				actv.SentAccountID = account.ID
				actv.Fee = iactv.Fee
				actv.FeeCurrency = iactv.FeeCurrency
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeReceive):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				// actv.RcvAmount = iactv.SentAmount
				actv.RcvAccount = iactv.RcvAccount
				// actv.SentAccount = iactv.SentAccount
				actv.RcvAccountID = account.ID
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeSell):
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvAccountID = account.ID
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAmount = iactv.SentAmount
				actv.SentPrice = decimal.NewFromFloat(1.0)
				actv.SentAccountID = account.ID
				actv.Fee = iactv.Fee
				actv.FeeCurrency = iactv.FeeCurrency
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeSend):
				actv.SentAmount = iactv.RcvAmount
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAmount = iactv.SentAmount
				actv.SentAccountID = account.ID
				actv.RcvAccount = iactv.RcvAccount
				price, _ := ps.GetCryptoPrice(actv.SentSymbol, actv.Date)
				actv.RcvAmount = actv.SentAmount.Mul(price)
				actv.RcvSymbol = "USD"
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeAdjustment):
				actv.SentAmount = iactv.SentAmount
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAmount = iactv.SentAmount
				actv.SentAccount = iactv.SentAccount
				actv.SentAccountID = account.ID
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeDeposit):

				actv.RcvAmount = iactv.RcvAmount
				actv.RcvSymbol = iactv.RcvCurrency
				actv.RcvAmount = iactv.RcvAmount
				actv.RcvAccount = iactv.RcvAccount
				actv.RcvPrice = decimal.NewFromFloat(1.0)
				actv.SentSymbol = iactv.RcvCurrency
				actv.SentAccount = iactv.SentAccount
				actv.SentPrice = decimal.NewFromFloat(1.0)
				actv.SentAmount = iactv.RcvAmount
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeWithdraw):
				actv.RcvSymbol = iactv.SentCurrency
				actv.RcvAmount = iactv.SentAmount
				actv.RcvAccount = iactv.RcvAccount
				actv.RcvPrice = decimal.NewFromFloat(1.0)
				actv.SentAmount = iactv.SentAmount
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAccount = iactv.SentAccount
				actv.SentPrice = decimal.NewFromFloat(1.0)
				actv.Status = domain.ActivityStatusPending

			case string(domain.ActivityTypeFee):
				actv.SentAmount = iactv.SentAmount
				actv.SentSymbol = iactv.SentCurrency
				actv.SentAccount = iactv.SentAccount
				actv.SentPrice = decimal.NewFromFloat(1.0)
				actv.Status = domain.ActivityStatusPending

			default:
				continue
			}
			actvs = append(actvs, actv)
		}
	}

	return actvs, nil

}

func resolveAccount(acctsm map[string]domain.Account, acctId string, account string) string {

	for _, acct := range acctsm {
		if acct.ID == acctId {
			return acctId
		}
		// log.Printf("account Id: %s account: %s", acct.ID, account)
		// match the exact account id
		if acct.ID == account {
			return acct.ID
		}
		for _, name := range acct.AlternateNames {
			if strings.Compare(strings.ToLower(name), strings.ToLower(account)) == 0 {
				return acct.ID
			}
		}
	}
	return ""
}
