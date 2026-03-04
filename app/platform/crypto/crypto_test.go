package crypto

import (
	"bytes"
	"crypto/rand"
	"errors"
	"strings"
	"testing"
)

func TestNewService(t *testing.T) {
	config := Config{
		Key:               "test-key-123",
		KeyDerivationSalt: "test-salt",
	}

	service, err := NewService(config)
	if err != nil {
		t.Fatalf("NewService failed: %v", err)
	}

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	if len(service.key) != 32 {
		t.Errorf("Expected key length 32, got %d", len(service.key))
	}
}

func TestNewServiceWithEmptyKey(t *testing.T) {
	config := Config{} // Empty config should fail

	_, err := NewService(config)
	if err == nil {
		t.Fatal("Expected error for empty key")
	}

	var cryptoErr *CryptoError
	if !errors.As(err, &cryptoErr) {
		t.Errorf("Expected CryptoError, got %T", err)
	}
}

func TestNewServiceWithKey(t *testing.T) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		t.Fatalf("Failed to generate test key: %v", err)
	}

	service, err := NewServiceWithKey(key)
	if err != nil {
		t.Fatalf("NewServiceWithKey failed: %v", err)
	}

	if service == nil {
		t.Fatal("Service should not be nil")
	}

	// Verify the key was copied, not referenced
	originalKey := make([]byte, 32)
	copy(originalKey, key)
	key[0] = ^key[0] // Modify original key

	if bytes.Equal(service.key, key) {
		t.Error("Service key should be a copy, not a reference")
	}

	if !bytes.Equal(service.key, originalKey) {
		t.Error("Service key should match the original key")
	}
}

func TestNewServiceWithInvalidKey(t *testing.T) {
	invalidKey := make([]byte, 16) // Wrong size

	_, err := NewServiceWithKey(invalidKey)
	if err == nil {
		t.Fatal("Expected error for invalid key size")
	}

	var cryptoErr *CryptoError
	if !errors.As(err, &cryptoErr) {
		t.Errorf("Expected CryptoError, got %T", err)
	}
}

func TestEncryptDecrypt(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	testCases := []struct {
		name      string
		plaintext string
	}{
		{"Short text", "Hello, World!"},
		{"Long text", strings.Repeat("This is a longer text for testing. ", 100)},
		{"Unicode text", "Hello 世界! 🌍 Здравствуй мир!"},
		{"Empty string", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if tc.plaintext == "" && tc.name == "Empty string" {
				// Test that empty plaintext returns an error
				_, err := service.Encrypt([]byte(tc.plaintext))
				if err == nil {
					t.Error("Expected error for empty plaintext")
				}
				return
			}

			// Encrypt
			encryptedData, err := service.Encrypt([]byte(tc.plaintext))
			if err != nil {
				t.Fatalf("Encryption failed: %v", err)
			}

			if encryptedData == nil {
				t.Fatal("Encrypted data should not be nil")
			}

			if len(encryptedData.Ciphertext) == 0 {
				t.Error("Ciphertext should not be empty")
			}

			if len(encryptedData.Nonce) == 0 {
				t.Error("Nonce should not be empty")
			}

			if encryptedData.Version != 1 {
				t.Errorf("Expected version 1, got %d", encryptedData.Version)
			}

			// Decrypt
			decrypted, err := service.Decrypt(encryptedData)
			if err != nil {
				t.Fatalf("Decryption failed: %v", err)
			}

			if string(decrypted) != tc.plaintext {
				t.Errorf("Decrypted text doesn't match original. Expected: %q, Got: %q", tc.plaintext, string(decrypted))
			}
		})
	}
}

func TestEncryptDecryptString(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	plaintext := "Hello, World! This is a test string."

	// Encrypt
	encrypted, err := service.EncryptString(plaintext)
	if err != nil {
		t.Fatalf("String encryption failed: %v", err)
	}

	if encrypted == "" {
		t.Error("Encrypted string should not be empty")
	}

	// Decrypt
	decrypted, err := service.DecryptString(encrypted)
	if err != nil {
		t.Fatalf("String decryption failed: %v", err)
	}

	if decrypted != plaintext {
		t.Errorf("Decrypted string doesn't match original. Expected: %q, Got: %q", plaintext, decrypted)
	}
}

