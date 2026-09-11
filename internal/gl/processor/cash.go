package processor

import (
	"context"
	"fmt"

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

		// validate if rcvsymbol is a currency. sentsymbol can be gusd.
		if ClassifyAsset(actv.RcvSymbol) != AssetClassCash {
			return nil, fmt.Errorf("Receive '%s' is not a currency", actv.RcvSymbol)
		}

		if ClassifyAsset(actv.SentSymbol) == AssetClassStablecoin {
			_, _, err = lm.ReduceLotQty(newctx, actv, actv.SentAmount)
		} else {
			_, err = lm.UpdateCashLot(newctx, actv, actv.SentAccountID, actv.SentSymbol, actv.SentAmount)
		}
		if err != nil {
			return nil, err
		}

		_, err = lm.UpdateCashLot(newctx, actv, actv.RcvAccountID, actv.RcvSymbol, actv.RcvAmount.Neg())
		if err != nil {
			return nil, err
		}

		pr.Value = actv.SentAmount
		p.logger.Debug("Process", "SentValue", actv.SentAmount)

	} else {

		// validate if rcvsymbol and sentsymbol is a currency.
		if ClassifyAsset(actv.RcvSymbol) != AssetClassCash {
			p.logger.Error("Process", "Actv", actv.ID, "Account", actv.AccountID)
			return nil, fmt.Errorf("Receive '%s' is not a currency", actv.RcvSymbol)
		}

		_, err = lm.UpdateCashLot(newctx, actv, actv.RcvAccountID, actv.RcvSymbol, actv.RcvAmount.Neg())
		if err != nil {
			return nil, err
		}

		// for rollover there is no sent symbol
		if len(actv.SentSymbol) > 0 {

			// validate if rcvsymbol and sentsymbol is a currency.
			if ClassifyAsset(actv.SentSymbol) != AssetClassCash {
				p.logger.Error("Process", "Actv", actv.ID, "Account", actv.AccountID)
				return nil, fmt.Errorf("Sent '%s' is not a currency", actv.SentSymbol)
			}

			_, err = lm.UpdateCashLot(newctx, actv, actv.SentAccountID, actv.SentSymbol, actv.SentAmount)
			if err != nil {
				return nil, err
			}

		}

		pr.Value = actv.RcvAmount
		p.logger.Debug("Process", "RcvValue", actv.RcvAmount)

	}

	// if strings.Compare(string(actv.TxnType), string(domain.ActivityTypeDeposit)) == 0 ||
	// 	strings.Compare(string(actv.TxnType), string(domain.ActivityTypeWithdraw)) == 0 {

	// 	_, err = lm.UpdateBankLot(newctx, actv)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// }

	return pr, nil
}
