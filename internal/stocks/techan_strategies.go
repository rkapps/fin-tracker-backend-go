package stocks

import (
	"log/slog"

	"github.com/sdcoffey/techan"
)

var StrategyCatalog = map[string]func(series *techan.TimeSeries) techan.RuleStrategy{
	"RSI_OverSold": RSIOverSoldStrategy,
}

// RSIOverSoldStrategy returns a rule strategy for oversold stocks
func RSIOverSoldStrategy(series *techan.TimeSeries) techan.RuleStrategy {

	closePrices := techan.NewClosePriceIndicator(series)
	rsi := techan.NewRelativeStrengthIndexIndicator(closePrices, 14)
	emaFast := techan.NewEMAIndicator(closePrices, 10)

	// 2. Define Entry Rules
	// Signal: RSI < 30 AND price crosses above 10-EMA (Confirmation)
	oversoldThreshold := techan.NewConstantIndicator(30)
	crossDownRule := techan.NewCrossDownIndicatorRule(rsi, oversoldThreshold)

	entryRule := techan.And(
		crossDownRule,
		techan.And(
			techan.NewCrossUpIndicatorRule(closePrices, emaFast), // Price > 10-EMA
			techan.PositionNewRule{},                             // Only if no position open
		),
	)

	// Exit Rule: RSI is over 70
	overboughtThreshold := techan.NewConstantIndicator(70)
	exitRule := techan.NewCrossUpIndicatorRule(rsi, overboughtThreshold)

	return techan.RuleStrategy{
		EntryRule: entryRule,
		ExitRule:  exitRule,
	}
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
