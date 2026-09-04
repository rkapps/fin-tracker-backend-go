package gl

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/core"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/gl/processor"
	"github.com/rkapps/fin-tracker-backend-go/internal/utils"
	"github.com/shopspring/decimal"
)

const (
	MAX_DECIMALS = 8
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
	user               domain.User
	acctsm             map[string]domain.Account
	acctLotSeqMap      map[string]int // lot seq counter per account
	costBasisMethod    domain.CostBasisMethod
	GLEntries          []*domain.GLEntry
	logConfig          *logger.Config
	logger             *logger.Logger
	lotsMap            map[string][]*domain.ActivityLot // keyed by accountID
	simulate           bool
	transferActivities map[string]*domain.Activity      // keyed by activityID
	transferLots       map[string][]*domain.ActivityLot // keyed by activityID
	debug              bool
}

// GainLossResult is the output of one GL run.
type GainLossResult struct {
	Actvs     []*domain.Activity
	GLEntries []*domain.GLEntry
	Lots      []*domain.ActivityLot
}

// NewGainLoss creates a fresh GainLoss for one run.
func NewGainLoss(user domain.User, accts []*domain.Account, simulate bool, logConfig *logger.Config) *GainLoss {

	plog := logConfig.For("gainloss")

	acctsm := make(map[string]domain.Account)
	for _, acct := range accts {
		acctsm[acct.ID] = *acct
	}

	return &GainLoss{
		acctsm:             acctsm,
		acctLotSeqMap:      make(map[string]int),
		lotsMap:            make(map[string][]*domain.ActivityLot),
		logger:             plog,
		logConfig:          logConfig,
		simulate:           simulate,
		transferActivities: make(map[string]*domain.Activity),
		transferLots:       make(map[string][]*domain.ActivityLot),
	}
}

// Run processes all activities and produces lots and GL entries.
func (gl *GainLoss) Run(ctx context.Context, actvs []*domain.Activity) (GainLossResult, error) {

	newctx := logger.WithContext(context.Background(), gl.logger)

	gl.logger.Debug("---Run---", "CostBasisMethod", gl.costBasisMethod)
	// TODO: sort activities chronologically before processing
	// TODO: range over activities
	//       resolve processor by activity type
	//       call processor.Process(ctx, activity, gl)
	//       accumulate lots and GL entries
	gr := &GainLossResult{}
	uactvs := []*domain.Activity{}

	sort.Slice(actvs, func(i, j int) bool {
		return actvs[i].Date.Before(actvs[j].Date)
	})

	for i, actv := range actvs {
		if i > 1000 {
			// break
		}

		gl.debug = false
		if //strings.Compare(actv.AccountID, "Solana-Fa8jM") == 0 ||
		strings.Compare(actv.ID, "0xb16d3d72068a6ce015c5639987134249fc231eb8aaa2172533c33350fe9465d1") == 0 {
			// gl.debug = true
		}
		if gl.debug {
			gl.logger.Info("---Run---", "Activity", actv.Debug(), "Date", actv.Date)
			gl.logger.Info("---Run---", "RcvAccount", actv.RcvAccountID, "Amount", fmt.Sprintf("%s-%v", actv.RcvSymbol, actv.RcvAmount))
			gl.logger.Info("---Run---", "SentAccount", actv.SentAccountID, "Amount", fmt.Sprintf("%s-%v", actv.SentSymbol, actv.SentAmount))
		}

		processor, err := ResolveProcessor(*actv, gl, gl.logConfig)
		if err != nil {
			// gl.logger.Error()
			gl.logger.Error("Run", "Error", err)
			continue
		}
		pr, err := processor.Process(newctx, actv, gl)
		if err != nil {
			gl.logger.Error("Run", "Error", err)
			continue
		}

		//set orphan to false
		actv.Orphan = false
		// update activity
		actv.Value = pr.Value
		if gl.debug {
			gl.logger.Info("---Run---", "Value", actv.Value)
		}
		// log.Printf("Account: %s Fee: %s-%v", actv.RcvAccountID, actv.FeeCurrency, actv.Fee)
		// gl.UpdateCashLot(ctx, actv, actv.AccountID, actv.FeeCurrency, actv.Fee)

		gl.logger.Trace("RUn", "Lots", len(gl.lotsMap))
		actv.RcvBalance = gl.getOpenBalance(actv.RcvAccountID, actv.RcvSymbol)
		gl.logger.Trace("Run", "RcvBalance", fmt.Sprintf("%s %v", actv.RcvSymbol, actv.RcvBalance))
		actv.SentBalance = gl.getOpenBalance(actv.SentAccountID, actv.SentSymbol)
		gl.logger.Trace("Run", "SentBalance", fmt.Sprintf("%s  %v", actv.SentSymbol, actv.SentBalance))
		gl.logger.Trace("Run", "Result", len(pr.Lots))

		uactvs = append(uactvs, actv)
	}

	for _, actv := range gl.transferActivities {
		actv.Orphan = true
	}
	gr.Actvs = uactvs
	gr.GLEntries = gl.GLEntries
	gr.Lots = utils.FlattenMap(gl.lotsMap)
	gl.logger.Info("Run", "UpdatedActivities", len(gr.Actvs))

	return *gr, nil
}

