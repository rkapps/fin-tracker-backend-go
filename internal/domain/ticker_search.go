package domain

// TickerSearch defines search criteria
type TickerSearch struct {
	Function     string   `json:"function"`
	Strategies   []string `json:"strategies"`
	Sectors      []string `json:"sectors"`
	Industries   []string `json:"industries"`
	SearchText   string   `json:"searchText"`
	PerfPeriod   string   `json:"perfPeriod"`
	FromPerfPerc float64  `json:"fromPerfPerc"`
	ToPerfPerc   float64  `json:"toPerfPerc"`
	FromYield    float64  `json:"fromYield"`
	ToYield      float64  `json:"toYield"`
	RsiPeriod    string   `json:"rsiPeriod"`
	FromRsi      int      `json:"fromRsi"`
	ToRsi        int      `json:"toRsi"`
	PrAbove      bool     `json:"prAbove"`
	PrMA         string   `json:"prMa"`
	PrPeriod     string   `json:"prPeriod"`
}
