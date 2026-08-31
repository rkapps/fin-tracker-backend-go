package gl

import (
	"fmt"
	"log"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/gl/processor"
)

func ResolveProcessor(actv domain.Activity, lm processor.LotManager, logConfig *logger.Config) (processor.ActivityProcessor, error) {

	switch actv.TxnType {
	case domain.ActivityTypeRollover, domain.ActivityTypeDeposit, domain.ActivityTypeWithdraw:
		return processor.NewCashActivityProcessor(logConfig), nil
	case domain.ActivityTypeBuy:
		return processor.NewAcquisitionActivityProcessor(logConfig), nil
	case domain.ActivityTypeReceive:
		return processor.NewReceiveActivityProcessor(logConfig), nil
	case domain.ActivityTypeSell:
		return processor.NewDisposalActivityProcessor(logConfig), nil
	case domain.ActivityTypeSend:
		return processor.NewSendActivityProcessor(logConfig), nil
	case domain.ActivityTypeFee, domain.ActivityTypeStakeFee:
		return processor.NewFeeActivityProcessor(logConfig), nil
	case domain.ActivityTypeAdjustment:
		return processor.NewAdjustmentActivityProcessor(logConfig), nil
	case domain.ActivityTypeTransfer:
		return processor.NewTransferActivityProcessor(logConfig), nil
	case domain.ActivityTypeTrade:
		return processor.NewTradeActivityProcessor(logConfig), nil
	case domain.ActivityTypeDividend, domain.ActivityTypeIncome, domain.ActivityTypeInterest:
		return processor.NewIncomeActivityProcessor(logConfig), nil
	case domain.ActivityTypeReward:
		return processor.NewRewardActivityProcessor(logConfig), nil
	case domain.ActivityTypeDelegation:
		return processor.NewDelegationActivityProcessor(logConfig), nil
	case domain.ActivityTypeStake, domain.ActivityTypeUnStake:
		return processor.NewFeeActivityProcessor(logConfig), nil
	case domain.ActivityTypeTradeIn, domain.ActivityTypeAddLiquidity:
		return processor.NewTradeInActivityProcessor(logConfig), nil
	case domain.ActivityTypeTradeOut, domain.ActivityTypeExitLiquidity:
		return processor.NewTradeOutActivityProcessor(logConfig), nil

	}

	log.Printf("ResolveProcess: %v", actv)
	return nil, fmt.Errorf("'%s' activity processor not available.", actv.TxnType)
}
