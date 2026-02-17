package migrations

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/migrations"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func init() {
	migrations.Register(3, "Ticker schema",
		func(database *mongodb.MongoDatabase) error {

			tickerColl := mongodb.GetMongoRepository[string, *domain.Ticker](database)

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

			autoCompleteFields := []string{domain.FIELD_SYMBOL, domain.FIELD_EXCHANGE, domain.FIELD_NAME, domain.FIELD_OVERVIEW}
			tokenFields := []string{domain.FIELD_SECTOR, domain.FIELD_INDUSTRY, domain.FIELD_STRATEGIES}
			numberFields := []string{domain.FIELD_MARKETCAP, domain.FIELD_YIELD, domain.FIELD_PRDIFFPERC_SEARCH}
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
		func(client *mongodb.MongoDatabase) error {
			return nil
		},
	)
}
