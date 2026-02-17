package domain

// TickerGroup holds the sector and industry combination
type TickerGroup struct {
	Sector   string `json:"sector" bson:"sector"`
	Industry string `json:"industry" bson:"industry"`
}

// TickerGroups holds a list of ticker groups.
type TickerGroups []*TickerGroup

type TickerGroupsAggregateResult struct {
	ID TickerGroup `bson:"_id"`
}
