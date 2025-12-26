package stocks

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	providers "rkapps/fin-providers-go"
	"rkapps/fin-tracker-backend-go/internal/utils"
	"strings"
	"time"

	mongodb "github.com/rkapps/storage-backend-go/mongodb"
	"go.mongodb.org/mongo-driver/v2/bson"
)

var (
	TIINGO_API_TOKEN = os.Getenv("TIINGO_API_TOKEN")
	ALPHA_API_KEY    = os.Getenv("ALPHA_KEY")
)

type StocksService struct {
	client *mongodb.MongoClient
}

func NewMongoService(client *mongodb.MongoClient) Service {

	return StocksService{
		client: client,
	}
}

// DeleteTicker returns the ticker for the exchange:symbol
func (s StocksService) DeleteTicker(ctx context.Context, id string) error {

	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	return tr.DeleteByID(ctx, id)
}

// getTickersByFilter returns all the tickers
func (s StocksService) getTickersByFilter(ctx context.Context, filter any, sort bson.D) (Tickers, error) {
	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	return tr.Find(ctx, filter, sort, 0, 0)
}

// GetTicker returns the ticker for the exchange:symbol
func (s StocksService) GetTicker(ctx context.Context, id string) (*Ticker, error) {

	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	return tr.FindByID(ctx, id)
}

// GetTickerHistory returns the ticker history for the exchange:symbol
func (s StocksService) GetTickerHistory(ctx context.Context, id string) ([]*TickerHistory, error) {

	th := mongodb.NewMongoRepository[*TickerHistory](*s.client)
	filter := bson.D{{Key: FIELD_ID, Value: id}}
	return th.Find(ctx, filter, nil, 0, 0)
}

// GetTickers returns the tickers for the symbols
func (s StocksService) GetTickers(ctx context.Context, symbols []string) (Tickers, error) {

	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	var qs []string
	for _, symbol := range symbols {
		qs = append(qs, strings.ToUpper(symbol))
	}
	filter := bson.M{
		FIELD_SYMBOL: bson.M{"$in": qs},
	}
	return tr.Find(ctx, filter, nil, 0, 0)
}

// LoadTickers returns the tickers for the symbols
func (s StocksService) LoadTickers(ctx context.Context, ts Tickers) error {

	return s.updateTickersEOD(ctx, ts, true)
}

// SearchTicker search tickers based on input fields
func (s StocksService) SearchTicker(ctx context.Context, ts TickerSearch) (Tickers, error) {

	var tks Tickers
	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	criteria := mongodb.SearchCriteria{}
	criteria.IndexName = "idx_search"

	criteria.Query = ts.SearchText
	criteria.AutoCompleteFields = []string{FIELD_SYMBOL, FIELD_EXCHANGE, FIELD_NAME, FIELD_OVERVIEW}

	//default sort field
	sortField := mongodb.CreateSortField(FIELD_MARKETCAP, -1)

	if len(ts.Function) > 0 {

		criteria.Limit = 50
		if strings.Compare(ts.Function, "Top Gainers") == 0 {

			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(FIELD_PRDIFFPERC_SEARCH, "gt", 0))
			sortField = mongodb.CreateSortField(FIELD_PRDIFFPERC_SEARCH, -1)

		} else if strings.Compare(ts.Function, "Top Losers") == 0 {

			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(FIELD_PRDIFFPERC_SEARCH, "lt", 0))
			sortField = mongodb.CreateSortField(FIELD_PRDIFFPERC_SEARCH, 1)

		} else if strings.Compare(ts.Function, "Top Gainers (Ytd)") == 0 {

			//performance.1D.diff
			fieldName := fmt.Sprintf("%s.%s.%s", FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, FIELD_DIFF)
			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(fieldName, "gt", 0))
			sortField = mongodb.CreateSortField(fieldName, -1)

		} else if strings.Compare(ts.Function, "Top Losers (Ytd)") == 0 {

			fieldName := fmt.Sprintf("%s.%s.%s", FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, FIELD_DIFF)
			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(fieldName, "lt", 0))
			sortField = mongodb.CreateSortField(fieldName, 1)
		}

	} else {

		if len(ts.Sectors) > 0 {
			criteria.TokenFields = append(criteria.TokenFields, mongodb.CreateSearchTokenField(FIELD_SECTOR, ts.Sectors))
		}
		if len(ts.Industries) > 0 {
			criteria.TokenFields = append(criteria.TokenFields, mongodb.CreateSearchTokenField(FIELD_INDUSTRY, ts.Industries))
		}

		if len(ts.Strategies) > 0 {
			criteria.TokenFields = append(criteria.TokenFields, mongodb.CreateSearchTokenField(FIELD_STRATEGIES, ts.Strategies))
		}

		//Add yield
		if ts.FromYield > 0 {
			slog.Debug("SearchTicker", "FromYield", ts.FromYield)
			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(FIELD_YIELD, "gte", float64(ts.FromYield)))
		}
		if ts.ToYield > 0 {
			slog.Debug("SearchTicker", "ToYield", ts.ToYield)
			criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(FIELD_YIELD, "lte", float64(ts.ToYield)))
		}

		//performance
		if len(ts.PerfPeriod) > 0 {

			slog.Debug("SearchTicker", "PerfPeriod", ts.PerfPeriod)

			if strings.Compare(ts.PerfPeriod, "1D") == 0 {

				gte := mongodb.CreateSearchRangeField(FIELD_PRDIFFPERC_SEARCH, "gte", float64(ts.FromPerfPerc))
				lte := mongodb.CreateSearchRangeField(FIELD_PRDIFFPERC_SEARCH, "lte", float64(ts.ToPerfPerc))
				criteria.RangeFields = append(criteria.RangeFields, gte)
				criteria.RangeFields = append(criteria.RangeFields, lte)

			} else if utils.CheckStringInArray(ts.PerfPeriod, PerfPeriods) {

				fieldName := fmt.Sprintf("%s.%s.%s", FIELD_PERFORMANCE_SEARCH, ts.PerfPeriod, FIELD_DIFF)
				criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(fieldName, "gte", float64(ts.FromPerfPerc)))
				criteria.RangeFields = append(criteria.RangeFields, mongodb.CreateSearchRangeField(fieldName, "lte", float64(ts.FromPerfPerc)))
			}
		}

	}

	//always add the active flag
	criteria.BooleanFields = append(criteria.BooleanFields, FIELD_ACTIVE)

	//sort
	criteria.SortFields = append(criteria.SortFields, sortField)

	tks, err := tr.Search(ctx, criteria)
	if err != nil {
		slog.Error("SearchTicker", "ERROR", err)
		return tks, err
	}
	if tks == nil {
		tks = Tickers{}
	}
	return tks, nil
}

