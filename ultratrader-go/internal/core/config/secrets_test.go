package config_test

import (
	"strings"
	"testing"

	"github.com/robertpelloni/bobtrader/ultratrader-go/internal/core/config"
)

func TestSecretsManager(t *testing.T) {
	// 32-byte key for AES-256
	key := "01234567890123456789012345678901"

	sm, err := config.NewSecretsManager(key)
	if err != nil {
		t.Fatalf("Failed to init SecretsManager: %v", err)
	}

	plaintext := "my_super_secret_binance_api_key_123!"

	// 1. Encrypt
	encrypted, err := sm.Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Encryption failed: %v", err)
	}

	if encrypted == plaintext {
		t.Errorf("Encrypted text is same as plaintext")
	}

	// 2. Decrypt
	decrypted, err := sm.Decrypt(encrypted)
	if err != nil {
		t.Fatalf("Decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted text %q does not match original plaintext %q", decrypted, plaintext)
	}

	// 3. Test uniqueness (IV/Nonce changes means same plaintext encrypts differently)
	encrypted2, _ := sm.Encrypt(plaintext)
	if encrypted == encrypted2 {
		t.Errorf("Encryption is not using a random nonce, outputs matched")
	}
}

func TestSecretsManager_InvalidKey(t *testing.T) {
	// 31 bytes instead of 32
	key := "0123456789012345678901234567890"

	_, err := config.NewSecretsManager(key)
	if err == nil {
		t.Errorf("Expected error for 31-byte key, got nil")
	}

	if !strings.Contains(err.Error(), "must be exactly 32 bytes") {
		t.Errorf("Expected 32-byte error message, got %v", err)
	}
}
