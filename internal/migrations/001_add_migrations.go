package migrations

import (
	"context"

	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {

	migrations.Register(1, "Migrations schema",
		func(database *mongodb.MongoDatabase) error {

			migrationColl := mongodb.GetMongoRepository[string, *migrations.Migration](database)
			err := migrationColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "version", Value: 1}},
					Options: options.Index().SetName("idx_version").SetUnique(false),
				},
			})
			return err
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}
