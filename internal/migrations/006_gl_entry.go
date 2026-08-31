package migrations

import (
	"context"
	"os"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {

	migrations.Register(os.Getenv("FINTRACKER_DB_NAME"), 6, "GLEntry Schema",
		func(database *mongodb.MongoDatabase) error {
			var err error
			col := mongodb.GetMongoRepository[string, *domain.GLEntry](database)
			if err := col.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()}); err != nil {
				return err
			}
			if err = col.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: domain.FIELD_UID, Value: 1}, {Key: domain.FIELD_ACCOUNT_ID, Value: 1}},
					Options: options.Index().SetName("idx_uid_account_date").SetUnique(false),
				},
			}); err != nil {
				return err
			}

			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}
