package mongo

import (
	"fmt"
	"log"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s FinTrackerMongoStorage) GetAccount(uid string, id string) (*domain.Account, error) {
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

func (s FinTrackerMongoStorage) GetAccountSyncState(uid string, id string) (*domain.AccountSyncState, error) {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	acct, err := s.accountSyncStates().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Get Account", "Error", err)
	}
	if acct.UID != uid {
		return nil, fmt.Errorf("Not authorized: %s", id)
	}
	return acct, err

}

func (s FinTrackerMongoStorage) GetAccountCredential(uid string, id string) (*domain.AccountCredential, error) {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	acct, err := s.accountCredentials().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Get Account", "Error", err)
	}
	if acct.UID != uid {
		return nil, fmt.Errorf("Not authorized: %s", id)
	}
	return acct, err

}

func (s FinTrackerMongoStorage) GetAccounts(uid string) (domain.Accounts, error) {
	filter := bson.M{domain.FIELD_UID: uid}
	accts, err := s.accounts().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		slog.Debug("Get Accounts", "Error", err)
	}
	slog.Debug("Get Accounts", "Filter", filter, "Count", len(accts))
	return accts, err

}

func (s FinTrackerMongoStorage) GetAccountSummaries(uid string) ([]*domain.AccountSummary, error) {
	filter := bson.M{domain.FIELD_UID: uid}
	accts, err := s.accountSummaries().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		slog.Debug("Get AccountSummaries", "Error", err)
	}
	slog.Debug("Get AccountSummaries", "Filter", filter, "Count", len(accts))
	return accts, err

}

func (s FinTrackerMongoStorage) DeleteAccount(uid string, id string) error {
	// filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ID: id}
	return s.accounts().DeleteByID(s.context(), id)
}

// DeleteAccountSummary
func (s FinTrackerMongoStorage) DeleteAccountSummaries(ids []string) error {

	err := s.accountSummaries().DeleteMany(s.context(), ids)
	if err != nil {
		log.Printf("Delete AccountSummary error: %v", err)
		return nil
	}
	return err
}

func (s FinTrackerMongoStorage) SaveAccount(data *domain.Account) error {
	return s.accounts().UpdateOne(s.context(), data)
}

func (s FinTrackerMongoStorage) SaveAccountSyncState(data *domain.AccountSyncState) error {
	return s.accountSyncStates().UpdateOne(s.context(), data)
}

func (s FinTrackerMongoStorage) SaveAccountCredential(data *domain.AccountCredential) error {
	return s.accountCredentials().UpdateOne(s.context(), data)
}

// Save AccountSummaries
func (s FinTrackerMongoStorage) SaveAccountSummaries(asumys []*domain.AccountSummary) error {
	ids := []string{}
	for _, asum := range asumys {
		ids = append(ids, asum.ID)
	}
	s.accountSummaries().BulkWrite(s.context(), ids, asumys)
	return nil
}
