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
	migrations.Register(11, "Transaction schema",
		func(database *mongodb.MongoDatabase) error {

			txnColl := mongodb.GetMongoRepository[string, *domain.Transaction](database)
			err := txnColl.CreateIndexes(context.Background(), []mongo.IndexModel{
				{
					Keys:    bson.D{{Key: "id", Value: 1}},
					Options: options.Index().SetName("idx_id").SetUnique(true),
				},
			})
			if err != nil {
				return err
			}

			//Create search index
			opts := options.SearchIndexes().SetName("idx_search").SetType("search")

			autoCompleteFields := []string{domain.FIELD_TRANSACTION_GROUP, domain.FIELD_TRANSACTION_CATEGORY, domain.FIELD_TRANSACTION_ACCOUNT, domain.FIELD_TRANSACTION_DESCRIPTION, domain.FIELD_TRANSACTION_TAG}
			exactFields := []string{domain.FIELD_UID}
			dateFields := []string{domain.FIELD_DATE}
			fieldsValue := bson.D{}
			for _, field := range autoCompleteFields {
				fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "autocomplete"}}}
				fieldsValue = append(fieldsValue, fieldValue)
			}
			for _, field := range exactFields {
				fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "token"}}}
				fieldsValue = append(fieldsValue, fieldValue)
			}

			for _, field := range dateFields {
				fieldValue := bson.E{Key: field, Value: bson.D{{Key: "type", Value: "date"}}}
				fieldsValue = append(fieldsValue, fieldValue)
			}

			err = txnColl.CreateSearchIndexes(context.Background(), []mongo.SearchIndexModel{
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
