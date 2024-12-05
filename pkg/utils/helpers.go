package utils

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"fmt"
	"time"

	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/pbkdf2"
	"golang.org/x/exp/rand"
)

const (
	iterationRounds = 10000
	subkeyLength    = 256 / 8
	saltSize        = 128 / 8
)

func init() {
	rand.Seed(uint64(time.Now().UnixNano()))
}

func HashPassword(password string) string {
	salt := []byte("somesalt")
	var time uint32 = 1
	var memory uint32 = 64 * 1024
	var threads uint8 = 4
	var keyLen uint32 = 32

	hash := argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen)

	saltBase64 := base64.RawStdEncoding.EncodeToString(salt)
	hashBase64 := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$m=%d,t=%d,p=%d$%s$%s", memory, time, threads, saltBase64, hashBase64)
}

func VerifyLegacyPassword(password, hash string) bool {
	decodedHash, err := base64.StdEncoding.DecodeString(hash)
	if err != nil {
		return false
	}

	if decodedHash[0] != 0x1 {
		return false
	}

	salt := decodedHash[1 : saltSize+1]
	subkey := decodedHash[saltSize+1:]

	derivedKey := pbkdf2.Key([]byte(password), salt, iterationRounds, subkeyLength, sha256.New)

	return subtle.ConstantTimeCompare(derivedKey, subkey) == 1
}

// func VerifyPassword(password, hash string) bool {
// 	var time uint32 = 1
// 	var memory uint32 = 64 * 1024
// 	var threads uint8 = 4
// 	var keyLen uint32 = 32

// 	var salt []byte
// 	var decodedHash []byte

// 	fmt.Sscanf(hash, "$argon2id$m=%d,t=%d,p=%d$%s$%s", &memory, &time, &threads, &salt, &decodedHash)

// 	hashBytes, _ := base64.RawStdEncoding.DecodeString(decodedHash)

// 	return argon2.IDKey([]byte(password), salt, time, memory, threads, keyLen) == hashBytes
// }

func GenerateRandomString(limit int) string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, limit)
	for i := range result {
		result[i] = chars[rand.Intn(len(chars))]
	}

	return string(result)
}
