package domain

import (
	"time"
)

type Transaction struct {
	UID         string    `bson:"uid" json:"-"`
	ID          string    `json:"id" bson:"id"`
	Date        time.Time `json:"date"`
	Group       string    `json:"group"`
	Category    string    `json:"category"`
	Account     string    `json:"account"`
	Description string    `json:"description"`
	Dbcr        string    `json:"dbcr"`
	Amount      float64   `json:"amount"`
	DAmount     float64   `json:"damount"`
	CAmount     float64   `json:"camount"`
	Tag         string    `json:"tag"`
}

// Tickers is an array of tickers
type Transactions []*Transaction

// Id returns the unique id for the ticker
func (t *Transaction) Id() string {
	return t.ID
}

func (t *Transaction) CollectionName() string {
	return TRANSACTION_COLLECTION_NAME
}

// TransactionAgg holds metadata for aggregate data of transactions
type TransactionAgg struct {
	ID struct {
		Year     int32  `json:"year"`
		Month    int32  `json:"month"`
		Group    string `json:"group"`
		Category string `json:"category"`
		Account  string `json:"account"`
	} `bson:"_id"`
	Amount float64 `json:"amount"`
}
