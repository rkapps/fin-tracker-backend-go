package domain

import (
	"encoding/json"
	"time"
)

// domain/raw_item.go
type RawItem struct {
	ID         string          `bson:"id"`
	UID        string          `bson:"uid"`
	AccountID  string          `bson:"accountId"`
	Provider   string          `bson:"provider"`
	Stream     string          `bson:"stream"`
	ExternalID string          `bson:"externalId"`
	Timestamp  time.Time       `bson:"timestamp"`
	Payload    json.RawMessage `bson:"payload"`
}

func (r *RawItem) Id() string             { return r.ID }
func (r *RawItem) CollectionName() string { return RAW_ITEM_COLLECTION_NAME }
