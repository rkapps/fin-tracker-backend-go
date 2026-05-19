package services

import (
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func filterAccount(acctIdsm map[string]string, acct *domain.Account, group string, category string, acctIds []string) bool {

	if len(acctIdsm) > 0 {
		if _, ok := acctIdsm[acct.ID]; !ok {
			return false
		}
	}

	return true
}

func filterBankAccount(acctIdsm map[string]*domain.Account, acctId string) bool {

	acct, ok := acctIdsm[acctId]
	if !ok {
		return false
	}

	if acct.Category == domain.CategoryCash {
		return false
	}

	return true
}
