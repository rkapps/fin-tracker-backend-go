package processor

import (
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func ResolveProcessor(actv domain.Activity, lm LotManager, logConfig *logger.Config) (ActivityProcessor, error) {

	switch actv.TxnType {
	case domain.ActivityTypeDividend, domain.ActivityTypeInterest, domain.ActivityTypeRollover, domain.ActivityTypeDeposit, domain.ActivityTypeWithdraw:
		return NewCashActivityProcessor(logConfig), nil
	case domain.ActivityTypeBuy:
		return NewAcquisitionActivityProcessor(logConfig), nil
	case domain.ActivityTypeSell:
		return NewDisposalActivityProcessor(logConfig), nil
	}

	return nil, fmt.Errorf("%s activity processor not available.", actv.TxnType)
}
