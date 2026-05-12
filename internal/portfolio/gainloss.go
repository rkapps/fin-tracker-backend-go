package portfolio

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/portfolio/processor"
	"github.com/shopspring/decimal"
)

// GainLoss implements processor.LotManager
var _ processor.LotManager = (*GainLoss)(nil) // compile time check

// GainLossService is the interface for running gain/loss computation.
type GainLossService interface {
	Run(ctx context.Context, accts []domain.Account, actvs []domain.Activity) (GainLossResult, error)
}

// GainLoss owns all state for a single GL run.
// Created fresh per user per run — never reused.
type GainLoss struct {
	acctsm            map[string]domain.Account
	lotsMap           map[string][]*domain.ActivityLot // keyed by accountID
	acctLotSeqMap     map[string]int                   // lot seq counter per account
	lotMatchingMethod domain.LotMatchingMethod
	logger            *logger.Logger
	simulate          bool
}

// GainLossResult is the output of one GL run.
type GainLossResult struct {
	Lots      []*domain.ActivityLot
	GLEntries []domain.GLEntry
}

func (gr *GainLossResult) appendLots(lots []*domain.ActivityLot) {
	gr.Lots = append(gr.Lots, lots...)
}

// NewGainLoss creates a fresh GainLoss for one run.
func NewGainLoss(accts []*domain.Account, method domain.LotMatchingMethod, simulate bool) *GainLoss {

	glogger := logger.New()
	plog := glogger.For("gainloss")

	acctsm := make(map[string]domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = *acct
	}

	return &GainLoss{
		acctsm:            acctsm,
		lotsMap:           make(map[string][]*domain.ActivityLot),
		acctLotSeqMap:     make(map[string]int),
		lotMatchingMethod: method,
		logger:            plog,
		simulate:          simulate,
	}
}

// Run processes all activities and produces lots and GL entries.
func (gl *GainLoss) Run(ctx context.Context, actvs []*domain.Activity) (GainLossResult, error) {

	newctx := logger.WithContext(context.Background(), gl.logger)

	gl.logger.Info("gainLossRun", "LotMatchingMethod", gl.lotMatchingMethod)
	// TODO: sort activities chronologically before processing
	// TODO: range over activities
	//       resolve processor by activity type
	//       call processor.Process(ctx, activity, gl)
	//       accumulate lots and GL entries
	gr := &GainLossResult{}

	for _, actv := range actvs {
		gl.logger.Debug("")
		gl.logger.Debug("---gainLossRun---", "Activity", actv.Debug())

		processor, err := processor.ResolveProcessor(*actv, gl)
		if err != nil {
			gl.logger.Error("gainLossRun", "Error", err)
			continue
		}
		pr, err := processor.Process(newctx, actv, gl)
		if err != nil {
			gl.logger.Error("gainLossRun", "Error", err)
			continue
		}

		// update the lots
		gr.appendLots(pr.Lots)

		// update activity
		actv.Value = pr.Value
		actv.RcvBalance = gl.getOpenBalance(actv.AccountID, actv.RcvSymbol)
		actv.SentBalance = gl.getOpenBalance(actv.AccountID, actv.SentSymbol)
		gl.logger.Debug("gainLossRun", "RcvBalance", fmt.Sprintf("%s %v", actv.RcvSymbol, actv.RcvBalance))
		gl.logger.Debug("gainLossRun", "SentBalance", fmt.Sprintf("%s  %v", actv.SentSymbol, actv.SentBalance))
		gl.logger.Trace("gainLossRun", "Result", len(pr.Lots))
	}

	return *gr, nil
}

// lot creation
func (gl *GainLoss) CreateAssetLot(ctx context.Context, actv *domain.Activity, symbol string, qty decimal.Decimal, value decimal.Decimal) *domain.ActivityLot {

	nlot := domain.NewLotFromActivity(*actv)
	nlot.LotSeq = gl.NextLotSeq(ctx, actv.AccountID)
	nlot.ID = fmt.Sprintf("%s-%d", actv.AccountID, nlot.LotSeq)
	nlot.Symbol = symbol
	nlot.OrigQty = qty
	nlot.Qty = qty
	nlot.CostValue = value
	if !nlot.Qty.IsZero() {
		nlot.Cost = nlot.CostValue.Div(nlot.Qty)
	}

	key := getAccountSymbolKey(actv.AccountID, symbol)
	lots := gl.lotsMap[key]
	if len(lots) == 0 {
		lots = []*domain.ActivityLot{}
	}
	lots = append(lots, nlot)
	gl.lotsMap[key] = lots
	gl.logger.Debug("CreateAssetLot", "Asset", fmt.Sprintf("%s Qty: %v-%v", nlot.Symbol, nlot.Qty, nlot.CostValue))

	return nlot
}

