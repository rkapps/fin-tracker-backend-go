package stocks

import (
	"log/slog"
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
	"RSI_OverSold":   RSIOverSoldStrategy,
	"RSI_OverBought": RsiStochasticOverboughtStrategy,
}

// RSIOverSoldStrategy returns a rule strategy for oversold stocks
func RSIOverSoldStrategy(series *techan.TimeSeries) bool {

	closePrices := techan.NewClosePriceIndicator(series)
	rsi := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)

	emaFast := techan.NewEMAIndicator(closePrices, 10)

	// Define Entry Rules
	// Signal: RSI < 30 AND price crosses above 10-EMA (Confirmation)
	rsiLevel := techan.NewConstantIndicator(30)
	rsiCrossDown := techan.NewCrossDownIndicatorRule(rsi, rsiLevel)

	entryRule := techan.And(
		rsiCrossDown,
		techan.And(
			techan.NewCrossUpIndicatorRule(closePrices, emaFast), // Price > 10-EMA
			techan.PositionNewRule{},                             // Only if no position open
		),
	)

	exitRule := techan.PositionNewRule{}
	rs := techan.RuleStrategy{
		EntryRule: entryRule,
		ExitRule:  exitRule,
	}
	record := techan.NewTradingRecord()
	return rs.ShouldEnter(series.LastIndex(), record)
}

func RsiStochasticOverboughtStrategy(series *techan.TimeSeries) bool {
	// 1. Initialize Indicators
	closePrices := techan.NewClosePriceIndicator(series)

	// RSI 14-period
	rsi := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)

	// Stochastic 14-period %K and 3-period %D
	// stochastic oscillator
	kPeriod := 14
	dPeriod := 3
	kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
	dIndicator := techan.NewSimpleMovingAverage(kIndicator, dPeriod)

	rsiLevel := techan.NewConstantIndicator(70)
	stochLevel := techan.NewConstantIndicator(80)

	// 2. Define the Exit Rule (The Sell Signal)
	// Must be positioned above 80/70 thresholds AND K must cross down D

	// Condition A: K is above 80
	kIsOverbought := techan.NewCrossUpIndicatorRule(kIndicator, stochLevel)

	// Condition B: D is above 80 (ensures both are deep in OB territory)
	dIsOverbought := techan.NewCrossUpIndicatorRule(dIndicator, stochLevel)
	// Condition C: RSI is above 70
	rsiCrossUp := techan.NewCrossUpIndicatorRule(rsi, rsiLevel)
	// Condition D: K crosses down below D (The trigger event)
	kCrossesDownD := techan.NewCrossDownIndicatorRule(kIndicator, dIndicator)

	// Combine all conditions: Must be overbought on both metrics AND the K/D cross happened
	exitRule := techan.And(techan.And(kIsOverbought, dIsOverbought), techan.And(rsiCrossUp,
		kCrossesDownD),
	)

	entryRule := techan.PositionNewRule{} // Opens a position immediately at start

	rs := techan.RuleStrategy{
		EntryRule:      entryRule,
		ExitRule:       exitRule,
		UnstablePeriod: 14, // Wait for enough data
	}
	record := techan.NewTradingRecord()
	return rs.ShouldExit(series.LastIndex(), record)
}

func examples(series *techan.TimeSeries) {

	closePrices := techan.NewClosePriceIndicator(series)
	lastIndex := series.LastIndex()

	//TREND INDICATORS
	for _, period := range SMAPeriods {
		smaIndicator := techan.NewSimpleMovingAverage(closePrices, period)
		sma := smaIndicator.Calculate(lastIndex)
		slog.Debug("updateTechnicals", "SMA Period", period, "Value", sma)
	}
	for _, period := range EMAPeriods {
		emaIndicator := techan.NewEMAIndicator(closePrices, period)
		ema := emaIndicator.Calculate(lastIndex)
		slog.Debug("updateTechnicals", "EMA Period", period, "Value", ema)
	}

	//macd
	macdLine := techan.NewMACDIndicator(closePrices, 12, 26)
	signalLine := techan.NewEMAIndicator(macdLine, 9)
	macdHistogram := techan.NewMACDHistogramIndicator(macdLine, 9)

	macdValue := macdLine.Calculate(lastIndex)
	signalValue := signalLine.Calculate(lastIndex)
	histogramValue := macdHistogram.Calculate(lastIndex)
	slog.Debug("updateTechnicals", "MACD", macdValue, "SIGNAL", signalValue, "HISTOGRAM", histogramValue)

	//MOMENTUM INDICATORS

	//RSI
	for _, period := range RSIPeriods {
		rsiIndicator := techan.NewRelativeStrengthIndexIndicator(closePrices, period)
		rsi := rsiIndicator.Calculate(lastIndex)
		slog.Debug("updateTechnicals", "RSI Period", period, "Value", rsi)
	}

	// stochastic oscillator
	kPeriod := 14
	dPeriod := 3
	kIndicator := techan.NewFastStochasticIndicator(series, kPeriod)
	dIndicator := techan.NewSimpleMovingAverage(kIndicator, dPeriod)

	kValue := kIndicator.Calculate(lastIndex)
	dValue := dIndicator.Calculate(lastIndex)
	slog.Debug("updateTechnicals", "KVALUE", kValue, "DVALUE", dValue)

	//VOLUME INDICATORS

	//average true range
	atrIndicator := techan.NewAverageTrueRangeIndicator(series, 14)
	atrValue := atrIndicator.Calculate(lastIndex)
	slog.Debug("updateTechnicals", "ATRVALUE", atrValue)

	//Bollinger bands
	middleBand := techan.NewSimpleMovingAverage(closePrices, 20)

	upperBand := techan.NewBollingerUpperBandIndicator(closePrices, 20, 2.0)
	lowerBand := techan.NewBollingerLowerBandIndicator(closePrices, 20, 2.0)

	upperValue := upperBand.Calculate(lastIndex)
	lowerValue := lowerBand.Calculate(lastIndex)
	middleValue := middleBand.Calculate(lastIndex)

	slog.Debug("updateTechnicals", "UPPER BAND", upperValue, "MIDDLE BAND", middleValue, "LOWER BAND", lowerValue)
}
