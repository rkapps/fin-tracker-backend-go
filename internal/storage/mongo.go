package storage

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/rkapps/storage-backend-go/core"
	"github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

func NewMongoStorage(database *mongodb.MongoDatabase) StorageService {
	return MongoStorage{database}
}

func (s MongoStorage) context() context.Context {
	return context.Background()
}

func (s MongoStorage) users() core.Repository[string, *domain.User] {
	return mongodb.GetMongoRepository[string, *domain.User](s.database)
}

func (s MongoStorage) tickers() core.Repository[string, *domain.Ticker] {
	return mongodb.GetMongoRepository[string, *domain.Ticker](s.database)
}

func (s MongoStorage) tickerHistory() core.Repository[string, *domain.TickerHistory] {
	return mongodb.GetMongoRepository[string, *domain.TickerHistory](s.database)
}

func (s MongoStorage) tickerSentiment() core.Repository[string, *domain.TickerSentiment] {
	return mongodb.GetMongoRepository[string, *domain.TickerSentiment](s.database)
}

func (s MongoStorage) tickerEmbedding() core.Repository[string, *domain.TickerEmbedding] {
	return mongodb.GetMongoRepository[string, *domain.TickerEmbedding](s.database)
}

// DeleteTicker returns the ticker for the exchange:symbol
func (s MongoStorage) DeleteTicker(id string) error {
	return s.tickers().DeleteByID(s.context(), id)
}

func (s MongoStorage) GetUser(id string) (*domain.User, error) {
	return s.users().FindByID(s.context(), id)
}

func (s MongoStorage) GetTicker(id string) (*domain.Ticker, error) {
	// log.Println(s.tickers().Count(s.context()))
	ticker, err := s.tickers().FindByID(s.context(), id)
	return ticker, err
}

func (s MongoStorage) GetTickerGroups() (domain.TickerGroups, error) {

	query := bson.M{
		"_id": bson.M{
			"sector":   "$sector",
			"industry": "$industry",
		},
	}

	queryStage := bson.D{{Key: "$group", Value: query}}
	pipeline := bson.A{}
	// matchStage := bson.E{Key: "$match", Value: bson.E{}}
	pipeline = append(pipeline, queryStage)
	// log.Println(pipeline)
	tgs := domain.TickerGroups{}
	results := []domain.TickerGroupsAggregateResult{}

	err := s.tickers().Aggregate(s.context(), pipeline, &results)
	for _, result := range results {
		tgs = append(tgs, &result.ID)
	}
	if err != nil {
		return nil, err
	}
	return tgs, nil
	// return s.tickers().Find(s.context(), id)
}

func (s MongoStorage) GetTickerHistory(symbol string) ([]*domain.TickerHistory, error) {
	filter := bson.D{{Key: domain.FIELD_HISTORY_SYMBOL, Value: symbol}}
	return s.tickerHistory().Find(s.context(), filter, bson.D{}, 0, 0)
}

func (s MongoStorage) GetTickerSentiments(symbol string) ([]*domain.TickerSentiment, error) {
	filter := bson.D{{Key: domain.FIELD_SYMBOL, Value: symbol}}
	return s.tickerSentiment().Find(s.context(), filter, bson.D{}, 0, 0)
}

func (s MongoStorage) GetTickerEmbeddings(symbol string) ([]*domain.TickerEmbedding, error) {
	filter := bson.D{{Key: domain.FIELD_SYMBOL, Value: symbol}}
	return s.tickerEmbedding().Find(s.context(), filter, bson.D{}, 0, 0)
}

func (s MongoStorage) GetTickers(symbols []string) (domain.Tickers, error) {
	var qs []string
	for _, symbol := range symbols {
		if len(symbol) == 0 {
			continue
		}
		qs = append(qs, strings.ToUpper(symbol))
	}
	var filter bson.M
	if len(qs) > 0 {
		filter = bson.M{
			domain.FIELD_SYMBOL: bson.M{"$in": qs},
		}
	} else {
		filter = bson.M{}
	}
	ts, err := s.tickers().Find(s.context(), filter, bson.D{}, 0, 0)
	slog.Debug("Get Tickers", "Symbols", symbols, "Filter", filter, "Count", len(ts))
	return ts, err

}

