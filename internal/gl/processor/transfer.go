package processor

import (
	"context"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type TransferActivityProcessor struct {
	logger *logger.Logger
}

func NewTransferActivityProcessor(logConfig *logger.Config) TransferActivityProcessor {
	plog := logConfig.For("processor.transfer")
	return TransferActivityProcessor{logger: plog}
}

// ensures AquisitionActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*TransferActivityProcessor)(nil)

func (p TransferActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	if strings.Compare(actv.RcvSymbol, "USD") == 0 {
		lm.UpdateCashLot(newctx, actv, actv.SentAccountID, actv.SentSymbol, actv.SentAmount.Neg())
		lm.UpdateCashLot(newctx, actv, actv.RcvAccountID, actv.RcvSymbol, actv.SentAmount)
	} else {

		// Reduce the lot of the asset and get the costvalue for the gl
		touched, value, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)

		for _, lot := range touched {
			if strings.Compare(actv.ID, "eec7a867a3a1bee054aebf93a5fcc83da5c05a7a1e5ff6b6354037805135a9d1") == 0 {
				p.logger.Debug("Process", "Amount", lot.Amount, "Value", lot.CostValue)
			}
			lm.CreateAssetLot(newctx, actv, actv.RcvAccountID, actv.RcvSymbol, lot.Amount, lot.CostValue)
		}
		pr.Value = value
	}
	return pr, nil
}
