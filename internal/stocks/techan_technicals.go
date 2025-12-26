package stocks

import (
	"time"

	"github.com/sdcoffey/big"
	"github.com/sdcoffey/techan"
)

func createTechanTimeSeries(tha []*TickerHistory) *techan.TimeSeries {

	series := techan.NewTimeSeries()
	for _, th := range tha {
		period := techan.NewTimePeriod(th.Date, time.Hour*24)
		candle := techan.NewCandle(period)
		candle.OpenPrice = big.NewFromString(th.Open.String())
		candle.ClosePrice = big.NewFromString(th.Close.String())
		candle.MaxPrice = big.NewFromString(th.High.String())
		candle.MinPrice = big.NewFromString(th.Low.String())

		series.AddCandle(candle)
	}
	return series
}

var StrategyCatalog = map[string]func(series *techan.TimeSeries) bool{
	"RSI_OverBought":           RSIOverBoughtStrategy,
	"RSI_OverSold":             RSIOverSoldStrategy,
	"RSI_StochasticOverBought": RSIOverSoldStrategy,
	"RSI_StochasticOverSold":   RSIOverSoldStrategy,
	"MACD_GoldenCross":         MACDGoldenCrossStrategy,
	"MACD DeathCross":          MACDDeathCrossStrategy,
	"BB_OverBought":            BollingBandsOverBoughtStrategy,
	"BB_OverSold":              BollingBandsOverSoldStrategy,
}

// RSIOverBoughtStrategy returns true based on RSI and EMA
func RSIOverBoughtStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()
	emaFast := techan.NewEMAIndicator(closePrices, 10)
	emaFastValue := emaFast.Calculate(lastIndex)
	closePrice := closePrices.Calculate(lastIndex)

	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	rsiOBValue := techan.NewConstantIndicator(70).Calculate(0)
	currentRSI := rsiIndicator.Calculate(lastIndex)

	// Signal: RSI > 30 AND price crosses below 10-EMA (Confirmation)
	if currentRSI.GT(rsiOBValue) && closePrice.LT(emaFastValue) {
		return true
	} else {
		return false
	}

}

// RSIOverSoldStrategy returns a rule strategy for oversold stocks
func RSIOverSoldStrategy(series *techan.TimeSeries) bool {
	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()
	emaFast := techan.NewEMAIndicator(closePrices, 10)
	emaFastValue := emaFast.Calculate(lastIndex)
	closePrice := closePrices.Calculate(lastIndex)

	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	rsiOSValue := techan.NewConstantIndicator(30).Calculate(0)
	currentRSI := rsiIndicator.Calculate(lastIndex)

	// Signal: RSI < 30 AND price crosses above 10-EMA (Confirmation)
	if currentRSI.LT(rsiOSValue) && closePrice.GT(emaFastValue) {
		return true
	} else {
		return false
	}

}

// RsiStochasticOverboughtStrategy returns true on overbought stocks based on rsi and stocashtic oscillator
func RsiStochasticOverboughtStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)

	lastIndex := series.LastIndex()
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	rsiValue := rsiIndicator.Calculate(lastIndex)
	rsiOBValue := techan.NewConstantIndicator(70).Calculate(0)

	kPeriod := 14
	dPeriod := 3
	kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
	dIndicator := techan.NewSlowStochasticIndicator(kIndicator, dPeriod)

	kValue := kIndicator.Calculate(lastIndex)
	dValue := dIndicator.Calculate(lastIndex)

	stochOBValue := techan.NewConstantIndicator(80).Calculate(0)

	if rsiValue.GT(rsiOBValue) && kValue.GT(stochOBValue) && kValue.LT(dValue) {
		return true
	}
	return false
}