func TestDecryptWithNilData(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	_, err = service.Decrypt(nil)
	if err == nil {
		t.Error("Expected error for nil encrypted data")
	}

	var cryptoErr *CryptoError
	if !errors.As(err, &cryptoErr) {
		t.Errorf("Expected CryptoError, got %T", err)
	}
}

func TestDecryptWithInvalidVersion(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	invalidData := &EncryptedData{
		Ciphertext: []byte("dummy"),
		Nonce:      []byte("dummy"),
		Version:    999,
	}

	_, err = service.Decrypt(invalidData)
	if err == nil {
		t.Error("Expected error for invalid version")
	}

	var cryptoErr *CryptoError
	if !errors.As(err, &cryptoErr) {
		t.Errorf("Expected CryptoError, got %T", err)
	}
}

func TestEncryptionDeterminism(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	plaintext := "Test data for determinism check"

	// Encrypt the same data multiple times
	encrypted1, err := service.Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("First encryption failed: %v", err)
	}

	encrypted2, err := service.Encrypt([]byte(plaintext))
	if err != nil {
		t.Fatalf("Second encryption failed: %v", err)
	}

	// Results should be different due to random nonces
	if bytes.Equal(encrypted1.Ciphertext, encrypted2.Ciphertext) {
		t.Error("Encryption should not be deterministic (different nonces should produce different ciphertexts)")
	}

	if bytes.Equal(encrypted1.Nonce, encrypted2.Nonce) {
		t.Error("Nonces should be different for each encryption")
	}

	// But both should decrypt to the same plaintext
	decrypted1, err := service.Decrypt(encrypted1)
	if err != nil {
		t.Fatalf("First decryption failed: %v", err)
	}

	decrypted2, err := service.Decrypt(encrypted2)
	if err != nil {
		t.Fatalf("Second decryption failed: %v", err)
	}

	if string(decrypted1) != plaintext || string(decrypted2) != plaintext {
		t.Error("Both decryptions should produce the original plaintext")
	}
}

func TestKeyRotation(t *testing.T) {
	service, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create service: %v", err)
	}

	// Get original key fingerprint
	originalFingerprint := service.GetKeyFingerprint()

	// Rotate key
	err = service.RotateKey()
	if err != nil {
		t.Fatalf("Key rotation failed: %v", err)
	}

	// Get new key fingerprint
	newFingerprint := service.GetKeyFingerprint()

	if originalFingerprint == newFingerprint {
		t.Error("Key fingerprint should change after rotation")
	}
}

func TestKeyFingerprint(t *testing.T) {
	service1, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create first service: %v", err)
	}

	service2, err := NewService(Config{Key: "test-key"})
	if err != nil {
		t.Fatalf("Failed to create second service: %v", err)
	}

	service3, err := NewService(Config{Key: "different-key"})
	if err != nil {
		t.Fatalf("Failed to create third service: %v", err)
	}

	fingerprint1 := service1.GetKeyFingerprint()
	fingerprint2 := service2.GetKeyFingerprint()
	fingerprint3 := service3.GetKeyFingerprint()

	// Same key should produce same fingerprint
	if fingerprint1 != fingerprint2 {
		t.Error("Same keys should produce same fingerprints")
	}

	// Different keys should produce different fingerprints
	if fingerprint1 == fingerprint3 {
		t.Error("Different keys should produce different fingerprints")
	}

	// Fingerprint should be base64 encoded
	if fingerprint1 == "" {
		t.Error("Fingerprint should not be empty")
	}
}

func TestCryptoErrorUnwrap(t *testing.T) {
	originalErr := errors.New("original error")
	cryptoErr := &CryptoError{
		Operation: "test",
		Err:       originalErr,
	}

	if cryptoErr.Unwrap() != originalErr {
		t.Error("CryptoError should unwrap to original error")
	}

	if cryptoErr.Error() == "" {
		t.Error("CryptoError should have non-empty error message")
	}
}
