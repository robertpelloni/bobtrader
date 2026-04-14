package config

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"io"
)

// SecretsManager handles the encryption and decryption of sensitive configuration values.
// In a real production environment, this could interface with HashiCorp Vault or AWS Secrets Manager.
type SecretsManager struct {
	masterKey []byte
}

// NewSecretsManager initializes a secrets manager with a 32-byte master key.
func NewSecretsManager(masterKey string) (*SecretsManager, error) {
	keyBytes := []byte(masterKey)
	if len(keyBytes) != 32 {
		return nil, fmt.Errorf("master key must be exactly 32 bytes for AES-256")
	}

	return &SecretsManager{
		masterKey: keyBytes,
	}, nil
}

// Encrypt encrypts a plaintext string using AES-GCM and returns a base64 encoded string.
func (s *SecretsManager) Encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := aesGCM.Seal(nonce, nonce, []byte(plaintext), nil)
	return base64.StdEncoding.EncodeToString(ciphertext), nil
}

// Decrypt decodes and decrypts a base64 encoded ciphertext string.
func (s *SecretsManager) Decrypt(encodedCiphertext string) (string, error) {
	ciphertext, err := base64.StdEncoding.DecodeString(encodedCiphertext)
	if err != nil {
		return "", fmt.Errorf("base64 decode error: %w", err)
	}

	block, err := aes.NewCipher(s.masterKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", fmt.Errorf("ciphertext too short")
	}

	nonce, cipherbytes := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := aesGCM.Open(nil, nonce, cipherbytes, nil)
	if err != nil {
		return "", fmt.Errorf("decrypt error: %w", err)
	}

	return string(plaintext), nil
}
