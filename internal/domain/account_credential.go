package domain

type AccountCredential struct {
	ID       string
	UID      string
	Provider string // "coinbase", "kraken", "alpaca"

	// Encrypted at the storage layer. Plaintext is provider.Config JSON,
	// e.g. coinbase.Config{KeyName, PrivateKey} or kraken.Config{APIKey, APISecret}.
	EncryptedConfig []byte

	// Encrypted fields
	// APIKey     string // Encrypted
	// APISecret  string // Encrypted
	// Passphrase string // Encrypted (some exchanges need this)

	// Metadata
	// Label       string   // User's nickname
	// Permissions []string // "read", "trade" - what access was granted
	// ExpiresAt   *time.Time
	// CreatedAt   time.Time
	// LastUsed    *time.Time
	// Active      bool
}

// Id returns the unique id for the ticker
func (a *AccountCredential) Id() string {
	return a.ID
}

func (a *AccountCredential) CollectionName() string {
	return ACCOUNT_CREDENTIAL_COLLECTION_NAME
}
