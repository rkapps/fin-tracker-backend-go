package services

import (
	"context"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/dto"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type AccountsService struct {
	accountStorage    storage.AccountStorageService
	providerStorage   storage.ProviderStorageService
	providerRegistry  core.SyncRegistry
	encryptionService core.EncryptionService
	logConfig         *logger.Config
	logger            *logger.Logger
}

func NewAccountsService(
	accountStorage storage.AccountStorageService,
	providerStorage storage.ProviderStorageService,
	providerRegistry core.SyncRegistry,
	encryptionService core.EncryptionService,
	logConfig *logger.Config,
) AccountsService {
	alog := logConfig.For("accounts.service")
	return AccountsService{accountStorage: accountStorage,
		providerStorage: providerStorage, providerRegistry: providerRegistry, encryptionService: encryptionService, logConfig: logConfig, logger: alog,
	}
}

func (a AccountsService) CreateAccount(ctx context.Context, uid string, acctReq dto.AccountRequest) (*domain.Account, error) {

	a.logger.Debug("CreateAccount", "Account", acctReq)
	if len(acctReq.ID) == 0 {
		return nil, fmt.Errorf("account id cannot be blank")
	}

	nacct, err := a.GetAccount(uid, acctReq.ID)
	if nacct != nil {
		slog.Error("CreateAccount", "GetAccount", err)
		return nil, fmt.Errorf("account with id '%s' already exists", nacct.ID)
	}

	// Account.UnmarshalJSON now calls this instead of inlining the switch
	detail, err := core.NewAccountDetail(domain.AccountCategory(acctReq.Category), acctReq.Detail)
	if err != nil {
		return nil, fmt.Errorf("coud not unmarshall detail from '%s'", acctReq.Detail)
	}

	acct := &domain.Account{
		ID:              acctReq.ID,
		UID:             uid,
		Name:            acctReq.Name,
		Active:          acctReq.Active,
		Category:        domain.AccountCategory(acctReq.Category),
		Type:            domain.AccountType(acctReq.Type),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		CostBasisMethod: domain.CostBasisMethod(acctReq.CostBasisMethod),
		AlternateNames:  acctReq.AlternateNames,
		Detail:          detail,
	}
	a.logger.Debug("CreateAccount", "Detail", acct.Detail)

	// create account state
	err = a.CreateAccountState(ctx, uid, acct.ID)
	if err != nil {
		slog.Error("CreateAccount", "CreateAccountState", err)
		return nil, fmt.Errorf("CreateAccountState error: %v", err)
	}

	err = a.accountStorage.SaveAccount(acct)
	if err != nil {
		slog.Error("CreateAccount", "SaveAcccount", err)
		return nil, fmt.Errorf("CreateAccount error: %v", err)
	}

	err = a.UpdateAccountCredential(ctx, uid, acct, acctReq)
	if err != nil {
		slog.Error("CreateAccount", "UpdateAccountCredential", err)
	}

	return a.GetAccount(uid, acct.ID)
}

func (a AccountsService) CreateAccountState(ctx context.Context, uid string, id string) error {
	//
	astate := &domain.AccountState{}
	astate.UID = uid
	astate.ID = id
	return a.accountStorage.SaveAccountState(astate)
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
	return a.accountStorage.DeleteAccount(uid, acctId)
}

func (a AccountsService) DeleteImportedActivities(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.accountStorage.GetImportedActivities(uid, acctId)
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
	a.accountStorage.DeleteImportedActivities(ids)
	return nil
}

func (a AccountsService) DeleteActivities(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.accountStorage.GetActivitiesForAccount(uid, acctId)
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
	a.accountStorage.DeleteActivities(ids)
	return nil
}

func (a AccountsService) DeleteActivityLots(ctx context.Context, uid string, acctId string, startDate time.Time) error {

	actvs, err := a.accountStorage.GetActivityLotsForAccount(uid, acctId)
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
	a.accountStorage.DeleteActivityLots(ids)
	return nil
}

func (a AccountsService) GetAccounts(uid string) (domain.Accounts, error) {
	return a.accountStorage.GetAccounts(uid)
}

func (a AccountsService) GetAccount(uid string, id string) (*domain.Account, error) {
	return a.accountStorage.GetAccount(uid, id)
}

func (s AccountsService) GetDecryptedCredential(uid, id string) (json.RawMessage, error) {
	cred, err := s.accountStorage.GetAccountCredential(uid, id)
	if err != nil {
		return nil, err
	}
	return s.encryptionService.Decrypt(cred.EncryptedConfig)
}

