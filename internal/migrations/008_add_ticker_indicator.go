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
	migrations.Register(8, "Ticker Indicator schema",
		func(database *mongodb.MongoDatabase) error {

			thColl := mongodb.GetMongoRepository[string, *domain.TickerIndicator](database)
			// // dur := 24 * time.Hour
			// // seconds := int64(dur.Seconds())
			// if err := thColl.CreateTimeSeriesCollection(context.Background(), "date", "metadata", "hours"); err != nil {
			// 	return err
			// }
			err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
			})

			err = thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "symbol", Value: 1}, {Key: "date", Value: 1}},
					Options: options.Index().SetName("idx_symbol_date").SetUnique(false),
				},
			})

			return err
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)
}
