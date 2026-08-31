package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type SendActivityProcessor struct {
	logger *logger.Logger
}

func NewSendActivityProcessor(logConfig *logger.Config) SendActivityProcessor {
	plog := logConfig.For("processor.send")
	return SendActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*SendActivityProcessor)(nil)

func (p SendActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	// Reduce the lot of the asset and get the costvalue for the gl
	touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)
	p.logger.Debug("Process", "Touched", touched, "Value", value)

	// only store on send, not for stakefee etc...
	// if actv.TxnType == domain.ActivityTypeSend {
	lm.StoreTransfer(ctx, actv, touched)
	// }

	lm.UpdateFeeLot(ctx, actv)

	pr.Value = value
	// p.logger.Debug("Process", "Touched", pr.Lots)
	// p.logger.Debug("Process", "CostValue", value, "RcvValue", actv.RcvAmount)

	return pr, nil
}
