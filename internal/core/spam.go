package core

import (
	"strings"

	"github.com/rkapps/fin-tracker-backend-go/internal/domain"
	"github.com/rkapps/fin-tracker-backend-go/internal/storage"
)

type CryptoSpamService struct {
	cstorage storage.CryptoStorageService
	spamm    map[string][]*domain.CryptoSpam
}

func NewCryptoSpamService(cstorage storage.CryptoStorageService) CryptoSpamService {
	spamm := make(map[string][]*domain.CryptoSpam)
	return CryptoSpamService{cstorage, spamm}
}

func (c CryptoSpamService) LoadCryptoSpams() {

	spams, _ := c.cstorage.GetCryptoSpams()
	for _, spam := range spams {
		c.spamm[spam.Blockchain] = append(c.spamm[spam.Blockchain], spam)
	}
}

func (c CryptoSpamService) IsSpamSolanaSignature(signature string, blockchain string) bool {
	spams := c.spamm[blockchain]
	for _, spam := range spams {
		if strings.Compare(strings.ToLower(spam.Source), strings.ToLower(signature)) == 0 {
			return true
		}
	}
	return false
}

func (c CryptoSpamService) IsSpamEthereumContractAddress(caddress string, blockchain string) bool {
	spams := c.spamm[blockchain]
	for _, spam := range spams {
		if strings.Compare(strings.ToLower(spam.Caddress), strings.ToLower(caddress)) == 0 {
			return true
		}
	}
	return false
}

func (c CryptoSpamService) IsSpamEthereumFromAddress(faddress string, blockchain string) bool {
	spams := c.spamm[blockchain]
	for _, spam := range spams {
		if strings.Compare(strings.ToLower(spam.Faddress), strings.ToLower(faddress)) == 0 {
			return true
		}
	}
	return false
}

func (c CryptoSpamService) IsSpamEthereumSymbol(symbol string, blockchain string) bool {
	spams := c.spamm[blockchain]
	for _, spam := range spams {
		if strings.Compare(strings.ToLower(spam.Symbol), strings.ToLower(symbol)) == 0 {
			return true
		}
	}
	return false
}
