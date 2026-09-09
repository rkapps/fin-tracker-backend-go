package polkadot

import (
	"os"
	"testing"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
)

func TestRewards(t *testing.T) {

	logConfig := logger.New()
	slog := logConfig.For("test")

	polkadot_api_key := os.Getenv("POLKADOT_API_KEY")
	polkadot_test_address := os.Getenv("POLKADOT_TEST_ADDRESS")

	provider := New(NewPolkadotHttpClient(polkadot_api_key), logConfig)
	data, err := provider.HTTP.GetRewards(polkadot_test_address, 1, 10)
	if err != nil {
		slog.Error("TestRewards", "Error", err)
	}
	if len(data.Data.List) == 0 {
		t.Errorf("Address has zero rewards")
	}
}
