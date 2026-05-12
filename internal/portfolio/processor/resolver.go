package processor

import (
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func ResolveProcessor(actv domain.Activity, lm LotManager) (ActivityProcessor, error) {

	switch actv.TxnType {
	case domain.ActivityTypeDividend, domain.ActivityTypeInterest, domain.ActivityTypeRollover:
		return NewCashActivityProcessor(), nil
	case domain.ActivityTypeBuy:
		return NewAcquisitionActivityProcessor(), nil
	case domain.ActivityTypeSell:
		return NewDisposalActivityProcessor(), nil
	}

	return nil, fmt.Errorf("%s activity processor not available.", actv.TxnType)
}
