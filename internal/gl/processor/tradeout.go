package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type TradeOutActivityProcessor struct {
	logger *logger.Logger
}

func NewTradeOutActivityProcessor(logConfig *logger.Config) TradeOutActivityProcessor {
	plog := logConfig.For("processor.Tradeout")
	return TradeOutActivityProcessor{logger: plog}
}

// ensures TradeOutActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*TradeOutActivityProcessor)(nil)

func (p TradeOutActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	// Create the lot of the asset
	lot := lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, actv.SentAmount.Add(actv.Fee))
	if lot != nil {
		pr.Value = lot.CostValue
	}
	return pr, nil
}
