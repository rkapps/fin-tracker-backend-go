package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
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

	// update the cash lot
	_, err := lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.SentSymbol, actv.SentAmount.Mul(decimal.NewFromFloat(-1.0)))
	if err != nil {
		return nil, err
	}

	pr.Value = actv.SentAmount
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
