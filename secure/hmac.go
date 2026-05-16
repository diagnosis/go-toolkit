package secure

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

func SignHMACSHA256(secret, body []byte) []byte {
	h := hmac.New(sha256.New, secret)
	h.Write(body)
	signature := h.Sum(nil)
	return signature
}

func HexSignHMACSHA256(secret, body []byte) string {
	signature := SignHMACSHA256(secret, body)
	return hex.EncodeToString(signature)
}

func VerifyHMACSHA256(secret, body, sig []byte) bool {
	expected := SignHMACSHA256(secret, body)
	return hmac.Equal(sig, expected)
}
