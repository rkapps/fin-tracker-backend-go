package stocks

import (
	"context"
	"log"
	"os"
	"testing"

	"github.com/rkapps/storage-backend-go/mongodb"
	"github.com/sdcoffey/techan"
)

func getClient() *mongodb.MongoClient {

	reg := mongodb.GetBsonRegistryForDecimal()
	client, err := mongodb.NewMongoClientWithRegistry(os.Getenv("MONGO_ATLAS_CONN_STR"), "test", reg)
	if err != nil {
		log.Fatalf("error connecting to client")
	}
	return client
}

func getHistoryData(id string) ([]*TickerHistory, error) {

	client := getClient()
	stocksService := NewMongoService(client)
	return stocksService.GetTickerHistory(context.Background(), id)
}

func TestTechan(t *testing.T) {

	id := "NYSEARCA:GLD"
	tha, err := getHistoryData(id)
	log.Printf("Ticker ID: %s count: %d", id, len(tha))
	if err != nil {
		log.Println(err)
	}
	series := createTechanTimeSeries(tha)
	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()
	record := &techan.TradingRecord{}

	t.Run("macd", func(t *testing.T) {

		//TREND INDICATORS
		for _, period := range SMAPeriods {
			smaIndicator := techan.NewSimpleMovingAverage(closePrices, period)
			sma := smaIndicator.Calculate(lastIndex)
			log.Printf("SMA Period: %v value: %v", period, sma)
		}
		for _, period := range EMAPeriods {
			emaIndicator := techan.NewEMAIndicator(closePrices, period)
			ema := emaIndicator.Calculate(lastIndex)
			log.Printf("EMA Period: %v value: %v", period, ema)
		}

		//macd
		macdLine := techan.NewMACDIndicator(closePrices, 12, 26)
		signalLine := techan.NewEMAIndicator(macdLine, 9)
		macdHistogram := techan.NewMACDHistogramIndicator(macdLine, 9)

		macdValue := macdLine.Calculate(lastIndex)
		signalValue := signalLine.Calculate(lastIndex)
		histogramValue := macdHistogram.Calculate(lastIndex)
		log.Printf("MACD: %v SIGNAL: %v HISTOGRAM: %v", macdValue, signalValue, histogramValue)
		log.Println("-------------------------------------------------------------------------")
		goldenCrossRule := techan.NewCrossUpIndicatorRule(macdLine, signalLine)

		sma50 := techan.NewSimpleMovingAverage(closePrices, 50)
		sma200 := techan.NewSimpleMovingAverage(closePrices, 200)
		maDeathCross := techan.NewCrossDownIndicatorRule(sma50, sma200)
		macdDeathCross := techan.NewCrossDownIndicatorRule(macdLine, signalLine)

		if goldenCrossRule.IsSatisfied(lastIndex, record) {
			log.Println("Momentum Bullish MACD > Signal Line")
		} else if maDeathCross.IsSatisfied(lastIndex, record) && macdDeathCross.IsSatisfied(lastIndex, record) {
			log.Println("Momentum Bearish MACD < Signal Line")
		}
		log.Println("-------------------------------------------------------------------------")

	})

	t.Run("momentum", func(t *testing.T) {

		//MOMENTUM INDICATORS
		//RSI
		emaFast := techan.NewEMAIndicator(closePrices, 10)
		closePrice := closePrices.Calculate(lastIndex)
		emaFastValue := emaFast.Calculate(lastIndex)
		log.Printf("EMA-10: %v Close Price: %v", emaFastValue, closePrice)

		for _, period := range RSIPeriods {
			rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, period)
			rsi := rsiIndicator.Calculate(lastIndex)
			log.Printf("RSI Period: %v value: %v", period, rsi)
		}
		rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
		rsiOBValue := techan.NewConstantIndicator(70).Calculate(0)
		rsiOSValue := techan.NewConstantIndicator(30).Calculate(0)
		currentRSI := rsiIndicator.Calculate(lastIndex)
		log.Printf("------------------------------------------------------------------------------------")
		if currentRSI.GT(rsiOBValue) && closePrice.LT(emaFastValue) {
			log.Printf("RSI Overbought rsi > 70 and price < ema-10")
		} else if currentRSI.LT(rsiOSValue) && closePrice.GT(emaFastValue) {
			log.Printf("RSI OverSold rsi < 30 and price > ema-10")
		} else {
			log.Printf("RSI ---Normal")
		}
		log.Printf("------------------------------------------------------------------------------------")

		// stochastic oscillator
		kPeriod := 14
		dPeriod := 3
		kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
		dIndicator := techan.NewSlowStochasticIndicator(kIndicator, dPeriod)

		kValue := kIndicator.Calculate(lastIndex)
		dValue := dIndicator.Calculate(lastIndex)
		log.Printf("KVALUE: %v DVALUE: %v", kValue, dValue)
		stochOSValue := techan.NewConstantIndicator(20).Calculate(0)
		stochOBValue := techan.NewConstantIndicator(80).Calculate(0)
		bullishCross := techan.NewCrossUpIndicatorRule(kIndicator, dIndicator)
		bearishCross := techan.NewCrossDownIndicatorRule(kIndicator, dIndicator)

		log.Printf("------------------------------------------------------------------------------------")
		if bullishCross.IsSatisfied(lastIndex, record) {
			if kValue.GT(stochOBValue) {
				log.Printf("Stochastic Bullish crossover - overbought zone")
			}
		} else if bearishCross.IsSatisfied(lastIndex, record) {
			log.Println("stoachastic bearish")
			if dValue.GT(stochOSValue) {
				log.Printf("Stochastic Bearish crossover - oversold zone")
			}
		}
		log.Printf("------------------------------------------------------------------------------------")

	})

	t.Run("volume", func(t *testing.T) {

		//VOLUME INDICATORS
		//average true range
		atrIndicator := techan.NewAverageTrueRangeIndicator(series, 14)
		atrValue := atrIndicator.Calculate(lastIndex)
		log.Printf("Atr Value: %v", atrValue)

		typicalPrices := techan.NewTypicalPriceIndicator(series)
		window := 20
		sigma := 2.0
		//Bollinger bands
		middleBand := techan.NewSimpleMovingAverage(typicalPrices, window)
		upperBand := techan.NewBollingerUpperBandIndicator(typicalPrices, window, sigma)
		lowerBand := techan.NewBollingerLowerBandIndicator(typicalPrices, window, sigma)

		upperValue := upperBand.Calculate(lastIndex)
		lowerValue := lowerBand.Calculate(lastIndex)
		middleValue := middleBand.Calculate(lastIndex)
		typicalValue := typicalPrices.Calculate(lastIndex)

		log.Printf("Bollinger bands Upper: %v Middle: %v Lower: %v - Typical: %v", upperValue, middleValue, lowerValue, typicalValue)
		log.Printf("------------------------------------------------------------------------------------")
		if typicalValue.GT(upperValue) {
			log.Printf("Bearish/Overbought typical value < upper Value")
		} else if typicalValue.LT(lowerValue) {
			log.Printf("Bullish/oversold typical value > lower Value")
		}
		log.Printf("------------------------------------------------------------------------------------")

	})

}
