package domain

import "time"

type AccountSyncState struct {
	ID           string     `json:"id" bson:"id"`
	UID          string     `json:"-" bson:"uid"`
	LastSyncDate *time.Time `json:"lastSyncDate,omitempty" bson:"lastSyncDate,omitempty"`
	Resync       bool       `json:"resync" bson:"resync"`
	Refresh      bool       `json:"refresh" bson:"refresh"`
	SyncStatus   string     `json:"syncStatus" bson:"syncStatus"` // pending, syncing, success, failed
	ErrorMessage string     `json:"errorMessage,omitempty" bson:"errorMessage,omitempty"`
}

// Id returns the unique id for the ticker
func (a *AccountSyncState) Id() string {
	return a.ID
}

func (a *AccountSyncState) CollectionName() string {
	return ACCOUNT_SYNC_STATE_COLLECTION_NAME
}
