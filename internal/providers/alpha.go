package providers

import (
	"errors"
	"fmt"
	"os"
)

var (
	ALPHA_BASE_URL = "https://www.alphavantage.co"
	ALPHA_API_KEY  = os.Getenv("ALPHA_KEY")
)

type TickerOResponse struct {
	Exchange   string `json:"Exchange"`
	Symbol     string `json:"Symbol"`
	Name       string `json:"Name"`
	Overview   string `json:"Description"`
	Sector     string `json:"Sector"`
	Industry   string `json:"Industry"`
	MktCap     string `json:"MarketCapitalization"`
	PERatio    string `json:"PERatio"`
	PBRatio    string `json:"PriceToBookRatio"`
	PSRatio    string `json:"PriceToSalesRatioTTM"`
	PEGRatio   string `json:"PEGRatio"`
	EPS        string `json:"EPS"`
	DivAmt     string `json:"DividendPerShare"`
	Yield      string `json:"DividendYield"`
	ExDivDate  string `json:"ExDividendDate"`
	PayDate    string `json:"DividendDate"`
	PayRatio   string `json:"PayoutRatio"`
	Pr52WkHigh string `json:"52WeekHigh"`
	Pr52WkLow  string `json:"52WeekLow"`
}

func GetTickerDetailsFromAlpha(symbol string) (*TickerOResponse, string, error) {

	var or *TickerOResponse
	url := fmt.Sprintf("%s/query?function=%s&symbol=%s&apikey=%s", ALPHA_BASE_URL, "OVERVIEW", symbol, ALPHA_API_KEY)
	err := RunHTTPGet(url, &or)
	if err != nil {
		return or, url, err
	}
	if len(or.Symbol) == 0 || len(or.Name) == 0 {
		return or, url, errors.New("symbol is blank")
	}
	if len(or.Exchange) == 0 {
		return or, url, errors.New("exchange is blank")
	}

	return or, url, nil
}
