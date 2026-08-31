package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type FeeActivityProcessor struct {
	logger *logger.Logger
}

func NewFeeActivityProcessor(logConfig *logger.Config) FeeActivityProcessor {
	plog := logConfig.For("processor.fee")
	return FeeActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*FeeActivityProcessor)(nil)

func (p FeeActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)

	pr := NewProcessResult()
	value := lm.UpdateFeeLot(newctx, actv)
	pr.Value = value
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
