package secure

import (
	"crypto/hmac"
	"encoding/hex"
	"testing"
)

func TestSignHMACSHA256_Deterministic(t *testing.T) {
	secret := []byte("my-secret")
	body := []byte("my-body")

	sig1 := SignHMACSHA256(secret, body)
	sig2 := SignHMACSHA256(secret, body)

	if !hmac.Equal(sig1, sig2) {
		t.Error("same input should produce same signature")
	}
}

func TestVerifyHMACSHA256_ValidSignature(t *testing.T) {
	secret := []byte("my-secret")
	body := []byte("my-body")
	sig := SignHMACSHA256(secret, body)
	if !VerifyHMACSHA256(secret, body, sig) {
		t.Error("Valid Signature failed")
	}
}

func TestVerifyHMACSHA256_WrongSecret(t *testing.T) {
	secret := []byte("my-secret")
	body := []byte("my-body")
	sig := SignHMACSHA256(secret, body)
	if VerifyHMACSHA256([]byte("different-secret"), body, sig) {
		t.Error("Wrong Signature pass")
	}
}

func TestVerifyHMACSHA256_TamperedBody(t *testing.T) {
	secret := []byte("my-secret")
	body := []byte("tampered-body")
	sig := SignHMACSHA256(secret, body)
	if VerifyHMACSHA256([]byte("my-secret"), []byte("my-body"), sig) {
		t.Error("Wrong Body pass")
	}
}

func TestHexSignHMACSHA256_ValidHex(t *testing.T) {
	secret := []byte("my-secret")
	body := []byte("my-body")
	hx := HexSignHMACSHA256(secret, body)
	if len(hx) != 64 {
		t.Error("Wrong hex length")
	}
	if _, err := hex.DecodeString(hx); err != nil {
		t.Error("Invalid hex")
	}
}
