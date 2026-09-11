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

	// rollover, deposit and withdraw
	case domain.ActivityTypeRollover, domain.ActivityTypeDeposit, domain.ActivityTypeWithdraw:
		return processor.NewCashActivityProcessor(logConfig), nil

	// buy
	case domain.ActivityTypeBuy:
		return processor.NewAcquisitionActivityProcessor(logConfig), nil

	// receive
	case domain.ActivityTypeReceive:
		return processor.NewReceiveActivityProcessor(logConfig), nil

	// sell
	case domain.ActivityTypeSell:
		return processor.NewDisposalActivityProcessor(logConfig), nil

	// send
	case domain.ActivityTypeSend:
		return processor.NewSendActivityProcessor(logConfig), nil

	// fee, stakefee
	case domain.ActivityTypeFee, domain.ActivityTypeStakeFee:
		return processor.NewFeeActivityProcessor(logConfig), nil

	// adjustment
	case domain.ActivityTypeAdjustment:
		return processor.NewAdjustmentActivityProcessor(logConfig), nil

	// transfer
	case domain.ActivityTypeTransfer:
		return processor.NewTransferActivityProcessor(logConfig), nil

	// trade (currency-stablecoin)
	case domain.ActivityTypeTrade:
		return processor.NewTradeActivityProcessor(logConfig), nil

	// stock dividends, income and interest
	case domain.ActivityTypeDividend, domain.ActivityTypeIncome, domain.ActivityTypeInterest:
		return processor.NewIncomeActivityProcessor(logConfig), nil

	// rewards, rebates and airdrops
	case domain.ActivityTypeReward, domain.ActivityTypeRebate, domain.ActivityTypeAirdrop:
		return processor.NewRewardActivityProcessor(logConfig), nil

	// delegations on cardano, solana. used to reduce fee
	case domain.ActivityTypeDelegation:
		return processor.NewDelegationActivityProcessor(logConfig), nil

	// stake and unstake - nothing really happens
	case domain.ActivityTypeStake, domain.ActivityTypeUnStake:
		return processor.NewFeeActivityProcessor(logConfig), nil

	// addliquidity - dispose the sent
	case domain.ActivityTypeTradeIn, domain.ActivityTypeAddLiquidity:
		return processor.NewTradeInActivityProcessor(logConfig), nil

	// exitliquidity - add the receive
	case domain.ActivityTypeTradeOut, domain.ActivityTypeExitLiquidity:
		return processor.NewTradeOutActivityProcessor(logConfig), nil

	// lost
	case domain.ActivityTypeLost:
		return processor.NewLostActivityProcessor(logConfig), nil

	}

	log.Printf("ResolveProcess: %v", actv)
	return nil, fmt.Errorf("'%s' activity processor not available.", actv.TxnType)
}
