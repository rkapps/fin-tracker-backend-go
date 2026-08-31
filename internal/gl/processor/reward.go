package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type RewardActivityProcessor struct {
	logger *logger.Logger
}

func NewRewardActivityProcessor(logConfig *logger.Config) RewardActivityProcessor {
	plog := logConfig.For("processor.reward")
	return RewardActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*RewardActivityProcessor)(nil)

func (p RewardActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	newctx := logger.WithContext(ctx, p.logger)
	p.logger.Debug("Process")

	pr := NewProcessResult()

	// Create the lot of the asset
	lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, actv.SentAmount.Add(actv.Fee))

	// err = lm.CreateGLIncome(newctx, lot, actv)
	// if err != nil {
	// 	return nil, err
	// }

	pr.Value = actv.RcvAmount
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
