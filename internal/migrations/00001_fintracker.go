package migrations

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() {

	migrations.Register(storage.FINTRACKER_DB_NAME, 1, "Initial Schema",
		func(database *mongodb.MongoDatabase) error {
			var err error
			if err = createMigrationIndices(database); err != nil {
				return err
			}
			if err = createUserIndices(database); err != nil {
				return err
			}
			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}

func createMigrationIndices(database *mongodb.MongoDatabase) error {

	col := mongodb.GetMongoRepository[string, *migrations.Migration](database)
	err := col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()})
	return err

}

func createUserIndices(database *mongodb.MongoDatabase) error {
	col := mongodb.GetMongoRepository[string, *domain.User](database)
	err := col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex(), createUIDIndex()})
	return err
}
