package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type CashActivityProcessor struct {
	logger *logger.Logger
}

func NewCashActivityProcessor() CashActivityProcessor {
	logger := logger.New()
	plog := logger.For("processor.cash")
	return CashActivityProcessor{logger: plog}
}

// ensures CashActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*CashActivityProcessor)(nil)

func (p CashActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	ctx = logger.WithContext(ctx, p.logger)
	p.logger.Debug("Process")

	lot, err := lm.UpdateCashLot(ctx, actv, actv.AccountID, actv.RcvSymbol)
	if err != nil {
		return nil, err
	}
	pr := NewProcessResult()
	pr.appendLot(lot)
	pr.Value = actv.RcvAmount
	// p.logger.Trace("Process", len(pr.Lots))
	return pr, nil
}
