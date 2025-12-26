package migrations

import (
	"context"
	"rkapps/fin-tracker-backend-go/internal/stocks"
	"time"

	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {
	migrations.Register(4, "Ticker History schema",
		func(client *mongodb.MongoClient) error {

			thColl := mongodb.NewMongoRepository[*stocks.TickerHistory](*client)
			dur := 24 * time.Hour
			// seconds := int64(dur.Seconds())
			if err := thColl.CreateTimeSeriesCollection(context.Background(), "date", "metadata", dur); err != nil {
				return err
			}

			err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(false),
				},
			})

			return err
		},
		func(client *mongodb.MongoClient) error {
			return nil
		},
	)
}
