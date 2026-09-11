package processor

type AssetClass string

const (
	AssetClassCash       AssetClass = "cash"       // fiat currencies
	AssetClassStablecoin AssetClass = "stablecoin" // pegged, but still tracked with real lots
	AssetClassCrypto     AssetClass = "crypto"     // everything else
)

// assetClassOverrides lets you decide explicitly, per symbol, rather than
// guessing from a hardcoded "is this a stablecoin" heuristic.
var assetClassOverrides = map[string]AssetClass{
	"USD": AssetClassCash,
	"EUR": AssetClassCash,
	"GBP": AssetClassCash,

	"USDC": AssetClassStablecoin,
	"USDT": AssetClassStablecoin,
	"DAI":  AssetClassStablecoin,
	"GUSD": AssetClassStablecoin, // explicit decision: track like any other crypto lot, not cash
	// ... add more as you encounter them
}

func ClassifyAsset(symbol string) AssetClass {
	if class, ok := assetClassOverrides[symbol]; ok {
		return class
	}
	return AssetClassCrypto // default: unknown symbols get full lot treatment
}
