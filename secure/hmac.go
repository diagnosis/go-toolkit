package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// SignHMACSHA256 returns the HMAC-SHA256 signature of body using secret.
func SignHMACSHA256(secret, body []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(body)
	signature := h.Sum(nil)
	return signature
}

// HexSignHMACSHA256 returns the HMAC-SHA256 signature of body using
// secret, hex-encoded.
func HexSignHMACSHA256(secret, body []byte) string {
	signature := SignHMACSHA256(secret, body)
	return hex.EncodeToString(signature)
}

// VerifyHMACSHA256 reports whether sig is the valid HMAC-SHA256 signature
// of body under secret, comparing in constant time.
func VerifyHMACSHA256(secret, body, sig []byte) bool {
	expected := SignHMACSHA256(secret, body)
	return hmac.Equal(sig, expected)
}
