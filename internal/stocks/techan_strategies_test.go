package stocks

import (
	"context"
	"log"
	"os"
	"rkapps/fin-tracker-backend-go/internal/utils"
	"testing"
	"time"

	"github.com/rkapps/storage-backend-go/mongodb"
	"github.com/sdcoffey/techan"
)

func getHistory() []*TickerHistory {

	var tha []*TickerHistory
	date := time.Now()
	tha = append(tha, getTickerHistory(date, 412.75, 411.93, 412.63, 408.83))
	tha = append(tha, getTickerHistory(date.Add(-time.Hour*24), 410.3, 413.64, 413.76, 407.1))
	tha = append(tha, getTickerHistory(date.Add(-2*time.Hour*24), 406.98, 408.23, 408.52, 405.72))
	tha = append(tha, getTickerHistory(date.Add(-3*time.Hour*24), 397.92, 399.02, 400.63, 397.17))
	tha = append(tha, getTickerHistory(date.Add(-4*time.Hour*24), 398.28, 398.57, 402.21, 396.05))
	tha = append(tha, getTickerHistory(date.Add(-5*time.Hour*24), 398.08, 399.29, 399.98, 397.25))

	return tha
}

func getTickerHistory(date time.Time, open float64, close float64, high float64, low float64) *TickerHistory {

	th := &TickerHistory{}
	th.Date = date
	th.Open = utils.ConvertFloatToDecimal(open)
	th.Close = utils.ConvertFloatToDecimal(close)
	th.High = utils.ConvertFloatToDecimal(high)
	th.Low = utils.ConvertFloatToDecimal(low)
	return th
}

func getClient() *mongodb.MongoClient {

	reg := mongodb.GetBsonRegistryForDecimal()
	client, err := mongodb.NewMongoClientWithRegistry(os.Getenv("MONGO_ATLAS_CONN_STR"), "test", reg)
	if err != nil {
		log.Fatalf("error connecting to client")
	}
	return client
}

func getHistoryData(client *mongodb.MongoClient, id string) ([]*TickerHistory, error) {

	stocksService := NewMongoService(client)
	return stocksService.GetTickerHistory(context.Background(), id)
}

func TestTechan(t *testing.T) {

	t.Run("rsi", func(t *testing.T) {
		client := getClient()
		tha, err := getHistoryData(client, "NYSEARCA:GLD")
		log.Println(len(tha))
		if err != nil {
			log.Println(err)
		}
		index := len(tha) - 1
		series := createTechanTimeSeries(tha)
		closePrices := techan.NewClosePriceIndicator(series)

		// RSI 14-period
		period := 14
		rsi := techan.NewRelativeStrengthIndexIndicator(closePrices, period)

		log.Printf("Rsi for period: %v - %v", period, rsi.Calculate(index))

		kPeriod := 14
		dPeriod := 3
		kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
		dIndicator := techan.NewSlowStochasticIndicator(kIndicator, dPeriod)

		log.Printf("kIndicator: %v", kIndicator.Calculate(index))
		log.Printf("dIndicator: %v", dIndicator.Calculate(index))

		rsiLevel := techan.NewConstantIndicator(70)
		stochLevel := techan.NewConstantIndicator(80)

		kIsOverbought := techan.NewCrossUpIndicatorRule(kIndicator, stochLevel)
		dIsOverbought := techan.NewCrossUpIndicatorRule(dIndicator, stochLevel)

		rsiCrossDown := techan.NewCrossDownIndicatorRule(rsi, rsiLevel)
		kCrossesDownD := techan.NewCrossDownIndicatorRule(kIndicator, dIndicator)

		record := techan.NewTradingRecord()
		log.Printf("kIsOverbought :%v", kIsOverbought.IsSatisfied(series.LastIndex(), record))
		log.Printf("dIsOverbought :%v", dIsOverbought.IsSatisfied(series.LastIndex(), record))
		log.Printf("kCrossDownD :%v", kCrossesDownD.IsSatisfied(series.LastIndex(), record))
		log.Printf("rsiCrossDown :%v", rsiCrossDown.IsSatisfied(series.LastIndex(), record))
	})

	t.Run("macd", func(t *testing.T) {
		client := getClient()
		tha, err := getHistoryData(client, "NASDAQ:MSFT")
		log.Println(len(tha))
		if err != nil {
			log.Println(err)
		}

		series := createTechanTimeSeries(tha)
		index := series.LastIndex()

		ema := tha[index].EMA
		log.Println(ema["12"])
		log.Println(ema["26"])
		log.Println(ema["12"].Sub(ema["26"]))

		closePrices := techan.NewClosePriceIndicator(series)

		macdLine := techan.NewMACDIndicator(closePrices, 12, 26)
		macdHistogram := techan.NewMACDHistogramIndicator(macdLine, 9)
		signalLine := techan.NewEMAIndicator(macdLine, 9)

		macdValue := macdLine.Calculate(index)
		signalValue := signalLine.Calculate(index)
		histogramValue := macdHistogram.Calculate(index)
		log.Printf("MACD - macdline: %v signalline: %v histogram: %v", macdValue, signalValue, histogramValue)

	})
}
