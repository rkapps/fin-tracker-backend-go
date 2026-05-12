package processor

import "github.com/rkapps/fin-tracker-backend-go/internal/domain"

func defaultLotMatching(category domain.AccountCategory) domain.LotMatchingMethod {
	switch category {
	case domain.CategoryCrypto:
		return domain.LotMatchingHIFO // minimizes crypto tax burden
	default:
		return domain.LotMatchingFIFO // IRS default for securities
	}
}
