package cardano

import (
	"log"
	"testing"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
)

var (
	// blockfrost_project_id := os.Getenv("CARDANO_BLOCKFROST_PROJECT_ID")
	blockfrost_project_id = "mainnetKDCkqQjgs3NajWGVi6tdAenSRYwWGxT4"
)

func TestFetchTransactions(t *testing.T) {

	logConfig := logger.New()
	provider := New(NewBlockFrostHTTPClient(blockfrost_project_id), logConfig)
	address := "stake1uyw6k4etc5pfxw88dszqe4zqyl8q5puegzazlf40hzgqynqgps87e"

	saddrs, _ := provider.HTTP.GetAccountAddresses(t.Context(), address)
	for _, addr := range saddrs {
		txns := provider.fetchAllTransactions(t.Context(), addr.Address, 0)
		log.Println(len(txns))
	}
}

func TestTransaction(t *testing.T) {

	logConfig := logger.New()
	provider := New(NewBlockFrostHTTPClient(blockfrost_project_id), logConfig)
	txHash := "0ac41152049f66468a9ee3db7d28bc5ef2bc9c4c00b66abe3f926ff33d30fae5"
	info, _ := provider.HTTP.GetTransactionInfo(txHash)
	log.Println(info)
	metadata, _ := provider.HTTP.GetTransactionMetadata(txHash)
	log.Println(metadata)
	provider.HTTP.GetTransactionDelegations(txHash)
}
