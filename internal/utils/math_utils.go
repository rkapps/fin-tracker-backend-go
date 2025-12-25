package utils

import (
	"math"
	"strconv"

	"github.com/shopspring/decimal"
)

// ConvertDecimalToFloat64 converts decimal to float64
func ConvertDecimalToFloat64(dec decimal.Decimal) float64 {
	val, _ := dec.Float64()
	return val
}

// ConvertFloat64ToDecimal converts  float64 to decimal
func ConvertFloatToDecimal(value float64) decimal.Decimal {
	return decimal.NewFromFloatWithExponent(value, -6)
}

// ConvertIntoToDecimal converts int to decimal
func ConvertIntToDecimal(value int) decimal.Decimal {
	return decimal.NewFromInt(int64(value))
}

// ConvertDecimalToFloat64 converts decimal to float64
func ConvertStringToFloat64(dec string) float64 {
	val, _ := strconv.ParseFloat(dec, 64)
	return val
}

// PriceDiff calculates the difference in amount and percentage
func PriceDiff(last decimal.Decimal, prev decimal.Decimal) (diff decimal.Decimal, percent decimal.Decimal) {

	diff = last.Sub(prev)
	if !diff.IsZero() {
		percent = diff.Div(prev).Mul(decimal.NewFromInt(100))
	} else {
		percent = decimal.Zero
	}
	return diff, percent.RoundUp(2)
}

// RoundUp rounds the float value to 2 decimal places
func RoundUp(x float64) float64 {
	return math.Ceil(x*100) / 100
}
