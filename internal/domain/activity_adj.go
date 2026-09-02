package domain

import "time"

type ActivityAdj struct {
	UID     string `json:"-"`
	ID      string `json:"-"`
	TxnType string `json:"txnType"`
	Tag     string
	Date    *time.Time
}

// Id returns the unique id for the ticker
func (a *ActivityAdj) Id() string {
	return a.ID
}

func (a *ActivityAdj) CollectionName() string {
	return ACTIVITY_ADJ_COLLECTION_NAME
}
