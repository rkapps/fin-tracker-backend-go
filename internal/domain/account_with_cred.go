package domain

import "encoding/json"

type AccountWithCredential struct {
	Account Account
	Config  json.RawMessage // plaintext — already decrypted before this is built
}
