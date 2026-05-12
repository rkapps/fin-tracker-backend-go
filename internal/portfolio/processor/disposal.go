package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type DisposalActivityProcessor struct {
	logger *logger.Logger
}

func NewDisposalActivityProcessor() DisposalActivityProcessor {
	logger := logger.New()
	plog := logger.For("processor.disposal")
	return DisposalActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*DisposalActivityProcessor)(nil)

func (p DisposalActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)

	// Create the lot of the asset
	// lot := lm.CreateAssetLot(actv, actv.RcvSymbol, actv.RcvQuantity, actv.RcvAmount)
	lm.ReduceLotQty(newctx, actv)
	pr := NewProcessResult()
	// pr.appendLot(lot)

	// // update the cash lot
	// clot, _ := lm.UpdateCashLot(actv, actv.AccountID, actv.SentSymbol)
	// pr.appendLot(clot)

	// pr.Value = actv.SentAmount

	return pr, nil
}
