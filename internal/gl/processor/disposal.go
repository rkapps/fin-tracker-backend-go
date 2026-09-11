package processor

import (
	"context"
	"fmt"

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

	// validate if sentsymbol is a currency.
	if ClassifyAsset(actv.RcvSymbol) != AssetClassCash {
		p.logger.Error("Process", "Actv", actv.ID, "Account", actv.AccountID)
		return nil, fmt.Errorf("'%s' is not a currency", actv.RcvSymbol)
	}

	// Reduce the lot of the asset and get the costvalue for the gl
	touched, _, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)

	// // get the price and create the gl
	// price := actv.RcvAmount.Div(actv.SentAmount)

	// net proceeds actually landing in cash = gross RcvAmount - Fee
	netProceeds := actv.RcvAmount.Sub(actv.Fee)
	price := netProceeds.Div(actv.SentAmount)

	gl := lm.CreateGLDisposal(newctx, touched, actv, price)

	lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.RcvSymbol, netProceeds.Neg())

	actv.GlAmount = gl
	pr.Value = netProceeds
	p.logger.Debug("Process", "Value", actv.RcvAmount)

	return pr, nil
}
