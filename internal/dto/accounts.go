package dto

import (
	"encoding/json"
)

type AccountRequest struct {
	ID                string          `json:"id" bson:"id"`
	Name              string          `json:"name"`
	Active            bool            `json:"active" bson:"active"`
	Category          string          `json:"category"`
	Type              string          `json:"type"`
	TaxStatus         string          `json:"taxStatus"`
	AlternateNames    []string        `json:"alternateNames,omitempty" bson:"alternateNames,omitempty"`
	CostBasisMethod   string          `json:"costBasisMethod"`
	Detail            json.RawMessage `json:"detail"`                // matches CryptoDetail/BrokerageDetail/etc shape
	Credentials       json.RawMessage `json:"credentials,omitempty"` // only when Detail implies exchange/wallet
	Resync            bool            `json:"resync"`
	DeleteCredentials bool            `json:"deleteCredentials"`
}

// type AccountResponse struct {
// 	ID              string                 `json:"id" bson:"id"`
// 	Name            string                 `json:"name" bson:"name"`
// 	Active          bool                   `json:"active" bson:"active"`
// 	Category        domain.AccountCategory `json:"category" bson:"category"`
// 	Type            domain.AccountType     `json:"type" bson:"type"`
// 	AlternateNames  []string               `json:"alternateNames,omitempty" bson:"alternateNames,omitempty"`
// 	Detail          domain.AccountDetail   `json:"detail" bson:"detail"`
// 	TaxStatus       domain.TaxStatus       `json:"taxStatus" bson:"taxStatus"`
// 	CostBasisMethod domain.CostBasisMethod `json:"costBasisMethod" bson:"costBasisMethod"`
// 	CreatedAt       time.Time              `json:"createdAt" bson:"createdAt"`
// 	UpdatedAt       time.Time              `json:"updatedAt" bson:"updatedAt"`
// 	// Resync          bool                   `json:"resync"`
// 	// LastSyncedAt    time.Time              `json:"lastSyncedAt" bson:"lastSyncedAt"`
// 	// SyncStatus      string                 `json:"syncStatus"`
// }