func (a AccountsService) ImportActivities(ctx context.Context, uid string, acctId string, startDate time.Time, actvs []*domain.ActivityImport) error {

	a.logger.Info("ImportActivities", "AccountId", acctId)

	// first delete the activites from the startDate
	a.DeleteImportedActivities(ctx, uid, acctId, startDate)

	// Import activites
	for _, actv := range actvs {

		a.logger.Debug("ImportActivities", "rcvSymol", actv.RcvCurrency, "sentsymbol", actv.SentCurrency)

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
	return a.accountStorage.SaveImportedActivities(actvs)
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

func (a AccountsService) UpdateAccount(ctx context.Context, uid string, acctReq dto.AccountRequest) (*domain.Account, error) {

	a.logger.Debug("UpdateAccount", "Account", acctReq.ID)
	if len(acctReq.ID) == 0 {
		return nil, fmt.Errorf("account id cannot be blank")
	}

	acct, _ := a.GetAccount(uid, acctReq.ID)
	if acct == nil {
		return nil, fmt.Errorf("account with id '%s' does not exist", acctReq.ID)
	}

	// Account.UnmarshalJSON now calls this instead of inlining the switch
	detail, err := core.NewAccountDetail(domain.AccountCategory(acctReq.Category), acctReq.Detail)
	if err != nil {
		return nil, fmt.Errorf("coud not unmarshall detail from '%s'", acctReq.Detail)
	}

	acct.Active = acctReq.Active
	acct.AlternateNames = acctReq.AlternateNames
	acct.Category = domain.AccountCategory(acctReq.Category)
	acct.Detail = detail
	acct.CostBasisMethod = domain.CostBasisMethod(acctReq.CostBasisMethod)
	acct.Name = acctReq.Name
	acct.TaxStatus = domain.TaxStatus(acctReq.TaxStatus)
	acct.Type = domain.AccountType(acctReq.Type)
	acct.UpdatedAt = time.Now()

	// Create account state if it does not exist
	aState, _ := a.accountStorage.GetAccountState(uid, acctReq.ID)
	if aState == nil {
		// create account state
		err = a.CreateAccountState(ctx, uid, acct.ID)
		if err != nil {
			slog.Error("CreateAccount", "CreateAccountState", err)
			return nil, fmt.Errorf("CreateAccountState error: %v", err)
		}
	}

	// On Resync, delete all imported activities, sync_cursors, raw_items
	if acctReq.Resync {
		log.Println("reached------------------------------")
		err = a.providerStorage.DeleteAllRawItems(ctx, uid, acctReq.ID)
		if err != nil {
			a.logger.Error("UpdateAccount", "DeleteAllRawItems", err)
		}
		err = a.providerStorage.DeleteAllCursors(ctx, uid, acctReq.ID)
		if err != nil {
			a.logger.Error("UpdateAccount", "DeleteAllCursors", err)
		}
		err = a.DeleteImportedActivities(ctx, uid, acctReq.ID, time.Time{})
		if err != nil {
			a.logger.Error("UpdateAccount", "DeleteImportedActivities", err)
		}

	}

	// Delete Credentials on true otherwise update credentials
	if acctReq.DeleteCredentials {
		err = a.accountStorage.DeleteAccountCredential(uid, acctReq.ID)
	} else {
		err = a.UpdateAccountCredential(ctx, uid, acct, acctReq)
		if err != nil {
			return nil, fmt.Errorf("update account credentials error: %s", err)
		}
	}

	return acct, a.accountStorage.SaveAccount(acct)
}

func (a AccountsService) UpdateAccountCredential(ctx context.Context, uid string, acct *domain.Account, acctReq dto.AccountRequest) error {
	//
	provider := acct.ProviderName()
	if provider == "" || len(acctReq.Credentials) == 0 {
		return nil
	}

	p, err := a.providerRegistry.Get(provider)
	if err != nil {
		slog.Error("CreateAccount", "providerRegistry", err)
		return err
	}

	if err := p.ValidateConfig(acctReq.Credentials); err != nil {
		return err
	}

	encrypted, err := a.encryptionService.Encrypt(acctReq.Credentials)
	if err != nil {
		slog.Error("CreateAccount", "Encrypt", err)
		return fmt.Errorf("encrypting credentials: %w", err)
	}

	acred := &domain.AccountCredential{
		ID:              acctReq.ID,
		UID:             uid,
		Provider:        acct.ProviderName(),
		EncryptedConfig: encrypted,
	}

	return a.accountStorage.SaveAccountCredential(acred)

}
