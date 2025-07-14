package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/pbkdf2"
)

// Service represents a cryptographic service with encryption and decryption capabilities
type Service struct {
	key []byte
}

// Config holds configuration for the crypto service
type Config struct {
	// Key is the base key for encryption. If empty, a random key will be generated.
	Key string `json:"key"`
	// KeyDerivationSalt is used for key derivation from passwords
	KeyDerivationSalt string `json:"key_derivation_salt"`
}

// CryptoError represents a cryptographic operation error
type CryptoError struct {
	Operation string
	Err       error
}

func (e *CryptoError) Error() string {
	return fmt.Sprintf("crypto operation '%s' failed: %v", e.Operation, e.Err)
}

func (e *CryptoError) Unwrap() error {
	return e.Err
}

// NewService creates a new crypto service with the provided configuration
func NewService(config Config) (*Service, error) {
	var key []byte

	if config.Key == "" {
		return nil, &CryptoError{
			Operation: "key validation",
			Err:       errors.New("key cannot be empty"),
		}
	}

	salt := []byte(config.KeyDerivationSalt)
	if len(salt) == 0 {
		salt = []byte("fundermaps-default-salt") // Default salt if none provided
	}
	key = pbkdf2.Key([]byte(config.Key), salt, 100000, 32, sha256.New)

	return &Service{key: key}, nil
}

// NewServiceWithKey creates a new crypto service with a direct key (for testing or specific use cases)
func NewServiceWithKey(key []byte) (*Service, error) {
	if len(key) != 32 {
		return nil, &CryptoError{
			Operation: "key validation",
			Err:       errors.New("key must be exactly 32 bytes for AES-256"),
		}
	}

	// Make a copy of the key to avoid external modifications
	keyCopy := make([]byte, 32)
	copy(keyCopy, key)

	return &Service{key: keyCopy}, nil
}

// EncryptedData represents encrypted data with metadata
type EncryptedData struct {
	Ciphertext []byte `json:"ciphertext"`
	Nonce      []byte `json:"nonce"`
	Version    int    `json:"version"`
}

// Encrypt encrypts plaintext using AES-GCM and returns the encrypted data
func (s *Service) Encrypt(plaintext []byte) (*EncryptedData, error) {
	if len(plaintext) == 0 {
		return nil, &CryptoError{
			Operation: "encryption",
			Err:       errors.New("plaintext cannot be empty"),
		}
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, &CryptoError{
			Operation: "encryption",
			Err:       fmt.Errorf("failed to create cipher: %w", err),
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &CryptoError{
			Operation: "encryption",
			Err:       fmt.Errorf("failed to create GCM: %w", err),
		}
	}

	// Generate a random nonce
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, &CryptoError{
			Operation: "encryption",
			Err:       fmt.Errorf("failed to generate nonce: %w", err),
		}
	}

	// Encrypt and authenticate
	ciphertext := gcm.Seal(nil, nonce, plaintext, nil)

	return &EncryptedData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    1,
	}, nil
}

// Decrypt decrypts ciphertext using AES-GCM and returns the plaintext
func (s *Service) Decrypt(data *EncryptedData) ([]byte, error) {
	if data == nil {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       errors.New("encrypted data cannot be nil"),
		}
	}

	if data.Version != 1 {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("unsupported version: %d", data.Version),
		}
	}

	if len(data.Ciphertext) == 0 {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       errors.New("ciphertext cannot be empty"),
		}
	}

	// Create AES cipher
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to create cipher: %w", err),
		}
	}

	// Create GCM mode
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to create GCM: %w", err),
		}
	}

	// Verify nonce size
	if len(data.Nonce) != gcm.NonceSize() {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("invalid nonce size: expected %d, got %d", gcm.NonceSize(), len(data.Nonce)),
		}
	}

	// Decrypt and verify authentication
	plaintext, err := gcm.Open(nil, data.Nonce, data.Ciphertext, nil)
	if err != nil {
		return nil, &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to decrypt or authenticate: %w", err),
		}
	}

	return plaintext, nil
}

// EncryptString encrypts a string and returns a base64-encoded result
func (s *Service) EncryptString(plaintext string) (string, error) {
	data, err := s.Encrypt([]byte(plaintext))
	if err != nil {
		return "", err
	}

	// Combine nonce and ciphertext for easy storage
	combined := make([]byte, len(data.Nonce)+len(data.Ciphertext))
	copy(combined, data.Nonce)
	copy(combined[len(data.Nonce):], data.Ciphertext)

	return base64.StdEncoding.EncodeToString(combined), nil
}

// DecryptString decrypts a base64-encoded string
func (s *Service) DecryptString(encryptedData string) (string, error) {
	combined, err := base64.StdEncoding.DecodeString(encryptedData)
	if err != nil {
		return "", &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to decode base64: %w", err),
		}
	}

	// Create AES cipher to get nonce size
	block, err := aes.NewCipher(s.key)
	if err != nil {
		return "", &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to create cipher: %w", err),
		}
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", &CryptoError{
			Operation: "decryption",
			Err:       fmt.Errorf("failed to create GCM: %w", err),
		}
	}

	nonceSize := gcm.NonceSize()
	if len(combined) < nonceSize {
		return "", &CryptoError{
			Operation: "decryption",
			Err:       errors.New("encrypted data too short"),
		}
	}

	// Split nonce and ciphertext
	nonce := combined[:nonceSize]
	ciphertext := combined[nonceSize:]

	data := &EncryptedData{
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    1,
	}

	plaintext, err := s.Decrypt(data)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}

// RotateKey generates a new key for the service
func (s *Service) RotateKey() error {
	newKey := make([]byte, 32)
	if _, err := rand.Read(newKey); err != nil {
		return &CryptoError{
			Operation: "key rotation",
			Err:       err,
		}
	}

	s.key = newKey
	return nil
}

// GetKeyFingerprint returns a SHA-256 hash of the current key for identification
func (s *Service) GetKeyFingerprint() string {
	hash := sha256.Sum256(s.key)
	return base64.StdEncoding.EncodeToString(hash[:])
}