// SaveTicker saves ticker
func (s StocksService) SaveTicker(ctx context.Context, t *Ticker) error {

	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	tk, _ := tr.FindByID(ctx, t.ID)
	var err error
	if tk == nil {
		err = tr.InsertOne(ctx, t)
	} else {
		err = tr.UpdateOne(ctx, t)
	}
	return err
}

// UpdateEOD updates all stocks with EOD data
func (s StocksService) UpdateEOD(ctx context.Context) error {

	sort := bson.D{{Key: FIELD_SYMBOL, Value: 1}}
	ts, _ := s.getTickersByFilter(ctx, bson.D{}, sort)
	slog.Debug("UpdateEOD", "Tickers", len(ts))
	return s.updateTickersEOD(ctx, ts, false)
}

// UpdateRealtime updates all stocks with real time information
func (s StocksService) UpdateRealtime(ctx context.Context) error {

	sort := bson.D{{Key: FIELD_SYMBOL, Value: 1}}
	ts, _ := s.getTickersByFilter(ctx, bson.D{}, sort)
	slog.Debug("UpdateRealtime", "Tickers", len(ts))
	return s.updateTickersRealtime(ctx, ts)
}

// updateTickersRealtime updates tickers realtime
func (s StocksService) updateTickersRealtime(ctx context.Context, ts Tickers) error {

	today := time.Now()
	tom := time.Now().Add(time.Hour * 48)

	var symbols []string
	for _, t := range ts {
		if t.IsCrypto() {
			symbols = append(symbols, t.Symbol)
		}
	}
	tiingoApi := providers.NewTiingoApi(TIINGO_API_TOKEN)
	ctm := tiingoApi.GetCryptoHistoryEOD(symbols, today, tom)

	var ids []string
	for _, t := range ts {
		ids = append(ids, t.ID)

		tu := &TickerService{T: t}
		tu.updateTickerRealtime(ctm)
	}

	// Save the Ticker data
	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	if len(ts) > 0 {
		err := tr.BulkWrite(ctx, ids, ts)
		if err != nil {
			return err
		}
	}

	return nil
}

// UpdateTickersEOD updates tickers EOD
func (s StocksService) updateTickersEOD(ctx context.Context, ts Tickers, load bool) error {

	var ids []string
	var err error

	var tca []*TickerControl
	var ta []*Ticker

	tr := mongodb.NewMongoRepository[*Ticker](*s.client)
	tcr := mongodb.NewMongoRepository[*TickerControl](*s.client)
	thr := mongodb.NewMongoRepository[*TickerHistory](*s.client)

	alphaApi := providers.NewAlphaAPI(ALPHA_API_KEY)
	tiingoApi := providers.NewTiingoApi(TIINGO_API_TOKEN)
	binanceApi := providers.NewBinanceApi()

	for i, t := range ts {

		//Set the Id
		t.SetId()

		//if loading the data, remove the ticker control and ticker history
		if load {
			tcr := mongodb.NewMongoRepository[*TickerControl](*s.client)
			tcr.DeleteByID(ctx, t.ID)
			thr.DeleteMany(ctx, []string{t.ID})
		}

		//Find ticker control
		tc, _ := tcr.FindByID(ctx, t.ID)
		tservice := NewTickerService(t, tc, load, alphaApi, tiingoApi, binanceApi)
		t, tc, tha := tservice.UpdateTickerEOD()

		ta = append(ta, t)
		ids = append(ids, t.ID)

		// Save the TickerHistory data
		if len(tha) > 0 {
			err = thr.InsertMany(ctx, tha)
			if err != nil {
				slog.Debug("updateTickersEOD", "TickerHistory", len(tha), "Error", err)
				continue
			}
		}

		//update the ticket control once the history is updated
		if tc != nil {
			tca = append(tca, tc)
		}

		if i%25 == 0 || (i == len(ts)-1) {
			slog.Info("updateTickersEOD", "Updating Tickers...", fmt.Sprintf("%d/%d", i+1, len(ts)))
		}
	}

	slog.Info("updateTickersEOD", "Saving Ticker Control Data", len(tca))
	// Save the Ticker Control data
	if len(tca) > 0 {
		err = tcr.BulkWrite(ctx, ids, tca)
		if err != nil {
			return err
		}
	}

	slog.Info("updateTickersEOD", "Saving Ticker Data", len(ts))
	// Save the Ticker data
	if len(ts) > 0 {
		err = tr.BulkWrite(ctx, ids, ts)
		if err != nil {
			return err
		}
	}

	slog.Info("updateTickersEOD", "Saving Data", "Done")
	return err
}
