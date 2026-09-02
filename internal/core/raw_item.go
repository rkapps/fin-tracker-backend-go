package core

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"time"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
)

func GetGlobalRawItem(name, stream string, id string, raw json.RawMessage, date time.Time) domain.RawItem {

	return domain.RawItem{
		UID:        name,
		AccountID:  name,
		ID:         rawItemID(name, name, name, stream, id),
		ExternalID: id,
		Provider:   name,
		Stream:     stream,
		Payload:    raw,
		Timestamp:  date,
	}

}
func GetRawItem(account domain.Account, name, stream string, id string, raw json.RawMessage, date time.Time) domain.RawItem {

	return domain.RawItem{
		UID:        account.UID,
		AccountID:  account.ID,
		ID:         rawItemID(account.UID, account.ID, name, stream, id),
		ExternalID: id,
		Provider:   name,
		Stream:     stream,
		Payload:    raw,
		Timestamp:  date,
	}
}

func rawItemID(uid, accountID, provider, stream, externalID string) string {
	h := sha256.Sum256([]byte(uid + ":" + accountID + ":" + provider + ":" + stream + ":" + externalID))
	return hex.EncodeToString(h[:])[:32]
}
