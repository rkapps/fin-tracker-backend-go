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
	CategoryWrap       SelectorCategory = "wrap"       // weth swap
	CategoryMultiToken SelectorCategory = "multitoken" // multiple tokens
	CategoryUnknown    SelectorCategory = "unknown"
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

	"0x3805550f": {CategoryDeposit, domain.ActivityTypeReceive, "exit"},
	"0x0f6795f2": {CategoryDeposit, domain.ActivityTypeReceive, "processExits"},
	"0x993e1c42": {CategoryDeposit, domain.ActivityTypeReceive, "withdrawERC20For"},
	"0xd7fd19dd": {CategoryDeposit, domain.ActivityTypeReceive, "relayMessage"},

	//withdrawal
	"0x4b14557e": {CategoryWithdrawal, domain.ActivityTypeSend, "requestDepositFor"},
	"0xe3dec8fb": {CategoryWithdrawal, domain.ActivityTypeSend, "depositFor"},
	"0x4faa8a26": {CategoryWithdrawal, domain.ActivityTypeSend, "depositEtherFor"},

	"0xb1a1a882": {CategoryWithdrawal, domain.ActivityTypeSend, "depositETH (optimism)"},
	"0xeee3f07a": {CategoryWithdrawal, domain.ActivityTypeSend, "depositEthFor (ronin)"},
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
	"0xa9059cbb": {CategoryTransfer, "", "transfer"},

	//swap
	"0xa0712d68": {CategorySwap, domain.ActivityTypeTrade, "mint"},
	// "0xa0712d68": {CategorySwap, domain.ActivityTypeTrade, "mint"},

	"0x1cff79cd": {CategorySwap, domain.ActivityTypeTrade, "execute"},
	"0xadc9772e": {CategorySwap, domain.ActivityTypeTrade, "stake"},
	"0x46ab38f1": {CategorySwap, domain.ActivityTypeTrade, "exitswapPoolAmountIn"},
	"0x8bdb3913": {CategorySwap, domain.ActivityTypeTrade, "exitPool"},

	"0xe2b39746": {CategorySwap, domain.ActivityTypeTrade, "multihopBatchSwapExactIn"},
	"0x18cbafe5": {CategorySwap, domain.ActivityTypeTrade, "swapExactTokensForETH"},
	"0x38ed1739": {CategorySwap, domain.ActivityTypeTrade, "swapExactTokensForTokens"},

	"0x945bcec9": {CategorySwap, domain.ActivityTypeTrade, "batchswap"},
	"0x5f575529": {CategorySwap, domain.ActivityTypeTrade, "swap (weth)"},
	// wrap
	"0x52bbbe29": {CategorySwap, domain.ActivityTypeTrade, "wrap"},

	"0xd0e30db0": {CategoryWrap, domain.ActivityTypeTrade, "deposit (weth)"},

	//unknown
	"0x029d3040": {CategoryUnknown, "", "sellVoucher"},
	"0x6ab15071": {CategoryUnknown, "", "stake"},
	"0x6140004b": {CategoryUnknown, "", "erc721"},
	"0x787a08a6": {CategoryUnknown, "", "cooldown (aave)"},
	"0x7c5264b4": {CategoryUnknown, "", "startExitWithBurntTokens"},
	"0xe02ae075": {CategoryUnknown, "", "leave (bancor)"},
	"0x7c544cc4": {CategoryUnknown, "", "migrateStkABPTWithPermit"},
	"0x4f91440d": {CategoryUnknown, "", "restake"},
	"0x357a0333": {CategoryUnknown, "", "initWithdrawal"},
	"0xaf086c7e": {CategoryUnknown, "", "issueMaxSynths"},
	"0x295da87d": {CategoryUnknown, "", "burnSynths"},
	"0x8e1a55fc": {CategoryUnknown, "", "build"},
	"0xdb006a75": {CategoryUnknown, "", "redeem"},
	"0x1aa3a008": {CategoryUnknown, "", "register"},
	"0x34fcd5be": {CategoryUnknown, "", "unknown-polygon"},
	"0xa1798512": {CategoryUnknown, "", "unknown-polygon-spam"},
	"0x18b35fe1": {CategoryUnknown, "", "unknown-polygon-spam"},
	"0xf97b19f5": {CategoryUnknown, "", "unknown-polygon-spam"},
	"0x8d16a14a": {CategoryUnknown, "", "unstakeClaimTokens"},

	"0x8aaa8f3b": {CategoryUnknown, "", "spam"},

	// multi tokens
	"0xf305d719": {CategoryMultiToken, "", "addLiquidityETH"},
	"0xb02f0b73": {CategoryMultiToken, "", "exitPool"},
	"0xb95cac28": {CategoryMultiToken, "", "joinPool"},

	"0x88316456": {CategoryMultiToken, "", "mint"},
}

func lookupSelector(method string) (Selector, bool) {
	if len(method) < 10 {
		return Selector{}, false
	}
	sel, ok := knownSelectors[method[:10]]
	return sel, ok
}

// not handled
// 0x1cff79cd
// execute
// hash - 0x8f88b2ad84876c16699d36aad88f90a125fadb196894da36256d96bae13208d8

// 0x6ab15071
// buyVoucher
// hash - 0x1b81345aef1f7deb7e1bd17444307d9b4a92647dd30141e4657f0281cae67a6d

// 0xe02ae075
// leave
// hash 0xb586a1e17fd6aed252bf0ea1499ae47085e58dbd807817d92b132a49d848e459

// 0xb95cac28
// joinPool
// hash 0x4f9b6faa56b57f4b28aef4b919407ba3a6d394b7fcffa90cb21ee8a9d776477b

// 0x4f91440d
// restake
// hash 0x88aaad4bfca43d2655611960f5179f9878389a3844503f3e1c740496e006e1ab

// 0x357a0333
// initWithdrawal
// hash 0x881961af03d91b5a7c2f5297b82f90177aac4d2d691327d0e77bba30be0ba9b8

// 0x295da87d
// burnSynths
// hash - 0xf1a00f473909a85b7f8ef563507b297269b151c46a6d65383c6e3dd52013eaa2

// 0x8e1a55fc
// build
// hash - 0xc650725cb64014a5fb1a14f04e866846066b17fddb6a7d83e7e2bb5d8ebebd69

// 0xdb006a75
// redeem
// hash - 0xc3da2593f8ab4acd520e9c6c07c0255e4487b902c174d7f6c53b0ef4d218cfda

// 0x88316456
// mint (erc721)
// hash - 0x9538a772d6786bc6487baf5e818e97c498bdbbc63b404f21dd7ae2592e46958e

// 0xb02f0b73
// exitPool
// hash - 0x17f8e564145e6efdad1d1ce352b410c633baf79bcdc92b70f31f928dec453e1a
