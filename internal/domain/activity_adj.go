package domain

type ActivityAdj struct {
	UID           string `json:"-"`
	ID            string `json:"-"`
	TxnType       string `json:"txnType"`
	Tag           string
	AdjustSeconds string `bson:"adjust_seconds"`
}

// Id returns the unique id for the ticker
func (a *ActivityAdj) Id() string {
	return a.ID
}

func (a *ActivityAdj) CollectionName() string {
	return ACTIVITY_ADJ_COLLECTION_NAME
}
