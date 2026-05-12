package portfolio

import (
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

// group accounts by provider
func groupAccountsByProvider(accounts []domain.Account) map[string][]domain.Account {
	accountm := make(map[string][]domain.Account)

	//TODO
	return accountm
}

// get account and symbol key
func getAccountSymbolKey(account, currency string) string {
	return fmt.Sprintf("%s:%s", account, currency)
}

// defaultLotMatchingMethod returns the default method for an account category.
func defaultLotMatchingMethod(category domain.AccountCategory) domain.LotMatchingMethod {
	switch category {
	case domain.CategoryCrypto:
		return domain.LotMatchingHIFO
	default:
		return domain.LotMatchingFIFO
	}
}
