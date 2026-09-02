package mongo

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s FinTrackerMongoStorage) GetCryptoPrices() ([]*domain.CryptoPrice, error) {
	return s.crypto_prices().Find(s.context(), bson.M{}, bson.D{}, 0, 0)
}

func (s FinTrackerMongoStorage) SaveCryptoPrices(cprices []*domain.CryptoPrice) error {
	ids := []string{}
	for _, cprice := range cprices {
		ids = append(ids, cprice.ID)
	}
	return s.crypto_prices().BulkWrite(s.context(), ids, cprices)
}
