package processor

import (
	"context"
	"fmt"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
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

	// validate if rcvsymbol and sentsymbol is a currency.
	if ClassifyAsset(actv.SentSymbol) == AssetClassCash {
		return nil, fmt.Errorf("Sent '%s' should not be a currency", actv.SentSymbol)
	}

	// validate if rcvsymbol and sentsymbol is a currency.
	if ClassifyAsset(actv.RcvSymbol) == AssetClassCash {
		return nil, fmt.Errorf("Receive '%s' should not be a currency", actv.RcvSymbol)
	}

	// Reduce the lot of the asset and get the costvalue for the gl
	touched, _, _ := lm.ReduceLotQty(newctx, actv, actv.SentAmount)

	var price decimal.Decimal
	if ClassifyAsset(actv.RcvSymbol) == AssetClassStablecoin {
		pr.Value = actv.RcvAmount
		price = actv.RcvAmount.Div(actv.SentAmount)

	} else if ClassifyAsset(actv.SentSymbol) == AssetClassStablecoin {
		pr.Value = actv.SentAmount
		price = decimal.NewFromFloat(1.0)
	} else {
		pr.Value = actv.SentAmount.Mul(actv.SentPrice)
		price = actv.SentAmount.Mul(actv.SentPrice)
		if !price.IsZero() {
			price = actv.RcvAmount.Div(price)
		}
	}

	// create disposal lot
	gl := lm.CreateGLDisposal(newctx, touched, actv, price)

	actv.GlAmount = gl

	// Create the lot of the asset
	lm.CreateAssetLot(newctx, actv, actv.AccountID, actv.RcvSymbol, actv.RcvAmount, pr.Value)

	lm.UpdateFeeLot(ctx, actv)

	return pr, nil
}
