package processor

import (
	"context"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

type CashActivityProcessor struct {
	logger *logger.Logger
}

func NewCashActivityProcessor(logConfig *logger.Config) CashActivityProcessor {
	plog := logConfig.For("processor.cash")
	return CashActivityProcessor{logger: plog}
}

// ensures CashActivityProcessor implements ActivityProcessor at compile time
var _ ActivityProcessor = (*CashActivityProcessor)(nil)

func (p CashActivityProcessor) Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error) {

	p.logger.Debug("Process")
	newctx := logger.WithContext(ctx, p.logger)
	pr := NewProcessResult()

	var err error
	if actv.TxnType == domain.ActivityTypeWithdraw {
		_, err = lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.SentSymbol, actv.SentAmount)
	} else {
		_, err = lm.UpdateCashLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount)
	}

	if err != nil {
		return nil, err
	}

	if strings.Compare(string(actv.TxnType), string(domain.ActivityTypeDeposit)) == 0 ||
		strings.Compare(string(actv.TxnType), string(domain.ActivityTypeWithdraw)) == 0 {

		p.logger.Debug("Process")

		_, err = lm.UpdateBankLot(newctx, actv)
		if err != nil {
			return nil, err
		}
	}
	pr.Value = actv.RcvAmount
	p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	return pr, nil
}