// lot creation
func (gl *GainLoss) CreateAssetLot(ctx context.Context, actv *domain.Activity, acctId string, symbol string, qty decimal.Decimal, value decimal.Decimal) *domain.ActivityLot {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	nlot := domain.NewLotFromActivity(*actv)
	nlot.AccountID = acctId
	nlot.LotSeq = gl.NextLotSeq(ctx, nlot.AccountID)
	nlot.ID = fmt.Sprintf("%s-%d", nlot.AccountID, nlot.LotSeq)
	nlot.Symbol = symbol
	nlot.OrigAmount = qty
	nlot.Amount = qty.Round(MAX_DECIMALS)
	nlot.CostValue = value
	if !nlot.Amount.IsZero() {
		nlot.Cost = nlot.CostValue.Div(nlot.Amount).Round(MAX_DECIMALS)
	}
	nlot.CostValue = nlot.CostValue.Round(MAX_DECIMALS)

	key := getAccountSymbolKey(nlot.AccountID, symbol)
	lots := gl.lotsMap[key]
	if len(lots) == 0 {
		lots = []*domain.ActivityLot{}
	}
	lots = append(lots, nlot)
	gl.lotsMap[key] = lots
	logger.Debug("CreateAssetLot", "Asset", fmt.Sprintf("%s Qty: %v-%v", nlot.Symbol, nlot.Amount, nlot.CostValue))

	return nlot
}

func (gl *GainLoss) CloseLot(ctx context.Context, lot *domain.ActivityLot) error {
	return nil
}

