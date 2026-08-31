package cardano

import (
	"testing"

	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
)

var (
	addrAcctm = map[string]string{}
)

func loadTestAddresses() {
	var testAddresses []string
	utils.LoadFromFile("testdata/addresses.json", &testAddresses)
	for _, addr := range testAddresses {
		addrAcctm[addr] = "1"
	}
}

func loadTransaction(fileName string) Transaction {
	var txn Transaction
	utils.LoadFromFile(fileName, &txn)
	return txn
}
func TestReceiveActivity(t *testing.T) {

	logger := providers.LoadTestLogger()
	loadTestAddresses()
	txn := loadTransaction("testdata/receive_2.json")
	utxo := NewAccountActivity(addrAcctm, nil, logger, false)
	utxo.add_transaction(txn)
	utxo.getActivity()

}

func TestSendActivity(t *testing.T) {

	logger := providers.LoadTestLogger()
	loadTestAddresses()
	txn := loadTransaction("testdata/send.json")
	utxo := NewAccountActivity(addrAcctm, nil, logger, false)
	utxo.add_transaction(txn)
	utxo.getActivity()

}
