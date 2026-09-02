package mongo

import (
	"context"
	"fmt"
	"log"
	"log/slog"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// DeleteAllSyncCursors implements Repo.
func (s FinTrackerMongoStorage) DeleteAllCursors(ctx context.Context, uid, id string) error {

	cursors, err := s.LoadAllCursors(ctx, uid, id)
	ids := []string{}
	for _, cursor := range cursors {
		ids = append(ids, cursor.ID)
	}
	if len(ids) > 0 {
		err := s.sync_cursors().DeleteMany(s.context(), ids)
		if err != nil {
			log.Printf("Delete Activities error: %v", err)
			return nil
		}
	}
	return err
}

// DeleteAllRawItems implements Repo.
func (s FinTrackerMongoStorage) DeleteAllRawItems(ctx context.Context, uid, id string) error {

	cursors, err := s.LoadAllRawItems(ctx, uid, id)
	ids := []string{}
	for _, cursor := range cursors {
		ids = append(ids, cursor.ID)
	}
	if len(ids) > 0 {
		err := s.sync_raw_items().DeleteMany(s.context(), ids)
		if err != nil {
			log.Printf("Delete Activities error: %v", err)
			return nil
		}
	}
	return err
}

// LoadAllCursors implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) LoadAllCursors(ctx context.Context, uid, acctID string) ([]*domain.SyncCursor, error) {
	filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ACCOUNT_ID: acctID}
	log.Println(filter)
	accts, err := s.sync_cursors().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		slog.Debug("LoadAllCursors", "Error", err)
	}
	return accts, err
}

// LoadAllRawItems implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) LoadAllRawItems(ctx context.Context, uid, acctID string) ([]*domain.RawItem, error) {
	filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ACCOUNT_ID: acctID}
	accts, err := s.sync_raw_items().Find(s.context(), filter, bson.D{}, 0, 0)
	if err != nil {
		slog.Debug("LoadAllRawItems", "Error", err)
	}
	return accts, err
}

// LoadCursor implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) LoadCursor(ctx context.Context, uid, acctID, provider, stream string) (*domain.SyncCursor, error) {
	id := fmt.Sprintf("%s-%s-%s-%s", uid, acctID, provider, stream)
	acct, err := s.sync_cursors().FindByID(s.context(), id)
	if err != nil {
		slog.Debug("Load Cursor", "Error", err)
	}
	return acct, err

}

// MarkProcessed implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) MarkProcessed(ctx context.Context, rawIDs []string, transformVersion int) error {
	panic("unimplemented")
}

// SaveCursor implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) SaveCursor(ctx context.Context, cursor *domain.SyncCursor) error {

	id := fmt.Sprintf("%s-%s-%s-%s", cursor.UID, cursor.AccountID, cursor.Provider, cursor.Stream)
	cursor.ID = id
	return s.sync_cursors().UpdateOne(s.context(), cursor)

}

// UnprocessedRaw implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) UnprocessedRaw(ctx context.Context, uid, acctID string, transformVersion int) ([]domain.RawItem, error) {
	filter := bson.M{domain.FIELD_UID: uid, domain.FIELD_ACCOUNT_ID: acctID}
	results, _ := s.sync_raw_items().Find(ctx, filter, bson.D{}, 0, 0)
	var items []domain.RawItem
	for _, result := range results {
		items = append(items, *result)
	}
	return items, nil
	// panic("unimplemented")
}

// UpsertRaw implements [storage.ProviderStorageService].
func (s FinTrackerMongoStorage) UpsertRaw(ctx context.Context, uid, acctID, provider string, items []domain.RawItem) error {
	ids := []string{}
	ptrs := make([]*domain.RawItem, len(items))
	for i := range items {
		ids = append(ids, items[i].ID)
		ptrs[i] = &items[i]
	}
	return s.sync_raw_items().BulkWrite(ctx, ids, ptrs)
}
