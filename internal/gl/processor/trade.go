package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/providers"
)

type TradeActivityProcessor struct {
	logger *logger.Logger
}

func NewTradeActivityProcessor(logConfig *logger.Config) TradeActivityProcessor {
	plog := logConfig.For("processor.trade")
	return TradeActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*TradeActivityProcessor)(nil)

func (p TradeActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	if providers.IsCurrency(actv.RcvSymbol) {

		// Reduce the lot of the asset and get the costvalue for the gl
		touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)

		lm.CreateGLDisposal(newctx, touched, actv)
		// actv.GlAmount = gl

		// update cash lot --- this adds the amount for trade
		lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount)

		pr.Value = actv.RcvAmount
		p.logger.Debug("Process", "CostValue", value, "RcvValue", actv.RcvAmount)

	} else if providers.IsCurrency(actv.SentSymbol) {

		// update cash lot --- send negative of sentamount
		lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.SentSymbol, actv.SentAmount.Neg())

		// Create the lot of the asset
		lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, actv.SentAmount)
		pr.Value = actv.SentAmount
	} else {

		// Reduce the lot of the asset and get the costvalue for the gl
		touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)

		lm.CreateGLDisposal(newctx, touched, actv)
		// actv.GlAmount = gl
		pr.Value = value

		// for _, lot := range touched {
		// Create the lot of the asset
		lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, value)

	}
	lm.UpdateFeeLot(ctx, actv)

	return pr, nil
}
