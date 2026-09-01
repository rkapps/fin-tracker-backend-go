package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type DisposalActivityProcessor struct {
	logger *logger.Logger
}

func NewDisposalActivityProcessor(logConfig *logger.Config) DisposalActivityProcessor {
	plog := logConfig.For("processor.disposal")
	return DisposalActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*DisposalActivityProcessor)(nil)

func (p DisposalActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	// Reduce the lot of the asset and get the costvalue for the gl
	touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)
	// update the cash lot
	_, _ = lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount)

	_ = lm.UpdateFeeLot(ctx, actv)

	gl := lm.CreateGLDisposal(newctx, touched, actv)

	actv.GlAmount = gl
	pr.Value = actv.RcvAmount
	p.logger.Debug("Process", "CostValue", value, "RcvValue", actv.RcvAmount)

	return pr, nil
}
