package domain

import (
	"fmt"
	"time"

	"github.com/shopspring/decimal"
)

type CryptoPrice struct {
	ID     string          `json:"id"        bson:"_id"`
	Symbol string          `json:"symbol"`
	Ms     int64           `json:"ms"`
	Date   time.Time       `json:"date"`
	Price  decimal.Decimal `json:"price"`
}

// Id returns the unique id for the ticker
func (a *CryptoPrice) Id() string {
	return a.ID
}

func (a *CryptoPrice) SetId() {
	a.ID = fmt.Sprintf("%s-%d", a.Symbol, a.Ms)
}
func (a *CryptoPrice) CollectionName() string {
	return CRYPTO_PRICE_COLLECTION_NAME
}
