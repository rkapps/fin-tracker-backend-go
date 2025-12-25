package providers

import (
	"fmt"
	"log/slog"
	"os"
	"rkapps/fin-tracker-backend-go/internal/utils"
	"strings"
	"time"
)

var (
	TIINGO_API_TOKEN    = os.Getenv("TIINGO_API_TOKEN")
	TIINGO_EOD_URL      = "https://api.tiingo.com/tiingo/daily/"
	TIINGO_REALTIME_URL = "https://api.tiingo.com/iex/"
	TIINGO_CRYPTO_URL   = "https://api.tiingo.com/tiingo/crypto/prices"
	// fundasURL = "https://api.tiingo.com/tiingo/fundamentals/"
	// newsURL   = "https://api.tiingo.com/tiingo/news"
)

type cResponse struct {
	Symbol  string            `json:"ticker"`
	CTicker []*TTickerHistory `json:"priceData"`
}

type rResponse struct {
	Symbol   string     `json:"ticker"`
	TngoLast float64    `json:"tngoLast"`
	Date     *time.Time `json:"timestamp"`
}

type TTickerHistory struct {
	Symbol      string    `json:"symbol"`
	Date        time.Time `json:"date"`
	Open        float64   `json:"open"`
	High        float64   `json:"high"`
	Low         float64   `json:"low"`
	Close       float64   `json:"close"`
	Volume      float64   `json:"volume"`
	SplitFactor float64   `json:"splitFactor"`
}

// GetTickerRealTimeQuoteFromTiingo returns the real time last price and date
func GetTickerRealTimeQuoteFromTiingo(symbol string) (float64, *time.Time) {

	url := strings.Join([]string{TIINGO_REALTIME_URL, symbol, "?token=", TIINGO_API_TOKEN}, "")
	var s = new([]rResponse)
	RunHTTPGet(url, &s)
	for _, res := range *s {
		return res.TngoLast, res.Date
	}
	return 0.0, nil
}

// GetCryptoHistory returns the EOD quotes for the date range
func GetCryptoHistoryEODFromTiingo(symbols []string, st time.Time, et time.Time) map[string][]*TTickerHistory {
	return getCryptoHistory(symbols, &st, &et, "1day")
}

// getCryptoHistory returns the EOD quotes for the date range
func getCryptoHistory(symbols []string, st *time.Time, et *time.Time, frequency string) map[string][]*TTickerHistory {

	var thm = make(map[string][]*TTickerHistory)
	sym := make(map[string]string)
	var sBuilder strings.Builder
	for _, symbol := range symbols {
		ts := strings.ToLower(symbol) + "usd"
		sBuilder.WriteString(ts)
		sBuilder.WriteString(",")
		sym[ts] = symbol
	}

	dates := fmt.Sprintf("&startDate=%s", utils.DateFormat1(*st))
	if et != nil {
		dates = fmt.Sprintf("%s&endDate=%s", dates, utils.DateFormat1(*et))
	}

	url := strings.Join([]string{TIINGO_CRYPTO_URL, "?tickers=", sBuilder.String(), "&resampleFreq=", frequency, dates, "&token=", TIINGO_API_TOKEN}, "")
	var s = new([]cResponse)

	RunHTTPGet(url, s)

	for _, res := range *s {
		symbol := sym[res.Symbol]
		for _, th := range res.CTicker {
			th.Symbol = symbol
			th.Date = utils.DateAdjustForUtc(th.Date)
		}
		thm[symbol] = res.CTicker
	}

	return thm
}

// GetTickerHistoryQuotes returns the historical EOD quotes between the time range.
func GetTickersHistory(symbols []string, st time.Time, et time.Time) map[string][]*TTickerHistory {
	var thm = make(map[string][]*TTickerHistory)
	for _, symbol := range symbols {
		ths := getTickerHistory(symbol, st, et)
		thm[symbol] = ths
	}
	return thm
}

// GetEODTickerQuote gets EOD quote for a ticker
func getTickerHistory(symbol string, st time.Time, et time.Time) []*TTickerHistory {

	var tths []*TTickerHistory
	dates := fmt.Sprintf("&startDate=%s&endDate=%s", utils.DateFormat1(st), utils.DateFormat1(et))
	url := strings.Join([]string{TIINGO_EOD_URL, symbol, "/prices?token=", TIINGO_API_TOKEN, dates}, "")
	err := RunHTTPGet(url, &tths)
	if err != nil {
		slog.Debug("getTickerHistory", "url", url, "error", err)
	}

	return tths
}
