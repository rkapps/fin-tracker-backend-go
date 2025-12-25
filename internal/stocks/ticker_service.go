package stocks

import (
	"fmt"
	"log/slog"
	"rkapps/fin-tracker-backend-go/internal/providers"
	"rkapps/fin-tracker-backend-go/internal/utils"
	"strconv"

	"sort"
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
	"github.com/shopspring/decimal"
)

type TickerService struct {
	T    *Ticker
	Tc   *TickerControl
	Load bool
}

func (service *TickerService) UpdateTickerEOD() (*Ticker, *TickerControl, []*TickerHistory) {

	t := service.T
	tc := service.Tc

	date := time.Now()
	if tc == nil {
		tc = &TickerControl{Symbol: t.Symbol, Exchange: t.Exchange, CreateDate: &date}
		tc.SetId()

		//we are adding this for the first time. set this to true
		t.Active = true
	}

	//load ticket details
	if service.Load {
		service.loadTickerDetails()
	}

	tha, _ := service.loadTickerHistory()
	//sort by ascending
	sort.Slice(tha, func(i, j int) bool {
		return tha[i].Date.Before(tha[j].Date)
	})

	slog.Debug("UpdateTickersEOD", "ID", t.ID, "History Count", len(tha))

	if len(tha) == 0 {
		// no ticker history
		return t, tc, tha
	}

	//update technicals
	service.updateTechnicals(tha, tc.HistoryUpdatedDate)

	//update price
	service.updatePrice(tha)

	//update performance
	service.updatePerformance(tha)

	//match technical strategies
	t.Strategies = service.matchTechnicalStrategies(tha)

	var utha []*TickerHistory
	//Find history records that have not update inserted into repo
	for _, th := range tha {
		if tc.HistoryUpdatedDate == nil || (tc.HistoryUpdatedDate != nil && th.Date.After(*tc.HistoryUpdatedDate)) {
			utha = append(utha, th)
		}
	}

	tc.UpdatedDate = &date
	tc.HistoryUpdatedDate = tc.UpdatedDate

	t.PrDiffPercSearch = utils.ConvertDecimalToFloat64(t.PrDiffPerc)

	return t, tc, utha
}

// updatePrice updates the eod price and technicals for the ticket
func (service *TickerService) updatePrice(tha []*TickerHistory) {

	t := service.T

	var lth *TickerHistory
	var pth *TickerHistory
	lth = tha[len(tha)-1]
	if len(tha) > 1 {
		pth = tha[len(tha)-2]
	}
	t.Technicals = make(map[string]map[string]decimal.Decimal)
	t.Technicals[SMA] = lth.SMA
	t.Technicals[EMA] = lth.EMA
	t.Technicals[RSI] = lth.RSI
	t.PrDate = &lth.Date
	t.PrLast = lth.Close
	t.PrOpen = lth.Open
	t.PrHigh = lth.High
	t.PrLow = lth.Low
	t.PrClose = lth.Close
	if pth != nil {
		t.PrPrev = pth.Close
	}
	t.SetPriceDiff()
}

// updatePerformance updates the price difference
func (service *TickerService) updatePerformance(tha []*TickerHistory) {

	t := service.T

	thm := make(map[string]*TickerHistory)
	var th *TickerHistory
	for _, th := range tha {
		id := t.Exchange + ":" + t.Symbol + ":" + utils.DateFormat1(th.Date)
		thm[id] = th
	}
	t.Performance = make(map[string]map[string]decimal.Decimal)
	for _, period := range PerfPeriods {
		date := utils.DateForPeriod(period)

		for x := range 5 {
			ndate := date.Add(-time.Hour * time.Duration(24*x))
			id := t.Exchange + ":" + t.Symbol + ":" + utils.DateFormat1(ndate)
			th = thm[id]
			if th != nil {
				break
			}
		}

		t.Performance[period] = make(map[string]decimal.Decimal)
		if th == nil {
			t.Performance[period][FIELD_PRICE] = utils.ConvertIntToDecimal(0)
			t.Performance[period][FIELD_DIFF] = utils.ConvertIntToDecimal(0)
		} else {

			decClose := th.Close
			t.Performance[period][FIELD_PRICE] = decClose
			_, t.Performance[period][FIELD_DIFF] = utils.PriceDiff(t.PrLast, decClose)
		}
	}
}