// RsiStochasticOverSoldStrategy returns true on overold stocks based on rsi and stocashtic oscillator
func RsiStochasticOverSoldStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)

	lastIndex := series.LastIndex()
	rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	rsiValue := rsiIndicator.Calculate(lastIndex)
	rsiOSValue := techan.NewConstantIndicator(30).Calculate(0)

	kPeriod := 14
	dPeriod := 3
	kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
	dIndicator := techan.NewSlowStochasticIndicator(kIndicator, dPeriod)

	kValue := kIndicator.Calculate(lastIndex)
	dValue := dIndicator.Calculate(lastIndex)

	stochOSValue := techan.NewConstantIndicator(20).Calculate(0)

	if rsiValue.LT(rsiOSValue) && kValue.LT(stochOSValue) && kValue.GT(dValue) {
		return true
	}
	return false
}

// MACDGoldenCrossStrategy returns true if the stocks MACD line crosses over the SIGNAL line
func MACDGoldenCrossStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()
	// record := techan.NewTradingRecord()

	macdLine := techan.NewMACDIndicator(closePrices, 12, 26)
	// signalLine := techan.NewEMAIndicator(macdLine, 9)
	macdHistogram := techan.NewMACDHistogramIndicator(macdLine, 9)

	prevHistogram := macdHistogram.Calculate(lastIndex - 1).Float()
	currentHistogram := macdHistogram.Calculate(lastIndex).Float()
	if prevHistogram < 0 && currentHistogram > 0 {
		return true
	}
	// goldenCrossRule := techan.NewCrossUpIndicatorRule(macdLine, signalLine)

	// if goldenCrossRule.IsSatisfied(lastIndex, record) {
	// 	return true
	// }
	return false
}

// MACDDeathCrossStrategy returns true if the stocks MACD line crosses over the SIGNAL line
func MACDDeathCrossStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()
	// record := techan.NewTradingRecord()

	//macd
	macdLine := techan.NewMACDIndicator(closePrices, 12, 26)
	macdHistogram := techan.NewMACDHistogramIndicator(macdLine, 9)

	prevHistogram := macdHistogram.Calculate(lastIndex - 1).Float()
	currentHistogram := macdHistogram.Calculate(lastIndex).Float()

	if prevHistogram > 0 && currentHistogram < 0 {
		return true
	}
	// signalLine := techan.NewEMAIndicator(macdHistogram, 9)
	// sma50 := techan.NewSimpleMovingAverage(closePrices, 50)
	// sma200 := techan.NewSimpleMovingAverage(closePrices, 200)
	// maDeathCross := techan.NewCrossDownIndicatorRule(sma50, sma200)
	// macdDeathCross := techan.NewCrossDownIndicatorRule(macdLine, signalLine)

	// if maDeathCross.IsSatisfied(lastIndex, record) && macdDeathCross.IsSatisfied(lastIndex, record) {
	// 	return true
	// }
	return false

}

// BollingBandsOverBoughtStrategy returns true if the price is above the upper band
func BollingBandsOverBoughtStrategy(series *techan.TimeSeries) bool {

	lastIndex := series.LastIndex()
	typicalPrices := techan.NewTypicalPriceIndicator(series)

	window := 20
	sigma := 2.0
	//Bollinger bands
	upperBand := techan.NewBollingerUpperBandIndicator(typicalPrices, window, sigma)
	upperValue := upperBand.Calculate(lastIndex)
	typicalValue := typicalPrices.Calculate(lastIndex)
	if typicalValue.GT(upperValue) {
		return true
	}
	return false
}

// BollingBandsOverSoldStrategy returns true if the price is above middle and low band
func BollingBandsOverSoldStrategy(series *techan.TimeSeries) bool {

	lastIndex := series.LastIndex()
	typicalPrices := techan.NewTypicalPriceIndicator(series)

	window := 20
	sigma := 2.0
	//Bollinger bands
	lowerBand := techan.NewBollingerLowerBandIndicator(typicalPrices, window, sigma)
	lowerValue := lowerBand.Calculate(lastIndex)
	typicalValue := typicalPrices.Calculate(lastIndex)
	if typicalValue.LT(lowerValue) {
		return true
	}
	return false
}
