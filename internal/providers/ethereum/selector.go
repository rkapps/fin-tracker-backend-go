package ethereum

import "github.com/rkapps/fin-tracker-backend-go/internal/domain"

type SelectorCategory string

const (
	CategoryTransfer   SelectorCategory = "transfer"   // plain ERC-20 transfer()
	CategoryApprove    SelectorCategory = "approve"    // no value movement — filter out
	CategoryReward     SelectorCategory = "reward"     // claim/harvest — income
	CategoryWithdrawal SelectorCategory = "withdrawal" // unstake/removeLiquidity — disposal or transfer, NOT income
	CategoryDeposit    SelectorCategory = "deposit"    // stake/addLiquidity — acquisition, cost basis tracking
	CategorySwap       SelectorCategory = "swap"       // DEX trade — disposal + acquisition pair
)

type Selector struct {
	Category SelectorCategory
	TxnType  domain.ActivityType
	Label    string // human-readable, for logs/notes

}

var knownSelectors = map[string]Selector{
	"0x095ea7b3": {CategoryApprove, domain.ActivityTypeFee, "approve"},
	"0x5c19a95c": {CategoryApprove, domain.ActivityTypeFee, "delegate"},

	"0x9a99b4f0": {CategoryReward, domain.ActivityTypeReward, "claimRewards"},
	"0x372500ab": {CategoryReward, domain.ActivityTypeReward, "claimRewards (Balancer)"},
	"0x2b630140": {CategoryReward, domain.ActivityTypeReward, "claimDistributions (Balancer)"},
	"0xc7b8981c": {CategoryReward, domain.ActivityTypeReward, "withdrawRewards (Matic)"},
	"0xc804c39a": {CategoryReward, domain.ActivityTypeReward, "claimWeeks"},

	//deposit
	"0x2e1a7d4d": {CategoryWithdrawal, domain.ActivityTypeExitLiquidity, "send"},
	"0x782ed90c": {CategoryWithdrawal, domain.ActivityTypeExitLiquidity, "exitLiquidity"},
	"0x630d8c63": {CategoryWithdrawal, domain.ActivityTypeExitLiquidity, "claimBalance (bancor)"},
	"0xac9650d8": {CategoryWithdrawal, domain.ActivityTypeExitLiquidity, "multicall"},
	"0x8d16a14a": {CategoryWithdrawal, domain.ActivityTypeExitLiquidity, "unstakeClaimTokens"},

	"0x3805550f": {CategoryDeposit, domain.ActivityTypeReceive, "exit"},
	"0x0f6795f2": {CategoryDeposit, domain.ActivityTypeReceive, "processExits"},
	"0x993e1c42": {CategoryDeposit, domain.ActivityTypeReceive, "withdrawERC20For"},

	//withdrawal
	"0x4b14557e": {CategoryWithdrawal, domain.ActivityTypeSend, "requestDepositFor"},
	"0xe3dec8fb": {CategoryWithdrawal, domain.ActivityTypeSend, "depositFor"},
	"0x85eb3a35": {CategoryWithdrawal, domain.ActivityTypeSend, "depositERC20For (ronin)"},
	"0x8b9e4f93": {CategoryWithdrawal, domain.ActivityTypeSend, "depositERC20ForUser (matic)"},
	"0xcb827474": {CategoryWithdrawal, domain.ActivityTypeSend, "initiateDeposit (synthetix)"},
	"0x3ce33bff": {CategoryWithdrawal, domain.ActivityTypeSend, "matic bridge"},
	"0xb6b55f25": {CategoryWithdrawal, domain.ActivityTypeSend, "deposit (synthetix)"},

	"0xec9fce80": {CategoryWithdrawal, domain.ActivityTypeAddLiquidity, "depositAndJoin"},
	"0xe4a76726": {CategoryWithdrawal, domain.ActivityTypeAddLiquidity, "addLiquidity"},
	"0xa694fc3a": {CategoryWithdrawal, domain.ActivityTypeAddLiquidity, "stake"},

	// ???
	"0xf91b8a71": {CategoryWithdrawal, domain.ActivityTypeSend, "swapErc20 (solana)"},

	//transfer
	"0xa9059cbb": {CategoryDeposit, domain.ActivityTypeReceive, "transfer"},

	//swap
	"0x5f575529": {CategorySwap, domain.ActivityTypeTrade, "swap (weth)"},
	"0xe2b39746": {CategorySwap, domain.ActivityTypeTrade, "multihopBatchSwapExactIn"},
	"0x18cbafe5": {CategorySwap, domain.ActivityTypeTrade, "swapExactTokensForETH"},
	"0x945bcec9": {CategorySwap, domain.ActivityTypeTrade, "batchswap"},
	"0x52bbbe29": {CategorySwap, domain.ActivityTypeTrade, "swap"},
}

func lookupSelector(method string) (Selector, bool) {
	if len(method) < 10 {
		return Selector{}, false
	}
	sel, ok := knownSelectors[method[:10]]
	return sel, ok
}
