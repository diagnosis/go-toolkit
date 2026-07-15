// Package secure provides security primitives: password hashing, JWT signing
// and verification, refresh-token generation and hashing, one-time passwords,
// HMAC signatures, and OAuth user-info fetchers for Google and GitHub.
package secure

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"golang.org/x/crypto/argon2"
)

const (
	saltLength     = 16
	argonTime      = 3
	argonMemory    = 64 * 1024
	argonThreads   = 4
	argonKeyLength = 32
)

// HashPassword hashes password with Argon2id using a random 16-byte salt and
// returns a self-describing encoded string ("argon2id$v=...$m=...$...") that
// embeds the parameters, salt, and hash for later verification. Passwords
// must be 1–128 characters.
func HashPassword(password string) (string, error) {
	if len(password) < 1 {
		return "", fmt.Errorf("%w: password cannot be empty", errors.New("invalid password"))
	}
	if len(password) > 128 {
		return "", fmt.Errorf("%w: password too long (max 128 chars)", errors.New("invalid password"))
	}

	salt := make([]byte, saltLength)
	_, err := rand.Read(salt)
	if err != nil {
		return "", err
	}
	hash := argon2.IDKey([]byte(password), salt, argonTime, argonMemory, argonThreads, argonKeyLength)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	return fmt.Sprintf("argon2id$v=%d$m=%d$t=%d$p=%d$%s$%s", argon2.Version, argonMemory, argonTime, argonThreads, encodedSalt, encodedHash), nil
}

// VerifyPassword checks password against an encoded hash produced by
// HashPassword, re-deriving the key with the parameters embedded in the hash
// and comparing in constant time. It returns whether the password matches,
// or an error if the encoded hash cannot be parsed.
func VerifyPassword(password, hash string) (bool, error) {
	hashSlice := strings.Split(hash, "$")
	if len(hashSlice) != 7 {
		return false, errors.New("failed to parse hash")
	}
	if hashSlice[0] != "argon2id" {
		return false, errors.New("not an argon2id hash")
	}
	// version
	versionStr := strings.TrimPrefix(hashSlice[1], "v=")
	version, err := strconv.ParseUint(versionStr, 10, 32)
	if err != nil {
		return false, errors.New("failed to parse version")
	}
	if uint32(version) != argon2.Version {
		return false, errors.New("unsupported version")
	}
	// parse params
	memoryStr := strings.TrimPrefix(hashSlice[2], "m=")
	timeStr := strings.TrimPrefix(hashSlice[3], "t=")
	threadsStr := strings.TrimPrefix(hashSlice[4], "p=")

	parsedMemory, err := strconv.ParseUint(memoryStr, 10, 32)
	if err != nil {
		return false, errors.New("failed to parse memory")
	}
	parsedTime, err := strconv.ParseUint(timeStr, 10, 32)
	if err != nil {
		return false, errors.New("failed to parse time")
	}
	parsedThreads, err := strconv.ParseUint(threadsStr, 10, 8)
	if err != nil {
		return false, errors.New("failed to parse threads")
	}

	// decode salt
	decodedSalt, err := base64.RawStdEncoding.DecodeString(hashSlice[5])
	if err != nil {
		return false, errors.New("failed to decode salt")
	}
	// decode hash
	decodedHash, err := base64.RawStdEncoding.DecodeString(hashSlice[6])
	if err != nil {
		return false, errors.New("failed to decode hash")
	}

	// generate hash with same params
	got := argon2.IDKey([]byte(password), decodedSalt, uint32(parsedTime), uint32(parsedMemory), uint8(parsedThreads), uint32(len(decodedHash)))
	match := subtle.ConstantTimeCompare(got, decodedHash) == 1
	return match, nil
}
