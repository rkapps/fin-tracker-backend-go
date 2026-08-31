package kraken

import (
	"context"
	"fmt"
)

// Config are the per-account credentials (stored in core.Account.Config).
type Config struct {
	Api_Key    string `json:"api_key"`
	Api_Secret string `json:"api_secret"`
}

// coinbase.Provider only knows this:
type API interface {
	GetAssets(ctx context.Context) (map[string]AssetInfo, error)
	GetAssetPairs(ctx context.Context) (map[string]AssetPairInfo, error)
	GetTradeHistory(ctx context.Context, cfg Config, start string, offset string) (*TradesHistoryResponse, error)
	GetLedgers(ctx context.Context, cfg Config, start string, offset string) (*LedgersResponse, error)
}

type krLedgerCursor struct {
	Offset   int     `json:"offset,omitempty"`    // used only while BackfillDone == false
	MaxFTime float64 `json:"max_ftime,omitempty"` // running max timestamp seen — becomes the next "start"
	Start    string  `json:"start,omitempty"`     // persisted once backfill completes; used every run after
}

type krTradesCursor struct {
	Offset   int     `json:"offset,omitempty"`
	MaxFTime float64 `json:"max_ftime,omitempty"`
	Start    string  `json:"start,omitempty"`
}

// KrakenResponse wraps the Kraken API JSON response
type KrakenResponse struct {
	Error  []string    `json:"error"`
	Result interface{} `json:"result"`
}

// LedgersResponse represents an associative array of ledgers infos
type LedgersResponse struct {
	Ledger map[string]LedgerInfo `json:"ledger"`
}

// LedgerInfo Represents the ledger informations
type LedgerInfo struct {
	UID     string
	Acct_Id string
	InfoId  string
	RefID   string  `json:"refid"`
	Time    float64 `json:"time"`
	Type    string  `json:"type"`
	Aclass  string  `json:"aclass"`
	Asset   string  `json:"asset"`
	Amount  string  `json:"amount"`
	Fee     string  `json:"fee"`
	Balance string  `json:"balance"`
	// Cost    float64 `json:"cost"`
}

func (l LedgerInfo) Debug() string {
	return fmt.Sprintf("Ledger: %v %s-%s", l.Time, l.Type, l.Acct_Id)
}

// TradesHistoryResponse represents a list of executed trade
type TradesHistoryResponse struct {
	Trades map[string]TradeHistoryInfo `json:"trades"`
	Count  int                         `json:"count"`
}

// TradeHistoryInfo represents a transaction
type TradeHistoryInfo struct {
	UID       string
	Acct_Id   string
	TradeId   string  `json:"tradeid"`
	OrderTxID string  `json:"ordertxid"`
	PostxID   string  `json:"postxid"`
	AssetPair string  `json:"pair"`
	Time      float64 `json:"time"`
	Type      string  `json:"type"`
	OrderType string  `json:"ordertype"`
	Price     float64 `json:"price,string"`
	Cost      float64 `json:"cost,string"`
	Fee       float64 `json:"fee,string"`
	Volume    float64 `json:"vol,string"`
	Margin    float64 `json:"margin,string"`
	Misc      string  `json:"misc"`
}

// Assets — GET /0/public/Assets (public, no auth)
type AssetsResponse struct {
	Assets map[string]AssetInfo `json:"-"` // keyed by Kraken's internal code, e.g. "XETH", "TIA.B"
}
type AssetInfo struct {
	Aclass          string `json:"aclass"`  // "currency"
	Altname         string `json:"altname"` // display symbol, e.g. "ETH"
	Decimals        int    `json:"decimals"`
	DisplayDecimals int    `json:"display_decimals"`
}

// AssetPairs — GET /0/public/AssetPairs (public, no auth)
type AssetPairsResponse struct {
	AssetPairs map[string]AssetPairInfo `json:"-"` // keyed by Kraken's pair code, e.g. "XETHZUSD"
}
type AssetPairInfo struct {
	Altname     string `json:"altname"` // e.g. "ETHUSD"
	Wsname      string `json:"wsname"`  // e.g. "ETH/USD"
	AclassBase  string `json:"aclass_base"`
	Base        string `json:"base"` // Kraken internal code, e.g. "XETH"
	AclassQuote string `json:"aclass_quote"`
	Quote       string `json:"quote"` // Kraken internal code, e.g. "ZUSD"
	// fees, leverage_buy, leverage_sell, ordermin, etc. — omit unless you need them
}
