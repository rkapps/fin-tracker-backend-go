package mongo

import (
	"fmt"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s MongoStorage) GetAccount(uid string, id string) (*domain.Account, error) {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	acct, err := s.accounts().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Get Account", "Error", err)
	}
	if acct != nil && acct.UID != uid {
		return nil, fmt.Errorf("Not authorized: %s", id)
	}
	return acct, err

}

func (s MongoStorage) GetAccountSyncState(uid string, id string) (*domain.AccountSyncState, error) {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	acct, err := s.account_sync_states().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Get Account", "Error", err)
	}
	if acct.UID != uid {
		return nil, fmt.Errorf("Not authorized: %s", id)
	}
	return acct, err

}

func (s MongoStorage) GetAccountCredential(uid string, id string) (*domain.AccountCredential, error) {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	acct, err := s.account_credentials().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Get Account", "Error", err)
	}
	if acct.UID != uid {
		return nil, fmt.Errorf("Not authorized: %s", id)
	}
	return acct, err

}

func (s MongoStorage) GetAccounts(uid string) (domain.Accounts, error) {
	filter := bson.M{domain.FIELD_UID: uid}
	accts, err := s.accounts().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		slog.Debug("Get Accounts", "Error", err)
	}
	slog.Debug("Get Accounts", "Filter", filter, "Count", len(accts))
	return accts, err

}

func (s MongoStorage) SaveAccount(data *domain.Account) error {
	return s.accounts().UpdateOne(s.context(), data)
}

func (s MongoStorage) SaveAccountSyncState(data *domain.AccountSyncState) error {
	return s.account_sync_states().UpdateOne(s.context(), data)
}

func (s MongoStorage) SaveAccountCredential(data *domain.AccountCredential) error {
	return s.account_credentials().UpdateOne(s.context(), data)
}