// lot consumption
func (gl *GainLoss) ReduceLotQty(ctx context.Context, actv *domain.Activity) error {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	acct := gl.acctsm[actv.AccountID]
	logger.Debug("ReduceLotQty", "Symbol", fmt.Sprintf("%s-%v", actv.SentSymbol, actv.SentQuantity))
	if len(acct.ID) == 0 {
		return fmt.Errorf("account does not exist for %s", actv.AccountID)
	}
	lots := gl.MatchLots(ctx, acct, actv.SentSymbol, actv.SentAmount)
	for _, lot := range lots {
		logger.Debug("ReduceLotQty", "lot", lot.Debug())
	}
	return nil
}

func (gl *GainLoss) CloseLot(ctx context.Context, lot *domain.ActivityLot) error {
	return nil
}

// MatchLots returns lots in the correct order for disposal.
// Method is resolved per account — crypto uses HIFO, securities use FIFO.
func (gl *GainLoss) MatchLots(ctx context.Context, account domain.Account, symbol string, qty decimal.Decimal) []*domain.ActivityLot {
	method := gl.resolveLotMatchingMethod(account)
	lots := gl.GetOpenLots(ctx, account, symbol)

	gl.logger.Debug("MatchLots",
		"openLots", len(lots),
	)
	gl.sortLots(method, lots) // ← no return needed
	return lots
}

// lot querying
func (gl *GainLoss) GetOpenLots(ctx context.Context, acct domain.Account, symbol string) []*domain.ActivityLot {

	var lots []*domain.ActivityLot
	var ulots []*domain.ActivityLot

	if len(acct.ID) > 0 {
		key := getAccountSymbolKey(acct.ID, symbol)

		lots = gl.lotsMap[key]
		for _, lot := range lots {
			if strings.Compare(string(lot.Status), string(domain.LotStatusOpen)) == 0 {
				ulots = append(ulots, lot)
			}
		}
	}

	return ulots
}

// seq management
func (gl *GainLoss) NextLotSeq(ctx context.Context, accountID string) int {
	gl.acctLotSeqMap[accountID]++
	return gl.acctLotSeqMap[accountID]
}

// GL entries
func (gl *GainLoss) CreateGLEntry(ctx context.Context, lot *domain.ActivityLot, activity *domain.Activity, value decimal.Decimal) domain.GLEntry {
	return domain.GLEntry{}
}

func (gl GainLoss) UpdateCashLot(ctx context.Context, actv *domain.Activity, acctId string, symbol string) (*domain.ActivityLot, error) {

	var lot *domain.ActivityLot
	key := getAccountSymbolKey(acctId, symbol)
	lots := gl.lotsMap[key]
	if len(lots) == 0 {
		lot = gl.CreateAssetLot(ctx, actv, symbol, decimal.Zero, decimal.Zero)
		lots = []*domain.ActivityLot{}
		lots = append(lots, lot)
	}

	lot = lots[0]
	switch actv.TxnType {
	case domain.ActivityTypeBuy:
		lot.Qty = lot.Qty.Sub(actv.RcvAmount)
		lot.CostValue = lot.CostValue.Sub(actv.RcvAmount)
	default:
		lot.Qty = lot.Qty.Add(actv.RcvAmount)
		lot.CostValue = lot.CostValue.Add(actv.RcvAmount)
	}

	gl.logger.Debug("CreateAssetLot", "Cash", fmt.Sprintf("%s  Qty: %v-%v", lot.Symbol, lot.Qty, lot.CostValue))
	lot.Cost = lot.CostValue.Div(lot.Qty)

	gl.lotsMap[key] = lots
	return lot, nil
}

// resolveLotMatchingMethod returns the correct method for an account.
// Account level overrides user preference. Falls back to category default.
func (gl *GainLoss) resolveLotMatchingMethod(account domain.Account) domain.LotMatchingMethod {
	// account level override — user explicitly set it
	if account.LotMatchingMethod != "" {
		return account.LotMatchingMethod
	}

	// user global preference
	if gl.lotMatchingMethod != "" {
		return gl.lotMatchingMethod
	}

	// category default
	return defaultLotMatchingMethod(account.Category)
}

func (gl *GainLoss) getOpenBalance(acctId string, symbol string) decimal.Decimal {

	balance := decimal.Zero
	key := getAccountSymbolKey(acctId, symbol)
	lots := gl.lotsMap[key]
	for _, lot := range lots {
		gl.logger.Trace("getOpenBalance", "qty", lot.Qty, "costvalue", lot.CostValue)
		if lot.Status == domain.LotStatusOpen {
			balance = balance.Add(lot.CostValue)
		}
	}
	return balance
}

func (gl *GainLoss) sortLots(method domain.LotMatchingMethod, lots []*domain.ActivityLot) {
	switch method {
	case domain.LotMatchingHIFO:
		sort.SliceStable(lots, func(i, j int) bool {
			return lots[i].Cost.GreaterThan(lots[j].Cost)
		})
	default:
		sort.SliceStable(lots, func(i, j int) bool {
			return lots[i].Date.Before(*lots[j].Date)
		})
	}
}
