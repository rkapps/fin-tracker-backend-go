package refresher

import (
	"context"
	"log/slog"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/shopspring/decimal"
)

type ImportedAccountRefresher struct {
	Storage storage.StorageService
}

func (r ImportedAccountRefresher) Refresh(ctx context.Context, account domain.Account) ([]*domain.Activity, error) {

	actvs := []*domain.Activity{}
	slog.Debug("Refresh", "Account", account.UID, "Storage", r.Storage)

	// user, err := r.Storage.GetUser(account.UID)
	// if err != nil {
	// 	return actvs, fmt.Errorf("User record does not exist")
	// }

	iactvs, err := r.Storage.GetImortedActivities(account.UID, account.ID)
	if err != nil {
		return actvs, nil
	}
	for _, iactv := range iactvs {
		actv := &domain.Activity{}
		actv.UID = iactv.UID
		actv.ID = iactv.ID
		actv.AccountID = iactv.AccountID
		actv.Date = *iactv.Date
		actv.TxnType = domain.ActivityType(strings.ToLower(iactv.TxnType))
		actv.Notes = iactv.Notes

		switch iactv.TxnType {
		case string(domain.ActivityTypeRollover), string(domain.ActivityTypeInterest), string(domain.ActivityTypeDividend):
			actv.RcvAmount = iactv.RcvAmount
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvPrice = decimal.NewFromFloat(1.0)
			actv.RcvSymbol = iactv.RcvCurrency
			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeBuy):
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvSymbol = iactv.RcvCurrency
			actv.RcvAmount = iactv.SentAmount
			actv.RcvPrice = actv.RcvAmount.Div(actv.RcvQuantity)
			actv.SentAmount = iactv.SentAmount
			actv.SentSymbol = iactv.SentCurrency
			actv.SentQuantity = iactv.SentAmount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.Status = domain.ActivityStatusPending

		case string(domain.ActivityTypeSell):
			actv.RcvQuantity = iactv.RcvAmount
			actv.RcvSymbol = iactv.RcvCurrency
			actv.RcvAmount = iactv.RcvAmount
			actv.RcvPrice = actv.RcvAmount.Div(actv.RcvQuantity)

			actv.SentAmount = iactv.RcvAmount
			actv.SentSymbol = iactv.SentCurrency
			actv.SentQuantity = iactv.SentAmount
			actv.SentPrice = decimal.NewFromFloat(1.0)
			actv.Status = domain.ActivityStatusPending

		default:
			continue
		}
		actvs = append(actvs, actv)
	}
	return actvs, nil

}
