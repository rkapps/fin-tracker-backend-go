package migrations

import (
	"context"
	"fmt"
	"rkapps/fin-tracker-backend-go/internal/stocks"

	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {
	migrations.Register(3, "Ticker schema",
		func(client *mongodb.MongoClient) error {

			tickerColl := mongodb.NewMongoRepository[*stocks.Ticker](*client)

			err := tickerColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
				{
					Keys:    bson.D{{Key: "symbol", Value: 1}},
					Options: options.Index().SetName("idx_symbol"),
				},
				{
					Keys:    bson.D{{Key: "exchange", Value: 1}},
					Options: options.Index().SetName("idx_exchange"),
				},
			})
			if err != nil {
				return err
			}

			//Create search index
			opts := options.SearchIndexes().SetName("idx_search").SetType("search")

			autoCompleteFields := []string{stocks.FIELD_SYMBOL, stocks.FIELD_EXCHANGE, stocks.FIELD_NAME, stocks.FIELD_OVERVIEW}
			tokenFields := []string{stocks.FIELD_SECTOR, stocks.FIELD_INDUSTRY, stocks.FIELD_STRATEGIES}
			numberFields := []string{stocks.FIELD_MARKETCAP, stocks.FIELD_YIELD, stocks.FIELD_PRDIFFPERC_SEARCH}
			fieldsValue := bson.D{}

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
			for _, period := range stocks.PerfPeriods {
				field := fmt.Sprintf("%s.%s.%s", stocks.FIELD_PERFORMANCE_SEARCH, period, stocks.FIELD_DIFF)
				fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "number"}}}
				fieldsValue = append(fieldsValue, fieldValue)
			}

			err = tickerColl.CreateSearchIndexes(context.Background(), []mongo.SearchIndexModel{
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
		},
		func(client *mongodb.MongoClient) error {
			return nil
		},
	)
}
