package secure

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"fmt"
	"math/big"
)

func GenerateOTP() (string, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(1_000_000))
	if err != nil {
		return "", err
	}
	padded := fmt.Sprintf("%06d", n.Int64())
	return padded, nil
}

func HashOTP(code string) string {
	sum := sha256.Sum256([]byte(code))
	return fmt.Sprintf("%x", sum[:])
}

func VerifyOTP(code, storedHash string) bool {
	hash := HashOTP(code)
	return subtle.ConstantTimeCompare([]byte(hash), []byte(storedHash)) == 1
}
