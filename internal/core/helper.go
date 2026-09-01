package core

import (
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

const amountTolerancePercent = "0.005" // 0.5% — adjust based on what you actually observe

var amountTolerancePercentDecimal = decimal.RequireFromString(amountTolerancePercent)

// AmountsMatch reports whether two amounts represent the same transfer,
// using a relative tolerance so it scales correctly across different
// transfer sizes (a fixed tolerance can't work for both small and large amounts).
func AmountsMatch(a, b decimal.Decimal) bool {
	diff := a.Sub(b).Abs()
	base := decimal.Max(a.Abs(), b.Abs())
	if base.IsZero() {
		return diff.IsZero()
	}
	relDiff := diff.Div(base)
	return relDiff.LessThanOrEqual(amountTolerancePercentDecimal)
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
