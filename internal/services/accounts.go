package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type AccountsService struct {
	storage storage.FinTrackerStorageService
	logger  *logger.Logger
}

func NewAccountsService(storage storage.FinTrackerStorageService) AccountsService {
	logger := logger.New()
	alog := logger.For("accounts")
	return AccountsService{storage: storage, logger: alog}
}

func (a AccountsService) CreateAccount(ctx context.Context, uid string, acct *domain.Account) (*domain.Account, error) {
	acct.UID = uid
	acct.ID = uuid.New().String()
	acct.CreatedAt = time.Now()
	a.logger.Info("CreateAccount", "Account", acct)

	// create account state
	err := a.CreateAccountState(ctx, uid, acct.ID)
	if err != nil {
		return nil, fmt.Errorf("CreateAccountState error: %v", err)
	}

	// // create account credential
	// a.CreateAccountCredential(ctx, uid, acct.ID)

	err = a.UpdateAccount(ctx, uid, acct.ID, acct)
	if err != nil {
		return nil, fmt.Errorf("CreateAccount error: %v", err)
	}

	return a.GetAccount(uid, acct.ID)
}

func (a AccountsService) CreateAccountState(ctx context.Context, uid string, id string) error {
	//
	astate := &domain.AccountSyncState{}
	astate.UID = uid
	astate.ID = id
	return a.storage.SaveAccountSyncState(astate)
}

func (a AccountsService) CreateAccountCredential(ctx context.Context, uid string, id string, acred *domain.AccountCredential) error {
	//
	acred.UID = uid
	acred.ID = id
	return a.storage.SaveAccountCredential(acred)
}

func (a AccountsService) DeleteAccount(ctx context.Context, uid string, acctId string) error {

	var err error
	// Delete activities
	if err = a.DeleteImportedActivities(ctx, uid, acctId, time.Time{}); err != nil {
		return err
	}
	if err = a.DeleteActivities(ctx, uid, acctId, time.Time{}); err != nil {
		return err
	}
	if err = a.DeleteActivityLots(ctx, uid, acctId, time.Time{}); err != nil {
		return err
	}
	return a.storage.DeleteAccount(uid, acctId)
}

func (a AccountsService) DeleteImportedActivities(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.storage.GetImortedActivities(uid, acctId)
	if err != nil {
	}
	ids := []string{}
	// find ids to delete
	for _, actv := range actvs {
		if actv.Date.Before(startDate) {
			continue
		}
		ids = append(ids, actv.ID)
	}
	a.logger.Info("DeleteImportActivities", "Ids", len(ids))
	// Delete activities
	a.storage.DeleteImortedActivities(ids)
	return nil
}

func (a AccountsService) DeleteActivities(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.storage.GetActivitiesForAccount(uid, acctId)
	if err != nil {
	}
	ids := []string{}
	if len(ids) == 0 {
		return nil
	}
	// find ids to delete
	for _, actv := range actvs {
		if actv.Date.Before(startDate) {
			continue
		}
		ids = append(ids, actv.ID)
	}
	a.logger.Info("DeleteActivities", "Ids", len(ids))
	// Delete activities
	a.storage.DeleteActivities(ids)
	return nil
}

func (a AccountsService) DeleteActivityLots(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.storage.GetActivityLotsForAccount(uid, acctId)
	if err != nil {
	}
	ids := []string{}
	if len(ids) == 0 {
		return nil
	}
	// find ids to delete
	for _, actv := range actvs {
		if actv.Date.Before(startDate) {
			continue
		}
		ids = append(ids, actv.ID)
	}
	a.logger.Info("DeleteActivityLots", "Ids", len(ids))
	// Delete activities
	a.storage.DeleteActivityLots(ids)
	return nil
}

func (a AccountsService) GetAccounts(uid string) (domain.Accounts, error) {
	return a.storage.GetAccounts(uid)
}

func (a AccountsService) GetAccount(uid string, id string) (*domain.Account, error) {
	return a.storage.GetAccount(uid, id)
}

func (a AccountsService) ImportActivities(ctx context.Context, uid string, acctId string, startDate time.Time, actvs []*domain.ActivityImport) error {

	a.logger.Info("ImportActivities", "AccountId", acctId)

	// first delete the activites from the startDate
	a.DeleteImportedActivities(ctx, uid, acctId, startDate)

	// Import activites
	for _, actv := range actvs {

		id := fmt.Sprintf("%s-%s-%s-%s-%s-%s-%.8v-%s-%s-%s-%.8v-%.8v-%s-%s",
			acctId,
			actv.Date.Format("2006-01-02T15:04:05"), // Full timestamp if available
			actv.TxnType,
			actv.RcvAccount, actv.RcvAddress, actv.RcvCurrency, actv.RcvAmount,
			actv.SentAccount, actv.SentAddress, actv.SentCurrency, actv.SentAmount,
			actv.Fee,
			actv.FeeCurrency,
			actv.Notes, // Include this too
		)
		h := sha1.New()
		h.Write([]byte(id))
		id = hex.EncodeToString(h.Sum(nil))

		actv.UID = uid
		actv.AccountID = acctId
		actv.ID = id
	}
	return a.storage.SaveImportedActivities(actvs)
}

func (a AccountsService) LoadAccounts(ctx context.Context, user domain.User, accts domain.Accounts) error {

	ids := []string{}
	for _, acct := range accts {
		acct.UID = user.ID
		// acct.SetId()
		ids = append(ids, acct.ID)
	}
	// acctColl := mongodb.NewMongoRepository[*accounts.Account](*s.client)
	// return acctColl.BulkWrite(ctx, ids, accts)
	return nil
}

func (a AccountsService) UpdateAccount(ctx context.Context, uid string, id string, acct *domain.Account) error {
	acct.UID = uid
	acct.ID = id
	acct.UpdatedAt = time.Now()
	return a.storage.SaveAccount(acct)
}
