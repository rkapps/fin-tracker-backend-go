package domain

const (

	//Fields
	FIELD_ID                 = "id"
	FIELD_SYMBOL             = "symbol"
	FIELD_EXCHANGE           = "exchange"
	FIELD_NAME               = "name"
	FIELD_SECTOR             = "sector"
	FIELD_INDUSTRY           = "industry"
	FIELD_OVERVIEW           = "overview"
	FIELD_ACTIVE             = "active"
	FIELD_MARKETCAP          = "marketCap"
	FIELD_YIELD              = "yield"
	FIELD_PRDIFFPERC_SEARCH  = "prDiffPercSearch"
	FIELD_PERFORMANCE_SEARCH = "performanceSearch"
	FIELD_DIFF               = "diff"
	FIELD_PRICE              = "price"
	FIELD_STRATEGIES         = "strategies"
	FIELD_HISTORY_SYMBOL     = "metadata.symbol"

	//Collections
	TICKER_CONTROL_COLLECTION_NAME   = "ticker_control"
	TICKER_COLLECTION_NAME           = "ticker"
	TICKER_HISTORY_COLLECTION_NAME   = "ticker_history"
	TICKER_INDICATOR_COLLECTION_NAME = "ticker_indicator"
	TICKER_SENTIMENT_COLLECTION_NAME = "ticker_sentiment"
	TICKER_EMBEDDING_COLLECTION_NAME = "ticker_embedding"

	//ExNasdaq defines the string NASDAQ
	ExNasdaq string = "NASDAQ"
	//ExNyse defines the string NYSE
	ExNyse string = "NYSE"
	//ExNyseArca defines the string NYSEARCA
	ExNyseArca string = "NYSEARCA"
	//ExCurrency defines the string CURRENCY
	ExCurrency string = "CURRENCY"
	//ExOtc defines the string OTC
	ExOtc string = "OTC"

	//SMA defines the string SMA
	SMA string = "SMA"
	//EMA defines the string EMA
	EMA string = "EMA"
	//RSI defines the string RSI
	RSI string = "RSI"
)

var (
	//PerfPeriods defines the performance periods
	PerfPeriods []string = []string{"1W", "1M", "3M", "6M", "YTD", "1Y", "2Y", "3Y", "5Y"}
	//RSIPeriods defines the RSI periods
	RSIPeriods []int = []int{5, 9, 14, 20, 26}
	//SMAPeriods defines the SMA periods
	SMAPeriods []int = []int{5, 10, 20, 50, 100, 200}
	//EMAPeriods defines the EMA periods
	EMAPeriods []int = []int{5, 9, 12, 21, 26, 50, 200}
)

// Granularity constants
const (
	GranularityDaily   = "1d"
	GranularityHourly  = "1h"
	GranularityFiveMin = "5m"
	GranularityOneMin  = "1m"
)

// Indicator type constants
const (
	IndicatorTypeSMA  = "sma"
	IndicatorTypeEMA  = "ema"
	IndicatorTypeRSI  = "rsi"
	IndicatorTypeMACD = "macd"
	IndicatorTypeBB   = "bollinger_bands"
)
