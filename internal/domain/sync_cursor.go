package domain

import "time"

type SyncCursor struct {
	ID           string    `bson:"id"`
	UID          string    `bson:"uid"`
	AccountID    string    `bson:"accountId"`
	Provider     string    `bson:"provider"`
	Stream       string    `bson:"stream"`
	Cursor       string    `bson:"cursor"` // opaque; provider decodes it
	BackfillDone bool      `bson:"backfillDone"`
	LastSyncedAt time.Time `bson:"lastSyncedAt"`
	LastError    string    `bson:"lastError,omitempty"`
}

func (s *SyncCursor) Id() string             { return s.ID }
func (s *SyncCursor) CollectionName() string { return SYNC_CURSOR_COLLECTION_NAME }
