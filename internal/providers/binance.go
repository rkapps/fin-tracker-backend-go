package providers

import (
	"fmt"
	"strconv"
	"strings"
)

const (
	BINANCE_BASE_URL         = "https://api.binance.com/api/v3/"
	BINANCE_TICKER_PRICE_URL = "ticker/price?symbol="
)

// TickerPrice holds metadata for ticker price data returned by binance
type TickerPrice struct {
	Symbol string `json:"symbol"`
	Price  string `json:"price"`
}

// GetTickerPrice returns the latest price for the symbol
func GetTickerPriceFromBinance(symbol string) (float64, error) {

	var tp TickerPrice
	symbol = strings.ReplaceAll(symbol, "-", "")
	symbol = fmt.Sprintf("%sT", symbol)

	url := fmt.Sprintf("%s%s%s", BINANCE_BASE_URL, BINANCE_TICKER_PRICE_URL, strings.ToUpper(symbol))

	err := RunHTTPGet(url, &tp)
	if err != nil {
		return 0.0, err
	} else {
	}

	price, _ := strconv.ParseFloat(tp.Price, 64)
	return price, nil
}
