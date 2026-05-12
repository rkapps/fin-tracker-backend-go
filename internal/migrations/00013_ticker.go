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

	migrations.Register(storage.FINTRACKER_DB_NAME, 13, "Update Ticker indices",
		func(database *mongodb.MongoDatabase) error {
			var err error

			coll := mongodb.GetMongoRepository[string, *domain.TickerControl](database)
			err = coll.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
					Options: options.Index().SetName("idx_symbol").SetUnique(false),
				},
			})
			if err != nil {
				return err
			}

			thColl := mongodb.GetMongoRepository[string, *domain.TickerHistory](database)
			err = thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys: bson.D{
						{Key: "metadata.symbol", Value: 1},
						{Key: "date", Value: -1},
					},
					Options: options.Index().SetName("idx_symbol_date"),
				},
			})
			if err != nil {
				return err
			}

			ecoll := mongodb.GetMongoRepository[string, *domain.TickerEmbedding](database)
			err = ecoll.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
					Options: options.Index().SetName("idx_symbol").SetUnique(false),
				},
				{
					Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
					Options: options.Index().SetName("idx_date").SetUnique(false),
				},
			})
			if err != nil {
				return err
			}

			ncoll := mongodb.GetMongoRepository[string, *domain.TickerNews](database)
			err = ncoll.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
					Options: options.Index().SetName("idx_symbol").SetUnique(false),
				},
				{
					Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
					Options: options.Index().SetName("idx_date").SetUnique(false),
				},
			})
			if err != nil {
				return err
			}

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
