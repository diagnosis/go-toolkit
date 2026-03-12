package secure

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

const refreshTokenBytes = 32 // 256 bits of entropy

// GenerateRefreshToken returns a cryptographically random hex string (64 chars).
// This is the raw token — send it to the client, store only its hash.
func GenerateRefreshToken() (string, error) {
	b := make([]byte, refreshTokenBytes)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate refresh token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// HashRefreshToken returns a SHA-256 hex digest of the raw token.
// Store this in the database — never the raw token itself.
func HashRefreshToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}

// VerifyRefreshToken checks a raw token against a stored hash in constant time.
// Returns true only if the raw token produces the same hash.
func VerifyRefreshToken(raw, storedHash string) bool {
	expected := HashRefreshToken(raw)
	// Convert both to bytes for constant-time comparison
	expectedBytes, err := hex.DecodeString(expected)
	if err != nil {
		return false
	}
	storedBytes, err := hex.DecodeString(storedHash)
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(expectedBytes, storedBytes) == 1
}
