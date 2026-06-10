package mongo

import (
	"log"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DeleteImortedActivities implements Repo.
func (s FinTrackerMongoStorage) DeleteImortedActivities(ids []string) error {

	if len(ids) == 0 {
		return nil
	}
	err := s.acitivyImports().DeleteMany(s.context(), ids)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil
	}
	return err
}

// DeleteActivities implements Repo.
func (s FinTrackerMongoStorage) DeleteActivities(ids []string) error {

	err := s.acitivities().DeleteMany(s.context(), ids)
	if err != nil {
		log.Printf("Delete Activities error: %v", err)
		return nil
	}
	return err
}

// DeleteActivityLots implements Repo.
func (s FinTrackerMongoStorage) DeleteActivityLots(ids []string) error {

	err := s.acitivityLots().DeleteMany(s.context(), ids)
	if err != nil {
		log.Printf("Delete ActivityLot error: %v", err)
		return nil
	}
	return err
}

// GetImortedActivities
func (s FinTrackerMongoStorage) GetImortedActivities(uid string, acctId string) ([]*domain.ActivityImport, error) {
	filter := bson.M{"uid": uid, "accountId": acctId}
	actvs, err := s.acitivyImports().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil, err
	}
	return actvs, err
}

// GetActivities
func (s FinTrackerMongoStorage) GetActivities(uid string) ([]*domain.Activity, error) {
	filter := bson.M{"uid": uid}
	actvs, err := s.acitivities().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil, err
	}
	return actvs, err
}

// GetActivitiesforAccount
func (s FinTrackerMongoStorage) GetActivitiesForAccount(uid string, acctId string) ([]*domain.Activity, error) {
	filter := bson.M{"uid": uid, "accountId": acctId}
	actvs, err := s.acitivities().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil, err
	}
	return actvs, err
}

// GetActivityLots
func (s FinTrackerMongoStorage) GetActivityLots(uid string) ([]*domain.ActivityLot, error) {
	filter := bson.M{"uid": uid}
	lots, err := s.acitivityLots().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil, err
	}
	return lots, err
}

// GetActivityLots
func (s FinTrackerMongoStorage) GetActivityLotsForAccount(uid string, acctId string) ([]*domain.ActivityLot, error) {
	filter := bson.M{"uid": uid, "accountId": acctId}
	lots, err := s.acitivityLots().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		log.Printf("Delete Imported Activities error: %v", err)
		return nil, err
	}
	return lots, err
}

// Save imported activities
func (s FinTrackerMongoStorage) SaveImportedActivities(actvs []*domain.ActivityImport) error {
	ids := []string{}
	for _, actv := range actvs {
		ids = append(ids, actv.ID)
	}
	s.acitivyImports().BulkWrite(s.context(), ids, actvs)
	return nil
}

// Save activities
func (s FinTrackerMongoStorage) SaveActivities(actvs []*domain.Activity) error {
	ids := []string{}
	for _, actv := range actvs {
		ids = append(ids, actv.ID)
	}
	s.acitivities().BulkWrite(s.context(), ids, actvs)
	return nil
}

// Save activity lots
func (s FinTrackerMongoStorage) SaveActivityLots(lots []*domain.ActivityLot) error {
	ids := []string{}
	for _, lot := range lots {
		ids = append(ids, lot.ID)
	}
	s.acitivityLots().BulkWrite(s.context(), ids, lots)
	return nil
}
