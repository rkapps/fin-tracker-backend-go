package migrations

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {

	migrations.Register(storage.FINTRACKER_DB_NAME, 15, "Accounts Schema",
		func(database *mongodb.MongoDatabase) error {
			var err error
			if err = createAccountIndex(database); err != nil {
				return err
			}
			if err = createActivityIndex(database); err != nil {
				return err
			}
			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}

func createAccountIndex(database *mongodb.MongoDatabase) error {
	var err error
	col := mongodb.GetMongoRepository[string, *domain.Account](database)
	if err = col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex(), createUIDIndex()}); err != nil {
		return err
	}

	colc := mongodb.GetMongoRepository[string, *domain.AccountCredential](database)
	if err = colc.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex(), createUIDIndex()}); err != nil {
		return err
	}

	cols := mongodb.GetMongoRepository[string, *domain.AccountSyncState](database)
	if err = cols.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex(), createUIDIndex()}); err != nil {
		return err
	}

	return err
}

func createActivityIndex(database *mongodb.MongoDatabase) error {
	var err error
	col := mongodb.GetMongoRepository[string, *domain.Activity](database)
	if err := col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()}); err != nil {
		return err
	}
	if err = col.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: domain.FIELD_UID, Value: 1}, {Key: domain.FIELD_ACCOUNT_ID, Value: 1}, {Key: domain.FIELD_DATE, Value: 1}},
			Options: options.Index().SetName("idx_uid_account_date").SetUnique(false),
		},
	}); err != nil {
		return err
	}

	coli := mongodb.GetMongoRepository[string, *domain.ActivityImport](database)
	if err = coli.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()}); err != nil {
		return err
	}
	if err = coli.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: domain.FIELD_UID, Value: 1}, {Key: domain.FIELD_ACCOUNT_ID, Value: 1}, {Key: domain.FIELD_DATE, Value: 1}},
			Options: options.Index().SetName("idx_uid_account_date").SetUnique(false),
		},
	}); err != nil {
		return err
	}

	return err
}
