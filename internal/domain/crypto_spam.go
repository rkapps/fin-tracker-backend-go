package domain

type CryptoSpam struct {
	ID         string
	Blockchain string
	Caddress   string
	Faddress   string
	Source     string
	Symbol     string
	Hash       string
	Comment    string
}

// Id returns the unique id for the ticker
func (a *CryptoSpam) Id() string {
	return a.ID
}

func (a *CryptoSpam) CollectionName() string {
	return CRYPTO_SPAM_COLLECTION_NAME
}
