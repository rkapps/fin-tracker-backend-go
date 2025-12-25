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

func getHistoryData(client *mongodb.MongoClient) ([]*TickerHistory, error) {

	// th := mongodb.NewMongoRepository[*TickerHistory](*client)
	stocksService := NewMongoService(client)
	return stocksService.GetTickerHistory(context.Background(), "NYSEARCA:GLD")
}

func TestTechan(t *testing.T) {

	t.Run("rsi", func(t *testing.T) {
		client := getClient()
		tha, err := getHistoryData(client)
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
	})
}
