package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type AquisitionActivityProcessor struct {
	logger *logger.Logger
}

func NewAcquisitionActivityProcessor(logConfig *logger.Config) AquisitionActivityProcessor {
	plog := logConfig.For("processor.acquisition")
	return AquisitionActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*AquisitionActivityProcessor)(nil)

func (p AquisitionActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)

	pr := NewProcessResult()

	// Create the lot of the asset
	lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvQuantity, actv.RcvAmount)

	p.logger.Debug("Process")
	// update the cash lot
	_, err := lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.SentSymbol, actv.SentAmount)
	if err != nil {
		return nil, err
	}

	pr.Value = actv.SentAmount
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