func (gl *GainLoss) CreateGLDisposal(ctx context.Context, lots []*domain.ActivityLot, activity *domain.Activity) decimal.Decimal {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	tgainLoss := decimal.Zero
	acct := gl.acctsm[activity.AccountID]
	logger.Debug("CreateGLDisposal", "Amount", fmt.Sprintf("%v--%v", activity.RcvAmount, activity.SentAmount))

	for _, lot := range lots {
		price := activity.RcvAmount.Div(activity.SentAmount)
		proceeds := lot.Amount.Mul(price)
		gainLoss := proceeds.Sub(lot.CostValue)
		tgainLoss = tgainLoss.Add(gainLoss)

		txgainLoss := decimal.Zero
		if acct.TaxStatus == domain.TaxStatusTaxable {
			txgainLoss = gainLoss
		}

		term := classifyTerm(*lot.Date, activity.Date)

		detail := &domain.GLDisposalDetail{
			CostBasis:        lot.CostValue,
			CostBasisPerUnit: lot.Cost,
			Proceeds:         proceeds,
			ProceedsPerUnit:  price,
			GainLoss:         gainLoss,
			TaxableGainLoss:  txgainLoss,
			Term:             term,
			AcquiredDate:     *lot.Date,
		}

		glEntry := &domain.GLEntry{
			ID:         uuid.New().String(),
			UID:        lot.UID,
			AccountID:  lot.AccountID,
			ActivityID: activity.ID,
			LotID:      lot.ID,
			TxnType:    activity.TxnType,
			GLType:     domain.GLTypeDisposal,
			Symbol:     lot.Symbol,
			Quantity:   lot.Amount,
			Detail:     detail,
			TaxDate:    activity.Date,
		}
		gl.GLEntries = append(gl.GLEntries, glEntry)
		logger.Debug("CreateGlDisposal", "gainloss", gainLoss, "proceeds", proceeds, "costvalue", lot.CostValue)
	}

	logger.Debug("CreateGlDisposal", "tgainloss", tgainLoss)

	return tgainLoss
}
func (gl *GainLoss) CreateGLIncome(ctx context.Context, lot *domain.ActivityLot, activity *domain.Activity) error {

	logger := logger.FromContext(ctx) // ← gets processor's logger
	acct := gl.acctsm[activity.AccountID]

	txgainLoss := decimal.Zero
	gainLoss := activity.SentAmount
	if acct.TaxStatus == domain.TaxStatusTaxable {
		txgainLoss = gainLoss
	}
	detail := &domain.GLIncomeDetail{
		Income:        gainLoss,
		TaxableIncome: txgainLoss,
		ReceivedDate:  activity.Date,
	}

	glEntry := &domain.GLEntry{
		ID:         uuid.New().String(),
		UID:        lot.UID,
		AccountID:  lot.AccountID,
		ActivityID: activity.ID,
		LotID:      lot.ID,
		TxnType:    activity.TxnType,
		GLType:     domain.GLTypeIncome,
		Symbol:     activity.RcvSymbol,
		Quantity:   activity.RcvAmount,
		Detail:     detail,
		TaxDate:    activity.Date,
	}
	gl.GLEntries = append(gl.GLEntries, glEntry)
	logger.Debug("CreateGlIncome", "gainloss", gainLoss)
	return nil
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

// MatchOpenLots returns lots in the correct order for disposal.
// Method is resolved per account — crypto uses HIFO, securities use FIFO.
func (gl *GainLoss) MatchOpenLots(ctx context.Context, account domain.Account, symbol string) []*domain.ActivityLot {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	method := gl.resolveLotMatchingMethod(account)
	lots := gl.GetOpenLots(ctx, account, symbol)

	logger.Debug("MatchLots", "method", method, "openLots", len(lots))
	gl.sortLots(method, lots) // ← no return needed
	return lots
}

func (gl *GainLoss) MatchTransfer(ctx context.Context, actv *domain.Activity) ([]*domain.ActivityLot, *domain.Activity, bool) {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	for id, sentActv := range gl.transferActivities {
		if gl.debug {
			// if gl.matchTransfer(sentActv, actv) {
			logger.Debug("MatchLots", "id", id, "sactv", sentActv.RcvAccount, "ractv", actv.AccountID)
			logger.Debug("MatchLots", "sentSymbol", sentActv.SentSymbol, "rcvSymbol", actv.RcvSymbol)
			logger.Debug("MatchLots", "sentAmount", sentActv.SentAmount, "rcvAmount", actv.RcvAmount)
		}

		match := false
		// this matches the id from the import
		if (sentActv.RcvAccount == actv.AccountID && sentActv.SentSymbol == actv.RcvSymbol) ||
			(sentActv.SentSymbol == actv.RcvSymbol && core.AmountsMatch(sentActv.SentAmount, actv.RcvAmount)) {
			match = true
		}

		// polygon eth-weth
		if sentActv.SentSymbol == "ETH" && actv.RcvSymbol == "WETH" && core.AmountsMatch(sentActv.SentAmount, actv.RcvAmount) {
			match = true
		}

		// polygon usdc- usdc.e
		if sentActv.SentSymbol == "USDC.e" && actv.RcvSymbol == "USDC" && core.AmountsMatch(sentActv.SentAmount, actv.RcvAmount) {
			match = true
		}
		// polygon eth-weth
		if sentActv.SentSymbol == "USDC" && actv.RcvSymbol == "USDC.e" && core.AmountsMatch(sentActv.SentAmount, actv.RcvAmount) {
			match = true
		}

		// // polygon matic-pol
		// if sentActv.SentSymbol == "POL" && actv.RcvSymbol == "POL" && core.AmountsMatch(sentActv.SentAmount, actv.RcvAmount) {
		// 	match = true
		// }

		if match {
			lots := gl.transferLots[id]
			if gl.debug {
				logger.Info("MatchLots", "lots", lots)
			}
			delete(gl.transferLots, id)
			delete(gl.transferActivities, id)
			return lots, sentActv, true
		}
	}
	return nil, nil, false
}

// seq management
func (gl *GainLoss) NextLotSeq(ctx context.Context, accountID string) int {
	gl.acctLotSeqMap[accountID]++
	return gl.acctLotSeqMap[accountID]
}

// lot consumption
func (gl *GainLoss) ReduceLotQty(ctx context.Context, actv *domain.Activity, samount decimal.Decimal) ([]*domain.ActivityLot, decimal.Decimal, error) {

	logger := logger.FromContext(ctx) // ← gets processor's logger
	tvalue := decimal.Zero
	touched := make([]*domain.ActivityLot, 0)

	acct := gl.acctsm[actv.AccountID]
	if len(acct.ID) == 0 {
		return touched, tvalue, fmt.Errorf("account does not exist for %s", actv.AccountID)
	}
	if gl.debug {
		logger.Info("ReduceLotQty", "Account", actv.AccountID)
		logger.Info("ReduceLotQty", "Symbol", fmt.Sprintf("%s-%v", actv.SentSymbol, samount))
		key := getAccountSymbolKey(acct.ID, actv.SentSymbol)
		alots := gl.lotsMap[key]
		logger.Info("ReduceLotQty", "key", key, "lots", len(alots))
		for _, lot := range alots {
			logger.Debug("ReduceLotQty", "lot", lot.Debug())
		}
	}

	lots := gl.MatchOpenLots(ctx, acct, actv.SentSymbol)
	if gl.debug {
		logger.Debug("ReduceLotQty", "Lots", len(lots))
	}

	// set total qty
	tqty := decimal.Zero
	aqty := samount
	if gl.debug {
		for _, lot := range lots {
			logger.Debug("ReduceLotQty-000", "lot", lot.Debug())
		}
	}

	for _, lot := range lots {
		if gl.debug {
			logger.Debug("ReduceLotQty-1", "lot", lot.Debug())
		}
		cqty := lot.Amount
		if tqty.Add(cqty).GreaterThan(aqty) {
			dtqty := tqty
			dtqty = dtqty.Add(cqty).Sub(aqty)
			cqty = cqty.Sub(dtqty)
		}

		// snapshot the consumed qty before reducing
		touchedLot := domain.ActivityLot{}
		touchedLot.UID = lot.UID
		touchedLot.AccountID = lot.AccountID
		touchedLot.ActivityID = lot.ActivityID
		touchedLot.Date = lot.Date
		touchedLot.Amount = cqty.Round(MAX_DECIMALS)
		touchedLot.Cost = lot.Cost
		touchedLot.CostValue = cqty.Mul(lot.Cost).Round(MAX_DECIMALS)
		tvalue = tvalue.Add(touchedLot.CostValue)

		logger.Trace("ConsumeQty", "cqty", cqty)
		// reduce lot qty
		lot.Amount = lot.Amount.Sub(cqty)
		lot.CostValue = lot.Amount.Mul(lot.Cost)

		if gl.debug {
			logger.Debug("ReduceLotQty", "Amount", lot.Amount)
		}
		// close the lot if zero
		if lot.Amount.IsZero() {
			lot.Status = domain.LotStatusClosed
		}

		// sum up the total quantity and value
		tqty = tqty.Add(cqty)
		if gl.debug {
			logger.Debug("Touched", "lot", touchedLot.Debug())
			logger.Debug("ReduceLotQty-2", "lot", lot.Debug())
			logger.Trace("ConsumeQty", "tqty", tqty)
		}
		touched = append(touched, &touchedLot)

		if tqty.GreaterThanOrEqual(aqty) {
			break
		}
	}

	if gl.debug {
		for _, lot := range lots {
			logger.Debug("ReduceLotQty-555", "lot", lot.Debug())
		}
	}

	return touched, tvalue, nil
}

func (gl *GainLoss) StoreTransfer(ctx context.Context, actv *domain.Activity, lots []*domain.ActivityLot) {

	logger := logger.FromContext(ctx) // ← gets processor's logger
	logger.Debug("StoreTransfer", "lots", lots)

	gl.transferActivities[actv.ID] = actv
	gl.transferLots[actv.ID] = lots
}

func (gl GainLoss) UpdateCashLot(ctx context.Context, actv *domain.Activity, acctId string, symbol string, amount decimal.Decimal) (*domain.ActivityLot, error) {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	var lot *domain.ActivityLot
	key := getAccountSymbolKey(acctId, symbol)
	lots := gl.lotsMap[key]
	logger.Debug("UpdateCashLot", "Key", key, "lots", len(lots))

	if len(lots) == 0 {
		lot = gl.CreateAssetLot(ctx, actv, acctId, symbol, decimal.Zero, decimal.Zero)
		lots = []*domain.ActivityLot{}
		lots = append(lots, lot)
		gl.lotsMap[key] = lots
	}

	lot = lots[0]
	logger.Debug("UpdateCashLot", "Deposit Qty", fmt.Sprintf("%s-%v", symbol, amount))
	logger.Debug("UpdateCashLot", "Prev Qty", fmt.Sprintf("%v", lot.CostValue))

	switch actv.TxnType {
	case domain.ActivityTypeBuy, domain.ActivityTypeWithdraw:
		lot.Amount = lot.Amount.Sub(amount)
		lot.CostValue = lot.CostValue.Sub(amount)
	default:
		lot.Amount = lot.Amount.Add(amount)
		lot.CostValue = lot.CostValue.Add(amount)
	}

	logger.Debug("UpdateCashLot", "Updated Qty", fmt.Sprintf("%v", lot.CostValue))
	if lot.Amount.IsZero() {
		lot.Cost = decimal.Zero
	} else {
		lot.Cost = lot.CostValue.Div(lot.Amount)
	}
	lot.Cost = lot.Cost.Round(MAX_DECIMALS)

	return lot, nil
}

func (gl GainLoss) UpdateBankLot(ctx context.Context, actv *domain.Activity) (*domain.ActivityLot, error) {

	logger := logger.FromContext(ctx) // ← gets processor's logger

	acctId := ""
	symbol := ""
	amount := decimal.Zero
	switch actv.TxnType {
	case domain.ActivityTypeBuy:
		acctId = actv.SentAccountID
		symbol = actv.SentSymbol
		amount = actv.SentAmount.Add(actv.Fee)
	case domain.ActivityTypeDeposit:
		acctId = actv.SentAccountID
		symbol = actv.RcvSymbol
		amount = actv.RcvAmount
	case domain.ActivityTypeWithdraw:
		acctId = actv.RcvAccountID
		symbol = actv.SentSymbol
		amount = actv.SentAmount
	case domain.ActivityTypeSell:
		acctId = actv.RcvAccountID
		symbol = actv.RcvSymbol
		amount = actv.RcvAmount
	}

	if len(acctId) == 0 {
		return nil, fmt.Errorf("Id: %s", actv.ID)
	}
	if len(symbol) == 0 {
		return nil, fmt.Errorf("symbol: %s-%s", symbol, actv.ID)
	}

	var lot *domain.ActivityLot
	key := getAccountSymbolKey(acctId, symbol)

	logger.Debug("UpdateBankLot", "Key", key)
	logger.Debug("UpdateBankLot", "Actv Qty", fmt.Sprintf("%s-%v", symbol, amount))

	lots := gl.lotsMap[key]
	if len(lots) == 0 {
		lot = gl.CreateAssetLot(ctx, actv, acctId, symbol, decimal.Zero, decimal.Zero)
		lots = []*domain.ActivityLot{}
		lots = append(lots, lot)
		gl.lotsMap[key] = lots
	}

	lot = lots[0]
	logger.Debug("UpdateBankLot", "Prev Qty", fmt.Sprintf("%v", lot.CostValue))

	switch actv.TxnType {
	case domain.ActivityTypeDeposit, domain.ActivityTypeBuy:
		lot.Amount = lot.Amount.Sub(amount)
		lot.CostValue = lot.CostValue.Sub(amount)
	case domain.ActivityTypeWithdraw, domain.ActivityTypeSell:
		lot.Amount = lot.Amount.Add(amount)
		lot.CostValue = lot.CostValue.Add(amount)
	}

	logger.Debug("UpdateBankLot", "Updated Qty", fmt.Sprintf("%v", lot.CostValue))
	lot.Cost = lot.CostValue.Div(lot.Amount)

	return lot, nil
}

func (gl GainLoss) UpdateFeeLot(ctx context.Context, actv *domain.Activity) decimal.Decimal {

	value := decimal.Zero
	// logger := logger.FromContext(ctx) // ← gets processor's logger
	if strings.Compare(actv.FeeCurrency, "USD") == 0 {

		switch actv.TxnType {
		case domain.ActivityTypeDeposit, domain.ActivityTypeBuy:
			gl.UpdateCashLot(ctx, actv, actv.AccountID, actv.FeeCurrency, actv.Fee)
		case domain.ActivityTypeWithdraw, domain.ActivityTypeSell:
			gl.UpdateCashLot(ctx, actv, actv.AccountID, actv.FeeCurrency, actv.Fee.Neg())
		}

		value = actv.Fee
	} else {
		_, value, _ = gl.ReduceLotQty(ctx, actv, actv.Fee)
	}
	return value
}

// resolveLotMatchingMethod returns the correct method for an account.
// Account level overrides user preference. Falls back to category default.
func (gl *GainLoss) resolveLotMatchingMethod(account domain.Account) domain.CostBasisMethod {
	// account level override — user explicitly set it
	if account.CostBasisMethod != "" {
		return account.CostBasisMethod
	}

	// user global preference
	if gl.costBasisMethod != "" {
		return gl.costBasisMethod
	}

	// category default
	return defaultLotMatchingMethod(account.Category)
}

func (gl *GainLoss) getOpenBalance(acctId string, symbol string) decimal.Decimal {

	balance := decimal.Zero
	key := getAccountSymbolKey(acctId, symbol)
	lots := gl.lotsMap[key]
	if gl.debug {
		gl.logger.Debug("getOpenBalance", "Key", key, "Lots", len(lots))
	}

	for _, lot := range lots {
		if gl.debug {
			gl.logger.Debug("getOpenBalance", "lot", lot.Debug())
		}
		if lot.Status == domain.LotStatusOpen {
			balance = balance.Add(lot.Amount)
		}
	}

	return balance
}

func (gl *GainLoss) sortLots(method domain.CostBasisMethod, lots []*domain.ActivityLot) {
	switch method {
	case domain.CostBasisHIFO:

		sort.SliceStable(lots, func(i, j int) bool {
			return lots[i].Cost.GreaterThan(lots[j].Cost)
		})
	default:
		sort.SliceStable(lots, func(i, j int) bool {
			return lots[i].Date.Before(*lots[j].Date)
		})
	}
}

// get account and symbol key
func getAccountSymbolKey(account, currency string) string {
	return fmt.Sprintf("%s:%s", account, currency)
}

// defaultLotMatchingMethod returns the default method for an account category.
func defaultLotMatchingMethod(category domain.AccountCategory) domain.CostBasisMethod {
	switch category {
	case domain.CategoryCrypto:
		return domain.CostBasisHIFO
	default:
		return domain.CostBasisFIFO
	}
}

func classifyTerm(acquired, disposed time.Time) domain.TaxTerm {
	oneYearLater := acquired.AddDate(1, 0, 0)
	if disposed.After(oneYearLater) {
		return domain.TermLong
	}
	return domain.TermShort
}
