package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

type AdjustmentActivityProcessor struct {
	logger *logger.Logger
}

func NewAdjustmentActivityProcessor(logConfig *logger.Config) AdjustmentActivityProcessor {
	plog := logConfig.For("processor.adjustment")
	return AdjustmentActivityProcessor{logger: plog}
}

// ensures CashActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*AdjustmentActivityProcessor)(nil)

func (p AdjustmentActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	p.logger.Debug("Process", "Value", actv.SentAmount)
	// update the cash lot
	_, err := lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.SentSymbol, actv.SentAmount.Mul(decimal.NewFromFloat(-1.0)))
	if err != nil {
		return nil, err
	}
	pr.Value = actv.SentAmount
	p.logger.Debug("Process", "Value", pr.Value)

	return pr, nil
}
