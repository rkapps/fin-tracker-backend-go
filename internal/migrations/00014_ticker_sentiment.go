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

	migrations.Register(storage.FINTRACKER_DB_NAME, 14, "Update Ticker Sentiment indices",
		func(database *mongodb.MongoDatabase) error {
			var err error
			scoll := mongodb.GetMongoRepository[string, *domain.TickerSentiment](database)
			err = scoll.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
					Options: options.Index().SetName("idx_date").SetUnique(false),
				},
			})
			if err != nil {
				return err
			}

			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}
