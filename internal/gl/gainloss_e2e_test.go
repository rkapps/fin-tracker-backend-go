package gl

// Package gainloss_test — external test package (black-box testing).
// Place this file in the SAME DIRECTORY as your gainloss.go file, e.g.:
//   internal/services/gainloss/gainloss.go
//   internal/services/gainloss/gainloss_e2e_test.go
//
// Using "gainloss_test" (not "gainloss") means this test only exercises the
// package's exported API (NewGainLoss, Run) — the same way any external
// caller would use it. If you need access to unexported internals, switch
// the package name to "gainloss" instead and drop the import of the
// gainloss package itself.

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/rkapps/fin-tracker-backend-go/cmd/common/logger"
	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// Test helpers
// ---------------------------------------------------------------------

func testLogConfig(t *testing.T) *logger.Config {
	t.Helper()
	// Adjust this to however your codebase builds a *logger.Config for tests.
	// If logger.Config has a simple constructor, use it directly, e.g.:
	//   return logger.NewConfig(logger.LevelError, io.Discard)
	// logConfig := logger.New()
	return logger.New()
	// require.NoError(t, err)
	// return cfg
}

func mustDate(s string) time.Time {
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		panic(err)
	}
	return t
}

func d(f float64) decimal.Decimal {
	return decimal.NewFromFloat(f)
}

// buyActivity — a cash-funded buy.
//
//	acctID     — where the purchased asset lot is created (AccountID)
//	sentAcctID — where cash is deducted from (SentAccountID)
//	SentAmount is GROSS (subtotal) — Fee is added on top by the processor
//	to form the fee-inclusive cost basis.
func buyActivity(id, acctID, sentAcctID, sentSymbol string, sentAmount float64, rcvSymbol string, rcvAmount, fee float64, date string) *domain.Activity {
	return &domain.Activity{
		ID:            id,
		AccountID:     acctID,
		TxnType:       domain.ActivityTypeBuy,
		Date:          mustDate(date),
		SentAccountID: sentAcctID,
		SentSymbol:    sentSymbol,
		SentAmount:    d(sentAmount),
		RcvSymbol:     rcvSymbol,
		RcvAmount:     d(rcvAmount),
		Fee:           d(fee),
		FeeCurrency:   sentSymbol,
	}
}

// sellActivity — same-account sell: the disposed asset lot AND the resulting
// cash proceeds both live under acctID (DisposalActivityProcessor does not
// use SentAccountID/RcvAccountID at all).
// RcvAmount is GROSS proceeds — the processor subtracts Fee internally
// before crediting cash, but uses the GROSS amount for the GL price
// calculation (CreateGLDisposal has no fee logic of its own).
func sellActivity(id, acctID, sentSymbol string, sentAmount float64, rcvSymbol string, rcvAmount, fee float64, date string) *domain.Activity {
	return &domain.Activity{
		ID:          id,
		AccountID:   acctID,
		TxnType:     domain.ActivityTypeSell,
		Date:        mustDate(date),
		SentSymbol:  sentSymbol,
		SentAmount:  d(sentAmount),
		RcvSymbol:   rcvSymbol,
		RcvAmount:   d(rcvAmount),
		Fee:         d(fee),
		FeeCurrency: rcvSymbol,
	}
}

func newTestAccount(id string, category domain.AccountCategory, taxStatus domain.TaxStatus) *domain.Account {
	return &domain.Account{
		ID:        id,
		UID:       "user1",
		Category:  category,
		TaxStatus: taxStatus,
		Active:    true,
	}
}

// sumLotAmount sums the remaining Amount across every open lot matching
// accountID+symbol in the result set.
func sumLotAmount(lots []*domain.ActivityLot, accountID, symbol string) decimal.Decimal {
	total := decimal.Zero
	for _, lot := range lots {
		if lot.AccountID == accountID && lot.Symbol == symbol {
			total = total.Add(lot.Amount)
		}
	}
	return total
}

// sumLotCostValue sums CostValue across every lot matching accountID+symbol.
// For a cash "lot", CostValue tracks the running cash balance.
func sumLotCostValue(lots []*domain.ActivityLot, accountID, symbol string) decimal.Decimal {
	total := decimal.Zero
	for _, lot := range lots {
		if lot.AccountID == accountID && lot.Symbol == symbol {
			total = total.Add(lot.CostValue)
		}
	}
	return total
}

// sumDisposalGain sums GainLoss across every GLDisposal entry for a given
// source activity ID (a single sell can produce multiple entries — one per
// lot consumed, e.g. under HIFO).
func sumDisposalGain(t *testing.T, entries []*domain.GLEntry, activityID string) decimal.Decimal {
	t.Helper()
	total := decimal.Zero
	found := false
	for _, e := range entries {
		if e.ActivityID != activityID || e.GLType != domain.GLTypeDisposal {
			continue
		}
		detail, ok := e.Detail.(*domain.GLDisposalDetail)
		require.True(t, ok, "GLEntry %s: expected *domain.GLDisposalDetail, got %T", e.ID, e.Detail)
		total = total.Add(detail.GainLoss)
		found = true
	}
	require.True(t, found, "no GLDisposal entries found for activity %s", activityID)
	return total
}

func filterLots(lots []*domain.ActivityLot, accountID, symbol string) []*domain.ActivityLot {
	var out []*domain.ActivityLot
	for _, l := range lots {
		if l.AccountID == accountID && l.Symbol == symbol {
			out = append(out, l)
		}
	}
	return out
}

