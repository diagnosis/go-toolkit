package secure

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math/big"
)

// GenerateOTP returns a cryptographically random six-digit one-time password,
// zero-padded (e.g. "004217").
func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	padded := fmt.Sprintf("%06d", n.Int64())
	return padded, nil
}

// HashOTP returns the SHA-256 hex digest of code. Store this digest instead
// of the raw code.
func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", sum[:])
}

// VerifyOTP checks a raw code against a stored hash produced by HashOTP,
// comparing in constant time.
func VerifyOTP(code, storedHash string) bool {
	hash := HashOTP(code)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(storedHash)) == 1
}
