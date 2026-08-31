package core

import (
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

// amountTolerance accounts for precision differences between sources
// reporting the same transfer (e.g. send vs receive amounts differing
// in decimal places). Adjust here if real data shows mismatches.
const amountTolerance = "0.001"

var amountToleranceDecimal = decimal.RequireFromString(amountTolerance)

// AmountsMatch reports whether two amounts represent the same transfer,
// within a fixed tolerance to absorb precision/rounding differences
// between sources.
func AmountsMatch(a, b decimal.Decimal) bool {
	return a.Sub(b).Abs().LessThanOrEqual(amountToleranceDecimal)
}

func IsStableCoin(symbol string) bool {
	if strings.Compare(symbol, "USD") == 0 ||
		// strings.Compare(symbol, "USDT") == 0
		strings.Compare(symbol, "GUSD") == 0 {
		return true
	}
	return false
}

// GetBankAccount implements Service.
func GetBankAccount(accts []*domain.Account, name string) *domain.Account {
	// accts := a.repo.GetAccounts(ctx)
	for _, acct := range accts {

		if strings.Compare(string(acct.Category), "cash") == 0 {
			if strings.Compare(acct.Name, name) == 0 {
				return acct
			}
			// anames := strings.Split(acct.AlternateNames, ",")
			// log.Println(anames)
			for _, aname := range acct.AlternateNames {
				if strings.Compare(aname, name) == 0 {
					return acct
				}
			}
		}
	}
	return nil
}

// GetFirstBankAccount ----Temporary for coinbase because it does not list the withdawal and deposit bank
func GetFirstBankAccount(accts []*domain.Account) *domain.Account {
	for _, acct := range accts {
		if strings.Compare(string(acct.Category), "cash") == 0 {
			return acct
		}
	}
	return nil
}

// GetCoinbaseProAccount ----Pro deposi
func GetCoinbaseProAccount(accts []*domain.Account) *domain.Account {
	for _, acct := range accts {
		// log.Println(acct.ProviderName())
		if strings.Compare(string(acct.ProviderName()), "coinbase_pro") == 0 {
			return acct
		}
	}
	return nil
}
