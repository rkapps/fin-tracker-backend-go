package processor

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type AquisitionActivityProcessor struct {
	logger *logger.Logger
}

func NewAcquisitionActivityProcessor(logConfig *logger.Config) AquisitionActivityProcessor {
	plog := logConfig.For("processor.acquisition")
	return AquisitionActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*AquisitionActivityProcessor)(nil)

func (p AquisitionActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)

	pr := NewProcessResult()

	// validate if sentsymbol is a currency.
	if ClassifyAsset(actv.SentSymbol) != AssetClassCash {
		p.logger.Error("Process", "Actv", actv.ID, "Account", actv.AccountID)
		return nil, fmt.Errorf("'%s' is not a currency", actv.SentSymbol)
	}
	if actv.SentAmount.IsZero() {
		return nil, fmt.Errorf("Sent amount is zero")
	}

	total := actv.SentAmount.Add(actv.Fee)

	// Create the lot of the asset
	lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, total)

	// update the cash lot for the sent account
	_, err := lm.UpdateCashLot(newctx, actv, actv.SentAccountID, actv.SentSymbol, total)
	if err != nil {
		return nil, err
	}

	pr.Value = total
	p.logger.Debug("Process", "Value", total)

	return pr, nil
}
