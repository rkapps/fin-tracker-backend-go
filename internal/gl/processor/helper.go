package processor

import "github.com/rkapps/fin-tracker-backend-go/internal/domain"

func defaultLotMatching(category domain.AccountCategory) domain.CostBasisMethod {
	switch category {
	case domain.CategoryCrypto:
		return domain.CostBasisHIFO // minimizes crypto tax burden
	default:
		return domain.CostBasisFIFO // IRS default for securities
	}
}
