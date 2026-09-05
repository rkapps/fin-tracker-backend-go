package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type TradeInActivityProcessor struct {
	logger *logger.Logger
}

func NewTradeInActivityProcessor(logConfig *logger.Config) TradeInActivityProcessor {
	plog := logConfig.For("processor.tradein")
	return TradeInActivityProcessor{logger: plog}
}

// ensures TradeInActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*TradeInActivityProcessor)(nil)

func (p TradeInActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	// Reduce the lot of the asset and get the costvalue for the gl
	touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)
	// lm.CreateGLDisposal(newctx, touched, actv)
	gl := lm.CreateGLDisposal(newctx, touched, actv)
	actv.GlAmount = gl
	pr.Value = value

	return pr, nil
}
