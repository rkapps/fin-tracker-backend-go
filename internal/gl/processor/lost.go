package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type LostActivityProcessor struct {
	logger *logger.Logger
}

func NewLostActivityProcessor(logConfig *logger.Config) LostActivityProcessor {
	plog := logConfig.For("processor.lost")
	return LostActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*LostActivityProcessor)(nil)

func (p LostActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	// Reduce the lot of the asset and get the costvalue for the gl
	_, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)
	pr.Value = value
	p.logger.Debug("Process", "CostValue", value, "RcvValue", actv.RcvAmount)

	return pr, nil
}
