package migrations

import (
	"context"
	"rkapps/fin-tracker-backend-go/internal/stocks"

	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {
	migrations.Register(2, "Ticker Control schema",
		func(client *mongodb.MongoClient) error {

			tcColl := mongodb.NewMongoRepository[*stocks.TickerControl](*client)
			err := tcColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
			})
			return err
		},
		func(client *mongodb.MongoClient) error {
			return nil
		},
	)
}
