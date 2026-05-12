package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type AquisitionActivityProcessor struct {
	logger *logger.Logger
}

func NewAcquisitionActivityProcessor() AquisitionActivityProcessor {
	logger := logger.New()
	plog := logger.For("processor.acquisition")
	return AquisitionActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*AquisitionActivityProcessor)(nil)

func (p AquisitionActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	ctx = logger.WithContext(ctx, p.logger)

	// Create the lot of the asset
	lot := lm.CreateAssetLot(ctx, actv, actv.RcvSymbol, actv.RcvQuantity, actv.RcvAmount)

	pr := NewProcessResult()
	pr.appendLot(lot)

	// update the cash lot
	clot, _ := lm.UpdateCashLot(ctx, actv, actv.AccountID, actv.SentSymbol)
	pr.appendLot(clot)

	pr.Value = actv.SentAmount

	return pr, nil
}
