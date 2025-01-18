package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/exp/rand"
)

const (
	pbkdf2IterationRounds = 10000
	pbkdf2SubkeyLength    = 256 / 8
	pbkdf2SaltSize        = 128 / 8
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

func HashPassword(password string) string {
	salt := []byte("somesalt") // TODO: Replace with random string
	var time uint32 = 1
	var memory uint32 = 64 * 1024
	var threads uint8 = 4
	var keyLen uint32 = 32

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$m=%d,t=%d,p=%d$%s$%s", memory, time, threads, saltBase64, hashBase64)
}

func HashLegacyPassword(password string) string {
	salt := []byte("somesaltsomesalt") // TODO: Replace with random string

	hash := pbkdf2.Key([]byte(password), salt, pbkdf2IterationRounds, pbkdf2SubkeyLength, sha256.New)

	encodedHash := make([]byte, 1+len(salt)+len(hash))
	encodedHash[0] = 0x1
	copy(encodedHash[1:], salt)
	copy(encodedHash[1+len(salt):], hash)

	return base64.StdEncoding.EncodeToString(encodedHash)
}

func VerifyLegacyPassword(password string, hash string) bool {
	decodedHash, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}

	if decodedHash[0] != 0x1 {
		return false
	}

	salt := decodedHash[1 : pbkdf2SaltSize+1]
	subkey := decodedHash[pbkdf2SaltSize+1:]

	derivedKey := pbkdf2.Key([]byte(password), salt, pbkdf2IterationRounds, pbkdf2SubkeyLength, sha256.New)

	return subtle.ConstantTimeCompare(derivedKey, subkey) == 1
}

func VerifyPassword(password string, hash string) bool {
	var time uint32
	var memory uint32
	var threads uint8

	var remaining string
	_, err := fmt.Sscanf(hash, "$argon2id$m=%d,t=%d,p=%d$%s", &memory, &time, &threads, &remaining)
	if err != nil {
		return false
	}

	parts := strings.SplitN(remaining, "$", 2)
	if len(parts) != 2 {
		return false
	}
	saltBase64 := parts[0]
	decodedHashBase64 := parts[1]

	salt, err := base64.RawStdEncoding.DecodeString(saltBase64)
	if err != nil {
		return false
	}
	decodedHash, err := base64.RawStdEncoding.DecodeString(decodedHashBase64)
	if err != nil {
		return false
	}

	computedHash := argon2.IDKey([]byte(password), salt, time, memory, threads, uint32(len(decodedHash)))

	return subtle.ConstantTimeCompare(computedHash, decodedHash) == 1
}

func GenerateRandomString(limit int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, limit)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