func (service *TickerService) updateTechnicals(tha []*TickerHistory, historyUpdatedDate *time.Time) {

	slog.Debug("updateTechnicals", "HistoryUpdatedDate", historyUpdatedDate)
	series := techan.NewTimeSeries()

	for _, th := range tha {

		period := techan.NewTimePeriod(th.Date, time.Hour*24)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(th.Open.String())
		candle.ClosePrice = big.NewFromString(th.Close.String())
		candle.MaxPrice = big.NewFromString(th.High.String())
		candle.MinPrice = big.NewFromString(th.Low.String())

		series.AddCandle(candle)

		//no need to calculate for already existing historical data
		if historyUpdatedDate != nil && th.Date.Before(*historyUpdatedDate) {
			continue
		}

		closePrices := techan.NewClosePriceIndicator(series)
		lastIndex := series.LastIndex()
		// slog.Debug("updateTechnicals", "lastIndex", lastIndex, "Date", th.Date)

		th.SMA = make(map[string]decimal.Decimal)
		th.EMA = make(map[string]decimal.Decimal)
		th.RSI = make(map[string]decimal.Decimal)

		for _, period := range SMAPeriods {
			smaIndicator := techan.NewSimpleMovingAverage(closePrices, period)
			sma := smaIndicator.Calculate(lastIndex)
			th.SMA[strconv.Itoa(period)] = utils.ConvertFloatToDecimal(sma.Float())
			// slog.Debug("updateTechnicals", "SMA Period", period, "Value", sma)
		}
		for _, period := range EMAPeriods {
			emaIndicator := techan.NewEMAIndicator(closePrices, period)
			ema := emaIndicator.Calculate(lastIndex)
			th.EMA[strconv.Itoa(period)] = utils.ConvertFloatToDecimal(ema.Float())
			// slog.Debug("updateTechnicals", "EMA Period", period, "Value", ema)
		}
		//RSI
		for _, period := range RSIPeriods {
			rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, period)
			rsi := rsiIndicator.Calculate(lastIndex)
			th.RSI[strconv.Itoa(period)] = utils.ConvertFloatToDecimal(rsi.Float())
			// slog.Debug("updateTechnicals", "RSI Period", period, "Value", rsi)
		}
	}

}

// matchTechnicalStrategies updates ticker strategies by running the techan strategy rules
func (Service *TickerService) matchTechnicalStrategies(tha []*TickerHistory) []string {

	strategies := []string{}
	series := createTechanTimeSeries(tha)

	for strategy, buildFunc := range StrategyCatalog {
		record := techan.NewTradingRecord()

		if buildFunc(series) {
			slog.Debug("matchTechnicalStrategies", "Strategy", strategy, "TradingRecord", record)
			strategies = append(strategies, strategy)
		}
	}
	return strategies
}

// updateTickerRealtime updates tickers with realtime data
func (service TickerService) updateTickerRealtime(ctm map[string][]*providers.TTickerHistory) error {

	t := service.T
	today := time.Now()

	if t.IsStock() {

		lp, date := providers.GetTickerRealTimeQuoteFromTiingo(t.Symbol)
		// log.Printf("Ticker: %s lp: %v", t.Symbol, lp)

		if date != nil && utils.DateEqual(today, *date) {

			if t.PrDate == nil || !utils.DateEqual(*date, *t.PrDate) {
				t.PrDate = date
				t.PrPrev = t.PrLast
			}
			t.PrLast = utils.ConvertFloatToDecimal(lp)
			t.SetPriceDiff()
		}

	} else if t.IsCrypto() {

		tha := ctm[t.Symbol]
		if len(tha) == 0 {

			prLast, err := providers.GetTickerPriceFromBinance(t.Symbol)
			if err != nil {
				return err
			}
			// log.Printf("Binance price - %s: %f", t.Symbol, prLast)
			t.PrLast = utils.ConvertFloatToDecimal(prLast)
			if !t.PrPrev.IsZero() {
				t.SetPriceDiff()
			}

		} else {
			th := tha[len(tha)-1]
			t.PrLast = utils.ConvertFloatToDecimal(th.Close)
			t.PrDate = &th.Date
			t.SetPriceDiff()
		}
	}
	return nil
}