func (s MongoStorage) SearchTicker(ts domain.TickerSearch) (domain.Tickers, error) {

	var tks domain.Tickers
	criteria := core.SearchCriteria{}
	criteria.IndexName = "idx_search"

	criteria.Query = ts.SearchText
	criteria.AutoCompleteFields = []string{domain.FIELD_SYMBOL, domain.FIELD_EXCHANGE, domain.FIELD_NAME, domain.FIELD_OVERVIEW}

	//default sort field
	criteria.AddSortField(domain.FIELD_MARKETCAP, -1)

	if len(ts.Function) > 0 {

		criteria.Limit = 50
		if strings.Compare(ts.Function, "Top Gainers") == 0 {

			criteria.AddSearchRangeField(domain.FIELD_PRDIFFPERC_SEARCH, "gt", 0)
			criteria.AddSortField(domain.FIELD_PRDIFFPERC_SEARCH, -1)

		} else if strings.Compare(ts.Function, "Top Losers") == 0 {

			criteria.AddSearchRangeField(domain.FIELD_PRDIFFPERC_SEARCH, "lt", 0)
			criteria.AddSortField(domain.FIELD_PRDIFFPERC_SEARCH, -1)

		} else if strings.Compare(ts.Function, "Top Gainers (Ytd)") == 0 {

			//performance.1D.diff
			fieldName := fmt.Sprintf("%s.%s.%s", domain.FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, domain.FIELD_DIFF)
			criteria.AddSearchRangeField(fieldName, "gt", 0)
			criteria.AddSortField(fieldName, -1)

		} else if strings.Compare(ts.Function, "Top Losers (Ytd)") == 0 {

			fieldName := fmt.Sprintf("%s.%s.%s", domain.FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, domain.FIELD_DIFF)
			criteria.AddSearchRangeField(fieldName, "lt", 0)
			criteria.AddSortField(fieldName, 1)
		}

	} else {

		if len(ts.Sectors) > 0 {
			criteria.AddSearchTokenField(domain.FIELD_SECTOR, ts.Sectors)
		}
		if len(ts.Industries) > 0 {
			criteria.AddSearchTokenField(domain.FIELD_INDUSTRY, ts.Industries)
		}

		if len(ts.Strategies) > 0 {
			criteria.AddSearchTokenField(domain.FIELD_STRATEGIES, ts.Strategies)
		}

		//Add yield
		if ts.FromYield > 0 {
			slog.Debug("SearchTicker", "FromYield", ts.FromYield)
			criteria.AddSearchRangeField(domain.FIELD_YIELD, "gte", float64(ts.FromYield))
		}
		if ts.ToYield > 0 {
			slog.Debug("SearchTicker", "ToYield", ts.ToYield)
			criteria.AddSearchRangeField(domain.FIELD_YIELD, "lte", float64(ts.ToYield))
		}

		//performance
		if len(ts.PerfPeriod) > 0 {

			slog.Debug("SearchTicker", "PerfPeriod", ts.PerfPeriod)

			if strings.Compare(ts.PerfPeriod, "1D") == 0 || strings.Compare(ts.PerfPeriod, "N") == 0 {

				criteria.AddSearchRangeField(domain.FIELD_PRDIFFPERC_SEARCH, "gte", float64(ts.FromPerfPerc))
				criteria.AddSearchRangeField(domain.FIELD_PRDIFFPERC_SEARCH, "lte", float64(ts.ToPerfPerc))

			} else if utils.CheckStringInArray(ts.PerfPeriod, domain.PerfPeriods) {

				fieldName := fmt.Sprintf("%s.%s.%s", domain.FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, domain.FIELD_DIFF)
				criteria.AddSearchRangeField(fieldName, "gte", float64(ts.FromPerfPerc))
				criteria.AddSearchRangeField(fieldName, "lte", float64(ts.ToPerfPerc))

			}
		}

	}

	//always add the active flag
	// criteria.BooleanFields = append(criteria.BooleanFields, FIELD_ACTIVE)
	criteria.AddBooleanField(domain.FIELD_ACTIVE)

	//sort
	// criteria.SortFields = append(criteria.SortFields, sortField)
	tks, err := s.tickers().Search(s.context(), criteria)
	if err != nil {
		slog.Error("SearchTicker", "ERROR", err)
		return tks, err
	}
	if tks == nil {
		tks = domain.Tickers{}
	}
	return tks, nil

}
