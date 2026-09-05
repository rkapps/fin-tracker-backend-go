package crypto

import (
	"strings"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/shopspring/decimal"
)

var (
	base         = decimal.NewFromInt(10)
	MAX_DECIMALS = 8
)

const minRewardAmount = "0.25"

var minRewardAmountDecimal = decimal.RequireFromString(minRewardAmount)

func ConvertFloatToTime(curtime float64) (*time.Time, error) {
	var i (int64) = int64(curtime)
	tm := time.Unix(i, 10)
	return &tm, nil
}

func ConvertInt64ToTime(curtime int64) (*time.Time, error) {
	tm := time.Unix(curtime, 10)
	return &tm, nil
}

func ConvertStringToBaseDecimal(quantity string, dec decimal.Decimal) (decimal.Decimal, error) {

	qty, err := utils.ConvertStringToDecimal(quantity)
	if err == nil {
		qty = qty.Div(base.Pow(dec)).RoundUp(int32(MAX_DECIMALS))
	}
	return qty, err
}

func ConvertInt64ToBaseDecimal(quantity int64, dec decimal.Decimal) (decimal.Decimal, error) {

	qty := decimal.NewFromUint64(uint64(quantity))
	qty = qty.Div(base.Pow(dec))
	return qty, nil
}

// evaluate crypto price and return the value and bool
func EvaluateRewardValue(ps core.PriceService, symbol string, amount decimal.Decimal, date time.Time) (value decimal.Decimal, skip bool) {
	price, err := ps.GetCryptoPrice(symbol, date)
	if err != nil {
		// No price data — can't confirm this is negligible. Keep it.
		return decimal.Zero, false
	}
	value = amount.Mul(price).Round(int32(MAX_DECIMALS))
	if value.Abs().LessThan(minRewardAmountDecimal) {
		return value, true
	}
	return value, false
}

func ShouldSkipDustReward(amount decimal.Decimal) bool {
	return amount.LessThan(minRewardAmountDecimal)
}

// Get the account for the address
func GetAccountFromAddress(accts []domain.Account, address string) *domain.Account {
	for _, acct := range accts {
		if strings.Compare(strings.ToLower(acct.Address()), strings.ToLower((address))) == 0 {
			return &acct
		}
	}
	return nil
}

func GetBaseCurrency() string {
	return "USD"
}

func IsCurrency(symbol string) bool {
	if strings.Compare(symbol, "USD") == 0 ||
		strings.Compare(symbol, "GUSD") == 0 {
		return true
	}
	return false
}

func KrakenTime(curtime float64) (*time.Time, error) {

	var i (int64) = int64(curtime)
	tm := time.Unix(i, 10)
	return &tm, nil
}

func LoadTestLogger() *logger.Logger {
	logConfig := logger.New()
	return logConfig.For("test")
}
