package ethereum

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
	"github.com/shopspring/decimal"
)

func ConvertERC20Value(value string, tokenDecimals uint8) (decimal.Decimal, error) {
	decExp := decimal.NewFromInt(int64(tokenDecimals))
	return providers.ConvertStringToBaseDecimal(value, decExp)
}
