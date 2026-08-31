package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type DelegationActivityProcessor struct {
	logger *logger.Logger
}

func NewDelegationActivityProcessor(logConfig *logger.Config) DelegationActivityProcessor {
	plog := logConfig.For("processor.delegation")
	return DelegationActivityProcessor{logger: plog}
}

// ensures CashActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*DelegationActivityProcessor)(nil)

func (p DelegationActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	// newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()
	lm.UpdateFeeLot(ctx, actv)
	return pr, nil
}
