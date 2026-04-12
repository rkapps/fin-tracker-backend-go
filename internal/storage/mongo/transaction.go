package mongo

import (
	"log/slog"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/storage-backend-go/core"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/v2/bson"
)

// SearchTransactions implements Repo.
func (s MongoStorage) SearchTransactions(userId string, startDate time.Time, endDate time.Time, searchText string) (domain.Transactions, error) {

	criteria := core.SearchCriteria{}
	criteria.IndexName = "idx_search"

	criteria.Query = searchText
	criteria.AutoCompleteFields = []string{domain.FIELD_TRANSACTION_ACCOUNT, domain.FIELD_TRANSACTION_CATEGORY, domain.FIELD_TRANSACTION_GROUP, domain.FIELD_TRANSACTION_DESCRIPTION, domain.FIELD_TRANSACTION_TAG}

	criteria.AddSearchExactField(domain.FIELD_UID, userId)
	criteria.AddSearchDateRangeField(domain.FIELD_DATE, "gte", startDate)
	criteria.AddSearchDateRangeField(domain.FIELD_DATE, "lte", endDate)

	txns, err := s.transaction().Search(s.context(), criteria)
	if err != nil {
		slog.Error("SearchTransaction", "ERROR", err)
		return txns, err
	}
	if txns == nil {
		txns = domain.Transactions{}
	}
	return txns, nil
}

func (s MongoStorage) ImportTransactions(userId string, startDate time.Time, endDate time.Time, txns []*domain.Transaction) error {

	ctxns, _ := s.SearchTransactions(userId, startDate, endDate, "")
	ids := []string{}
	for _, txn := range ctxns {
		ids = append(ids, txn.ID)
	}
	// assign the id
	for _, txn := range txns {
		// txn.ID = userId
		txn.UID = userId
		txn.ID = primitive.NewObjectID().String()

	}

	err := s.transaction().DeleteMany(s.context(), ids)
	if err != nil {
		return err
	}
	return s.transaction().InsertMany(s.context(), txns)
}

func (s MongoStorage) SummaryTransactions(userId string, startDate time.Time, endDate time.Time) ([]domain.TransactionAgg, error) {

	var pipeline []interface{}
	var match map[string]interface{}
	match = make(map[string]interface{})

	match[domain.FIELD_UID] = bson.M{"$eq": userId}
	match[domain.FIELD_TRANSACTION_CATEGORY] = bson.M{"$ne": "Paycheck"}
	match[domain.FIELD_TRANSACTION_GROUP] = bson.M{"$ne": "Others"}
	match[domain.FIELD_TRANSACTION_ACCOUNT] = bson.M{"$ne": "Interest Payment"}

	if !startDate.IsZero() || !endDate.IsZero() {
		if !startDate.IsZero() && !endDate.IsZero() {
			match["date"] = bson.M{"$gte": startDate, "$lte": endDate}
		} else if !startDate.IsZero() {
			// fmt.Println(fromDate)
			match["date"] = bson.M{"$gte": startDate}
		} else {
			match["date"] = bson.M{"$lte": endDate}
		}
	}

	tf1c := bson.M{"$cond": []interface{}{bson.M{"$eq": []interface{}{"$dbcr", "credit"}}, "$amount", bson.D{{"$multiply", bson.A{"$amount", -1}}}}}

	query := bson.M{
		"_id": bson.M{
			"year":     bson.M{"$year": "$date"},
			"month":    bson.M{"$month": "$date"},
			"group":    "$group",
			"category": "$category",
			"account":  "$account",
		},
		"amount": bson.M{"$sum": tf1c},
	}

	queryStage := bson.M{
		"$group": query,
	}

	matchStage := bson.M{
		"$match": match,
	}
	pipeline = append(pipeline, matchStage, queryStage)
	// var results []map[string]interface{}
	var results []domain.TransactionAgg
	err := s.transaction().Aggregate(s.context(), pipeline, &results)
	if err != nil {
		println(err)
	}
	return results, nil
	// taggs := convertToTransactionAgg(results)
	// return taggs, nil
}

// func convertToTransactionAgg(results []map[string]interface{}) []domain.TransactionAgg {

// 	var taggs []domain.TransactionAgg
// 	var tagg domain.TransactionAgg
// 	// fmt.Printf("results: %v\n", len(results))
// 	for _, result := range results {

// 		// fmt.Printf("result: %v\n", result)
// 		tagg = domain.TransactionAgg{}
// 		for _, v := range result {
// 			switch entry := v.(type) {
// 			case map[string]interface{}:
// 				for k, v2 := range entry {
// 					switch v3 := v2.(type) {
// 					case string:
// 					case int32:
// 						if strings.Compare("month", k) == 0 {
// 							tagg.Month = v3
// 						} else if strings.Compare("year", k) == 0 {
// 							tagg.Year = v3
// 						}

// 					default:
// 						// fmt.Printf("--%T\n", v3)
// 						// break
// 					}

// 				}
// 				tagg.Group = fmt.Sprintf("%s", entry["group"])
// 				tagg.Category = fmt.Sprintf("%s", entry["category"])
// 				tagg.Account = fmt.Sprintf("%s", entry["account"])
// 			case float64:
// 				tagg.Amount = entry
// 			default:
// 				fmt.Printf("----%v", entry)
// 			}

// 		}
// 		taggs = append(taggs, tagg)
// 	}

// 	return taggs
// }
