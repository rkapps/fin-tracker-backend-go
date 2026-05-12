package migrations

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {

	migrations.Register(storage.FINTRACKER_DB_NAME, 2, "Tickers Schema",
		func(database *mongodb.MongoDatabase) error {
			var err error
			if err = createTickerControlIndices(database); err != nil {
				return err
			}
			if err = createTickerIndices(database); err != nil {
				return err
			}
			if err = createTickerHistoryIndices(database); err != nil {
				return err
			}
			if err = createTickerSentimentIndices(database); err != nil {
				return err
			}
			if err = createTickerEmbeddingIndices(database); err != nil {
				return err
			}
			if err = createTickerIndicatorIndices(database); err != nil {
				return err
			}
			if err = createTickerAlphaIndices(database); err != nil {
				return err
			}
			if err = createTickerNews(database); err != nil {
				return err
			}

			return nil
		},
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)

}

func createTickerControlIndices(database *mongodb.MongoDatabase) error {

	coll := mongodb.GetMongoRepository[string, *domain.TickerControl](database)
	err := coll.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()})
	if err != nil {
		return err
	}
	err = coll.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
			Options: options.Index().SetName("idx_symbol").SetUnique(false),
		},
	})
	return err

}

func createTickerIndices(database *mongodb.MongoDatabase) error {

	coll := mongodb.GetMongoRepository[string, *domain.Ticker](database)
	err := coll.CreateIndexes(context.Background(), []mongo.IndexModel{createIdIndex()})
	if err != nil {
		return err
	}
	err = coll.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
			Options: options.Index().SetName("idx_symbol").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_SECTOR, Value: 1}, {Key: domain.FIELD_INDUSTRY, Value: 1}},
			Options: options.Index().SetName("idx_sector_industry").SetUnique(false),
		},
	})

	//Create search index
	opts := options.SearchIndexes().SetName("idx_search").SetType("search")

	autoCompleteFields := []string{domain.FIELD_SYMBOL, domain.FIELD_EXCHANGE, domain.FIELD_NAME, domain.FIELD_OVERVIEW}
	tokenFields := []string{domain.FIELD_SECTOR, domain.FIELD_INDUSTRY, domain.FIELD_STRATEGIES, domain.FIELD_ASSET_TYPE}
	numberFields := []string{domain.FIELD_TOTAL_ASSETS, domain.FIELD_YIELD, domain.FIELD_PRDIFFPERC_SEARCH}
	booleanFields := []string{domain.FIELD_ACTIVE}

	fieldsValue := bson.D{}

	for _, field := range booleanFields {
		fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "boolean"}}}
		fieldsValue = append(fieldsValue, fieldValue)
	}

	for _, field := range autoCompleteFields {
		fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "autocomplete"}}}
		fieldsValue = append(fieldsValue, fieldValue)
	}
	for _, field := range tokenFields {
		fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "token"}}}
		fieldsValue = append(fieldsValue, fieldValue)
	}
	for _, field := range numberFields {
		fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "number"}}}
		fieldsValue = append(fieldsValue, fieldValue)
	}
	for _, period := range domain.PerfPeriods {
		field := fmt.Sprintf("%s.%s.%s", domain.FIELD_PERFORMANCE_SEARCH, period, domain.FIELD_DIFF)
		fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "number"}}}
		fieldsValue = append(fieldsValue, fieldValue)
	}

	err = coll.CreateSearchIndexes(context.Background(), []mongo.SearchIndexModel{
		{
			Options: opts,
			Definition: bson.D{
				{Key: "mappings", Value: bson.D{
					{Key: "dynamic", Value: false},
					{Key: "fields", Value: fieldsValue},
				}},
			},
		},
	},
	)

	return err
}

func createTickerHistoryIndices(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerHistory](database)
	if err := thColl.CreateTimeSeriesCollection(context.Background(), "date", "metadata", "hours"); err != nil {
		return err
	}

	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys: bson.D{
				{Key: "metadata.symbol", Value: 1},
				{Key: "date", Value: -1},
			},
			Options: options.Index().SetName("idx_symbol_date"),
		},
	})
	return err

}

func createTickerSentimentIndices(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerSentiment](database)
	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetName("idx_id").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}, {Key: domain.FIELD_RELEVANCE_SCORE, Value: 1}},
			Options: options.Index().SetName("idx_symbol").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
			Options: options.Index().SetName("idx_date").SetUnique(false),
		},
	})

	return err

}

func createTickerEmbeddingIndices(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerEmbedding](database)
	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetName("idx_id").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
			Options: options.Index().SetName("idx_symbol").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
			Options: options.Index().SetName("idx_date").SetUnique(false),
		},
	})

	return err

}

func createTickerIndicatorIndices(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerIndicator](database)
	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetName("idx_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "symbol", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetName("idx_symbol_date").SetUnique(false),
		},
	})

	return err
}

func createTickerAlphaIndices(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerAlpha](database)
	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetName("idx_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: "key", Value: 1}, {Key: "date", Value: -1}},
			Options: options.Index().SetName("idx_key_n_date").SetUnique(true),
		},
	})

	return err

}

func createTickerNews(database *mongodb.MongoDatabase) error {

	thColl := mongodb.GetMongoRepository[string, *domain.TickerNews](database)
	err := thColl.CreateIndexes(context.Background(), []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "id", Value: 1}},
			Options: options.Index().SetName("idx_id").SetUnique(true),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_SYMBOL, Value: 1}},
			Options: options.Index().SetName("idx_symbol").SetUnique(false),
		},
		{
			Keys:    bson.D{{Key: domain.FIELD_DATE, Value: 1}},
			Options: options.Index().SetName("idx_date").SetUnique(false),
		},
	})

	return err

}
