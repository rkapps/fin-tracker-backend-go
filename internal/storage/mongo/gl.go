package mongo

import (
	"log/slog"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func (s FinTrackerMongoStorage) GetGlEntries(uid string, accountIds []string, startDate time.Time, endDate time.Time) ([]*domain.GLEntry, error) {
	filter := bson.M{domain.FIELD_UID: uid}

	if len(accountIds) > 0 {
		filter["accountId"] = bson.M{"$in": accountIds}
	}

	dateFilter := bson.M{}
	if !startDate.IsZero() {
		dateFilter["$gte"] = startDate
	}
	if !endDate.IsZero() {
		dateFilter["$lte"] = endDate
	}
	if len(dateFilter) > 0 {
		filter["taxDate"] = dateFilter
	}

	sort := bson.D{{Key: "taxDate", Value: 1}} // ascending, so callers get chronological order

	entries, err := s.glEntries().Find(s.context(), filter, sort, 0, 0)
	if err != nil {
		slog.Error("GetGlEntries", "Error", err)
		return nil, err
	}
	slog.Debug("GetGlEntries", "Filter", filter, "Count", len(entries))
	return entries, nil

}

// SaveGlEntries
func (s FinTrackerMongoStorage) SaveGlEntries(glEntries []*domain.GLEntry) error {
	ids := []string{}
	for _, asum := range glEntries {
		ids = append(ids, asum.ID)
	}
	s.glEntries().BulkWrite(s.context(), ids, glEntries)
	return nil
}
