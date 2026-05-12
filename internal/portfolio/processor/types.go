package processor

import (
	"context"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
)

// LotManager is the interface processors use to interact with the GL engine.
// Implemented by GainLoss — processor never imports GainLoss directly.
type LotManager interface {
	CloseLot(ctx context.Context, lot *domain.ActivityLot) error
	CreateGLEntry(ctx context.Context, lot *domain.ActivityLot, activity *domain.Activity, qty decimal.Decimal) domain.GLEntry
	CreateAssetLot(ctx context.Context, actv *domain.Activity, symbol string, qty decimal.Decimal, value decimal.Decimal) *domain.ActivityLot

	MatchOpenLots(ctx context.Context, account domain.Account, symbol string) []*domain.ActivityLot
	NextLotSeq(ctx context.Context, accountID string) int
	ReduceLotQty(ctx context.Context, actv *domain.Activity) (decimal.Decimal, error)
	UpdateCashLot(ctx context.Context, activity *domain.Activity, acctId string, symbol string, amount decimal.Decimal) (*domain.ActivityLot, error)
}

type ActivityProcessor interface {
	Process(ctx context.Context, actv *domain.Activity, lm LotManager) (*ProcessorResult, error)
}

type ProcessorResult struct {
	Value decimal.Decimal
	Lots  []*domain.ActivityLot
	Gls   []domain.GLEntry
}

func NewProcessResult() *ProcessorResult {

	return &ProcessorResult{}
}

func (pr *ProcessorResult) appendLot(lot *domain.ActivityLot) {
	pr.Lots = append(pr.Lots, lot)
}
