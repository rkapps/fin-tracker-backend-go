package migrations

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func createIdIndex() mongo.IndexModel {

	return mongo.IndexModel{
		Keys:    bson.D{{Key: domain.FIELD_ID, Value: 1}},
		Options: options.Index().SetName("idx_id").SetUnique(true),
	}
}

func createUIDIndex() mongo.IndexModel {

	return mongo.IndexModel{
		Keys:    bson.D{{Key: domain.FIELD_UID, Value: 1}},
		Options: options.Index().SetName("idx_uid").SetUnique(false),
	}
}
