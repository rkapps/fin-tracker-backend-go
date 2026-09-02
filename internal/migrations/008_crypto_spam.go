package migrations

import (
	"context"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func init() {

	migrations.Register(os.Getenv("FINTRACKER_DB_NAME"), 8, "CryptoSpam Schema",
		func(database *mongodb.MongoDatabase) error {

			col := mongodb.GetMongoRepository[string, *domain.CryptoSpam](database)
			if err := col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()}); err != nil {
				return err
			}

			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}
