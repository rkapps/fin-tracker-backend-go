// Package crypto provides simple symmetric encryption for secrets at rest
// (API keys/secrets in AccountCredential). AES-256-GCM, one static key.
package core

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"
)

// EncryptionService is the interface AccountsService depends on. Swappable later
// for a KMS-backed implementation without touching any caller.
type EncryptionService interface {
	Encrypt(plaintext []byte) ([]byte, error)
	Decrypt(ciphertext []byte) ([]byte, error)
}

type AESGCMEncryptionService struct {
	gcm cipher.AEAD
}

// NewAESGCMEncryptionService builds the service from a 32-byte key (AES-256).
// Generate one with: openssl rand -base64 32
// Store it in your secrets manager / env — never in source.
func NewAESGCMEncryptionService(key []byte) (*AESGCMEncryptionService, error) {
	if len(key) != 32 {
		return nil, errors.New("crypto: key must be 32 bytes")
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &AESGCMEncryptionService{gcm: gcm}, nil
}

// Encrypt returns nonce||ciphertext, base — no separate storage of the nonce needed.
func (c *AESGCMEncryptionService) Encrypt(plaintext []byte) ([]byte, error) {
	nonce := make([]byte, c.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}
	return c.gcm.Seal(nonce, nonce, plaintext, nil), nil
}

// Decrypt expects the format produced by Encrypt: nonce||ciphertext.
func (c *AESGCMEncryptionService) Decrypt(ciphertext []byte) ([]byte, error) {
	ns := c.gcm.NonceSize()
	if len(ciphertext) < ns {
		return nil, errors.New("crypto: ciphertext too short")
	}
	nonce, ct := ciphertext[:ns], ciphertext[ns:]
	return c.gcm.Open(nil, nonce, ct, nil)
}
