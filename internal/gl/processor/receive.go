package processor

import (
	"context"
	"fmt"

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
	lots, _, success := lm.MatchTransfer(newctx, actv)
	// p.logger.Info("Process", "Match", success, "lots", lots)
	if success {
		value := decimal.Zero
		lotAccountId := ""
		for _, lot := range lots {
			// Ceate the lot of the asset
			lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, lot.Amount, lot.CostValue)
			value = value.Add(lot.CostValue)
			lotAccountId = lot.AccountID
		}
		pr.Value = value

		// If sentSymbol is blank, set it from the lot and set the sentAmount to the value
		if len(actv.SentSymbol) == 0 {
			// actv.SentSymbol = actv.RcvSymbol
			// actv.SentAmount = value
		}
		actv.Notes = fmt.Sprintf("From: %s", lotAccountId)
		p.logger.Debug("Process", "Value", value)
	} else {
		lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, actv.SentAmount)
		actv.Orphan = true
	}

	return pr, nil
}
