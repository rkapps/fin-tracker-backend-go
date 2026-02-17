package migrations

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {
	migrations.Register(2, "Ticker Control schema",
		func(database *mongodb.MongoDatabase) error {

			tcColl := mongodb.GetMongoRepository[string, *domain.TickerControl](database)
			err := tcColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
			})
			return err
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)
}
