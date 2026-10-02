// Package crypto encrypts credentials at rest with AES-256-GCM and derives
// the non-reversible fingerprint the API shows instead of a token.
package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
)

const nonceSize = 12

// Cipher seals and opens byte strings with one AES-256-GCM key.
type Cipher struct {
	aead cipher.AEAD
}

// New returns a Cipher for a 32-byte key.
func New(key []byte) (*Cipher, error) {
	if len(key) != 32 {
		return nil, fmt.Errorf("crypto: key must be 32 bytes, got %d", len(key))
	}
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}
	return &Cipher{aead: aead}, nil
}

// Encrypt returns nonce || ciphertext. An empty plaintext yields nil so a
// blank credential is stored as "absent".
func (c *Cipher) Encrypt(plain []byte) ([]byte, error) {
	if len(plain) == 0 {
		return nil, nil
	}
	nonce := make([]byte, nonceSize)
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return c.aead.Seal(nonce, nonce, plain, nil), nil
}

// Decrypt reverses Encrypt. Nil input yields an empty plaintext.
func (c *Cipher) Decrypt(sealed []byte) ([]byte, error) {
	if len(sealed) == 0 {
		return nil, nil
	}
	if len(sealed) < nonceSize {
		return nil, errors.New("crypto: ciphertext too short")
	}
	return c.aead.Open(nil, sealed[:nonceSize], sealed[nonceSize:], nil)
}

// Fingerprint returns "sha256:<8 hex> len=N" for a secret so the UI can
// tell tokens apart without seeing them.
func Fingerprint(secret string) string {
	if secret == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(secret))
	return fmt.Sprintf("sha256:%s len=%d", hex.EncodeToString(sum[:4]), len(secret))
}

// RandomHex returns n random bytes as a hex string.
func RandomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
