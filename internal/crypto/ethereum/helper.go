package ethereum

import (
	"sort"
	"time"

	"github.com/nanmu42/etherscan-api"
	"github.com/rkapps/fin-tracker-backend-go/internal/crypto"
	"github.com/shopspring/decimal"
)

func ConvertERC20Value(value string, tokenDecimals uint8) (decimal.Decimal, error) {
	decExp := decimal.NewFromInt(int64(tokenDecimals))
	return crypto.ConvertStringToBaseDecimal(value, decExp)
}

func sortTransfers(tsfrsm map[string][]etherscan.ERC20Transfer) []string {

	// Now get a date-ordered list of hashes to process
	hashes := make([]string, 0, len(tsfrsm))
	for h := range tsfrsm {
		hashes = append(hashes, h)
	}
	sort.Slice(hashes, func(i, j int) bool {
		// compare using the FIRST transfer's timestamp for each hash — all transfers
		// under one hash share the same transaction, so any one of them works
		return time.Time(tsfrsm[hashes[i]][0].TimeStamp).Before(time.Time(tsfrsm[hashes[j]][0].TimeStamp))
	})

	return hashes
}
