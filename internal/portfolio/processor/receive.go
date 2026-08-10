package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

type ReceiveActivityProcessor struct {
	logger *logger.Logger
}

func NewReceiveActivityProcessor(logConfig *logger.Config) ReceiveActivityProcessor {
	plog := logConfig.For("processor.receive")
	return ReceiveActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*ReceiveActivityProcessor)(nil)

func (p ReceiveActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)

	pr := NewProcessResult()
	lots, _, success := lm.MatchTransfer(ctx, actv)
	p.logger.Debug("Process", "Match", success, "lots", lots)
	if success {
		value := decimal.Zero
		for _, lot := range lots {
			// Ceate the lot of the asset
			lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, lot.Qty, lot.CostValue)
			value = value.Add(lot.CostValue)
		}
		pr.Value = value
		p.logger.Debug("Process", "Value", value)
	}

	return pr, nil
}