// ---------------------------------------------------------------------
// The end-to-end test
// ---------------------------------------------------------------------

func TestGainLoss_FullRun_BuySellHIFO(t *testing.T) {
	logConfig := testLogConfig(t)

	accts := []*domain.Account{
		newTestAccount("bank-1", domain.CategoryCash, domain.TaxStatusTaxable),
		newTestAccount("coinbase-1", domain.CategoryCrypto, domain.TaxStatusTaxable),
	}
	user := domain.User{ID: "user1", CurrencyCode: "USD"}

	// Six buys + one sell — the same dataset used for the manual HIFO
	// walkthrough. Buys are funded from bank-1; the sell's proceeds land
	// back in coinbase-1 (matching DisposalActivityProcessor's actual
	// same-account-only behavior).
	activities := []*domain.Activity{
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 2000.00, "BTC", 0.36674494, 29.80, "2024-01-01"),
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 492.66, "BTC", 0.09070838, 7.34, "2024-01-05"),
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 985.32, "BTC", 0.1309317, 14.68, "2024-01-10"),
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 492.66, "BTC", 0.07927771, 7.34, "2024-01-15"),
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 1970.64, "BTC", 0.20193984, 29.36, "2024-01-20"),
		buyActivity(uuid.NewString(), "coinbase-1", "bank-1", "USD", 700.82, "BTC", 0.05005665, 10.44, "2024-01-25"),
	}
	sellID := uuid.NewString()
	activities = append(activities, sellActivity(sellID, "coinbase-1", "BTC", 0.41000016, "USD", 7709.03, 114.86, "2024-08-10"))

	gl := NewGainLoss(user, accts, false, logConfig)
	result, err := gl.Run(context.Background(), activities)
	require.NoError(t, err)

	// --- Reconcile individual lots, not just the aggregate sum ---
	lots := filterLots(result.Lots, "bank-1", "USD")
	for _, lot := range lots {
		gl.logger.Info("TestGainLoss", fmt.Sprintf("%s-%v", lot.AccountID, lot.Status), fmt.Sprintf("%s-%v", lot.Symbol, lot.Amount))
	}

	btcBalance := gl.getOpenBalance("coinbase-1", "BTC")
	usdBalance := gl.getOpenBalance("coinbase-1", "USD")
	bankBalance := gl.getOpenBalance("bank-1", "USD")
	gl.logger.Info("Run", "Coinbase", fmt.Sprintf("%s %v", "BTC", btcBalance))
	gl.logger.Info("Run", "Coinbase", fmt.Sprintf("%s %v", "USD", usdBalance))
	gl.logger.Info("Run", "Bank", fmt.Sprintf("%s  %v", "USD", bankBalance))

	assert.True(t, btcBalance.Equal(d(0.50965906)),
		"coinbase-1 BTC balance: expected 0.50965906, got %s", btcBalance)

	assert.True(t, usdBalance.Equal(d(7594.17)),
		"coinbase-1 USD balance: expected 7594.17, got %s", usdBalance)

	assert.True(t, bankBalance.Equal(d(6741.06).Neg()), // adjust to your actual expected bank total
		"bank-1 USD balance: expected -6741.06, got %s", bankBalance)

	for _, entry := range result.GLEntries {
		gl.logger.Info("TestGainloss", "GlEntry", entry)
	}

	// --- BTC remaining in coinbase-1 ---
	totalBought := d(0.36674494).
		Add(d(0.09070838)).
		Add(d(0.1309317)).
		Add(d(0.07927771)).
		Add(d(0.20193984)).
		Add(d(0.05005665))
	expectedBTC := totalBought.Sub(d(0.41000016))

	btcRemaining := sumLotAmount(result.Lots, "coinbase-1", "BTC")
	assert.True(t, btcRemaining.Equal(expectedBTC),
		"BTC remaining: expected %s, got %s", expectedBTC, btcRemaining)

	// --- Sell's total realized gain (HIFO across however many lots it consumed) ---
	totalGain := sumDisposalGain(t, result.GLEntries, sellID)

	diff := totalGain.Sub(d(3712.17)).Abs()
	assert.True(t, diff.LessThan(d(0.01)),
		"gain: expected ~3712.17, got %s (diff %s)", totalGain, diff)

	// --- bank-1 USD: only ever decreases — funds every buy, never receives proceeds ---
	totalSpent := d(2029.80).Add(d(500.00)).Add(d(1000.00)).Add(d(500.00)).Add(d(2000.00)).Add(d(711.26))
	bankUSD := sumLotCostValue(result.Lots, "bank-1", "USD")
	assert.True(t, bankUSD.Equal(totalSpent.Neg()),
		"bank-1 USD: expected %s, got %s", totalSpent.Neg(), bankUSD)

	// --- coinbase-1 USD: only the NET sell proceeds land here ---
	netProceeds := d(7709.03).Sub(d(114.86))
	coinbaseUSD := sumLotCostValue(result.Lots, "coinbase-1", "USD")
	assert.True(t, coinbaseUSD.Equal(netProceeds),
		"coinbase-1 USD: expected %s, got %s", netProceeds, coinbaseUSD)

	// --- sanity: every activity should have been processed (no silent skips) ---
	assert.Len(t, result.Actvs, len(activities))
}
