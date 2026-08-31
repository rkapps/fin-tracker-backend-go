package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type IncomeActivityProcessor struct {
	logger *logger.Logger
}

func NewIncomeActivityProcessor(logConfig *logger.Config) IncomeActivityProcessor {
	plog := logConfig.For("processor.income")
	return IncomeActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*IncomeActivityProcessor)(nil)

func (p IncomeActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	newctx := logger.WithContext(ctx, p.logger)
	p.logger.Debug("Process")

	pr := NewProcessResult()

	// update the cash lot
	lot, err := lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount)
	if err != nil {
		return nil, err
	}

	err = lm.CreateGLIncome(newctx, lot, actv)
	if err != nil {
		return nil, err
	}

	pr.Value = actv.SentAmount
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
