package refresher

import (
	"context"
	"fmt"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type ImportedAccountRefresher struct {
	storage storage.FinTrackerStorageService
	logger  *logger.Logger
}

func NewImportAccountRefresher(storage storage.FinTrackerStorageService, logConfig *logger.Config) ImportedAccountRefresher {
	plog := logConfig.For("refresher.imported")
	return ImportedAccountRefresher{storage, plog}
}

func (r ImportedAccountRefresher) Refresh(ctx context.Context, account domain.Account, logConfig *logger.Config) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}
	r.logger.Debug("Refresh", "Account", account.UID)

	// user, err := r.Storage.GetUser(account.UID)
	// if err != nil {
	// 	return actvs, fmt.Errorf("User record does not exist")
	// }
	accts, err := r.storage.GetAccounts(account.UID)
	if err != nil {
		return nil, fmt.Errorf("accounts not found for user")
	}
	acctsm := make(map[string]*domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = acct
	}

	iactvs, err := r.storage.GetImortedActivities(account.UID, account.ID)
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
				r.logger.Error("Sent Bank error", "Id", actv.ID, "SentAccount", actv.SentAccount)
				return nil, fmt.Errorf("bank error: %s", actv.SentAccountID)
			}
		}

		if actv.TxnType == domain.ActivityTypeWithdraw {
			actv.RcvAccountID = resolveAccount(acctsm, actv.RcvAccountID, iactv.RcvAccount)
			actv.SentAccountID = resolveAccount(acctsm, actv.AccountID, iactv.SentAccount)
			if len(actv.RcvAccountID) == 0 {
				r.logger.Error("Rcv Bank error", "Id", actv.ID, "RcvAccount", actv.RcvAccount)
				return nil, fmt.Errorf("bank error: %s", actv.RcvAccountID)
			}
		}

		switch iactv.TxnType {
		case string(domain.ActivityTypeRollover), string(domain.ActivityTypeInterest), string(domain.ActivityTypeDividend):
			actv.RcvAmount = iactv.RcvAmount
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvPrice = decimal.NewFromFloat(1.0)
			actv.RcvSymbol = iactv.RcvCurrency
			actv.SentSymbol = iactv.SentCurrency
			actv.RcvAccountID = account.ID

			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeBuy):
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvSymbol = iactv.RcvCurrency
			actv.RcvAmount = iactv.SentAmount
			actv.RcvPrice = actv.RcvAmount.Div(actv.RcvQuantity)
			actv.RcvAccountID = account.ID

			actv.SentAmount = iactv.SentAmount
			actv.SentSymbol = iactv.SentCurrency
			actv.SentQuantity = iactv.SentAmount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.SentAccountID = account.ID

			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeSell):
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvSymbol = iactv.RcvCurrency
			actv.RcvAmount = iactv.RcvAmount
			actv.RcvAccountID = account.ID
			actv.SentAmount = iactv.RcvAmount
			actv.SentSymbol = iactv.SentCurrency
			actv.SentQuantity = iactv.SentAmount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.SentAccountID = account.ID
			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeDeposit):

			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvSymbol = iactv.RcvCurrency
			actv.RcvAmount = iactv.RcvAmount
			actv.RcvAccount = iactv.RcvAccount
			actv.RcvPrice = decimal.NewFromFloat(1.0)
			actv.SentSymbol = iactv.RcvCurrency
			actv.SentQuantity = iactv.RcvAmount
			actv.SentAccount = iactv.SentAccount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.SentAmount = iactv.RcvAmount
			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeWithdraw):
			actv.RcvQuantity = iactv.SentAmount
			actv.RcvSymbol = iactv.SentCurrency
			actv.RcvAmount = iactv.SentAmount
			actv.RcvAccount = iactv.RcvAccount
			actv.RcvPrice = decimal.NewFromFloat(1.0)
			actv.SentAmount = iactv.SentAmount
			actv.SentSymbol = iactv.SentCurrency
			actv.SentQuantity = iactv.SentAmount
			actv.SentAccount = iactv.SentAccount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.Status = domain.ActivityStatusPending

		default:
			continue
		}
		actvs = append(actvs, actv)
	}
	return actvs, nil

}

func resolveAccount(acctsm map[string]*domain.Account, acctId string, account string) string {

	for _, acct := range acctsm {
		if acct.ID == acctId {
			return acctId
		}
		for _, name := range acct.AlternateNames {
			if strings.Compare(strings.ToLower(name), strings.ToLower(account)) == 0 {
				return acct.ID
			}
		}
	}
	return ""
}