func (service *TickerService) loadTickerDetails() *Ticker {

	t := service.T

	if !t.IsStock() {
		return t
	}

	or, url, err := providers.GetTickerDetailsFromAlpha(t.Symbol)
	// time.Sleep(200 * time.Millisecond)

	if err != nil {
		slog.Info("LoadTickerDetailsFromAlpha", "Error", fmt.Sprintf("%s - %v", url, err))
		return t
	}

	var imcap, _ = strconv.Atoi(or.MktCap)

	if len(t.Overview) == 0 {
		t.Overview = or.Overview
	}

	t.MarketCap = int64(imcap)
	t.EPS = utils.ConvertStringToFloat64(or.EPS)
	t.PBRatio = utils.ConvertStringToFloat64(or.PBRatio)
	t.PEGRatio = utils.ConvertStringToFloat64(or.PEGRatio)
	t.PERatio = utils.ConvertStringToFloat64(or.PERatio)
	t.PSRatio = utils.ConvertStringToFloat64(or.PSRatio)
	t.DivAmt = utils.ConvertStringToFloat64(or.DivAmt)
	t.Yield = utils.ConvertStringToFloat64(or.Yield) * 100
	t.Pr52WkHigh = utils.ConvertStringToFloat64(or.Pr52WkHigh)
	t.Pr52WkLow = utils.ConvertStringToFloat64(or.Pr52WkLow)

	divDate := utils.DateFromString(or.PayDate)
	if !divDate.IsZero() {
		t.PayDate = &divDate
	}

	exDivDate := utils.DateFromString(or.ExDivDate)
	if !exDivDate.IsZero() {
		t.ExDivDate = &exDivDate
	}

	//Fix for bad paydate
	if t.ExDivDate == nil {
		t.PayDate = nil
		t.PayRatio = 0
	}
	if t.PayDate == nil {
		t.PayRatio = 0
		t.ExDivDate = nil
		t.Yield = 0
		t.DivAmt = 0
	}
	if t.DivAmt == 0 {
		t.Yield = 0
		t.ExDivDate = nil
		t.PayDate = nil
		t.PayRatio = 0
	}

	t.PayRatio = utils.ConvertStringToFloat64(or.PayRatio) * 100

	return t
}

func (service *TickerService) loadTickerHistory() ([]*TickerHistory, error) {

	var tha []*TickerHistory

	var tthm = make(map[string][]*providers.TTickerHistory)
	var ttha []*providers.TTickerHistory

	et := time.Now()
	st := time.Now().Add(-time.Hour * 24 * 365 * 6)

	t := service.T

	if t.IsCrypto() {
		tthm = providers.GetCryptoHistoryEODFromTiingo([]string{t.Symbol}, st, et)
	} else {
		tthm = providers.GetTickersHistory([]string{t.Symbol}, st, et)
	}

	ttha = tthm[t.Symbol]

	for _, tth := range ttha {

		th := &TickerHistory{}
		th.Metadata.Symbol = t.Symbol
		th.Metadata.Exchange = t.Exchange
		th.SetId()
		th.Date = utils.DateAdjustForUtc(tth.Date)
		th.Date = tth.Date
		th.Close = utils.ConvertFloatToDecimal(tth.Close)
		th.Open = utils.ConvertFloatToDecimal(tth.Open)
		th.High = utils.ConvertFloatToDecimal(tth.High)
		th.Low = utils.ConvertFloatToDecimal(tth.Low)
		th.Volume = utils.ConvertFloatToDecimal(tth.Volume)
		th.SplitFactor = tth.SplitFactor
		tha = append(tha, th)
		// if len(tha) == len(ttha) {
		// 	slog.Debug("loadTickerHistory", "TickerHistory", ttha[len(tha)-1])
		// 	slog.Debug("loadTickerHistory", "TickerHistory", th)
		// }
	}

	return tha, nil

}
