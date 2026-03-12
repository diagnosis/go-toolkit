package secure

import (
	"encoding/hex"
	"strings"
	"testing"
)

func TestGenerateRefreshToken(t *testing.T) {
	t.Run("generates non-empty token", func(t *testing.T) {
		token, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		if token == "" {
			t.Fatal("expected non-empty token")
		}
	})

	t.Run("token is valid hex", func(t *testing.T) {
		token, _ := GenerateRefreshToken()
		_, err := hex.DecodeString(token)
		if err != nil {
			t.Errorf("token is not valid hex: %v", err)
		}
	})

	t.Run("token is 64 chars (32 bytes hex-encoded)", func(t *testing.T) {
		token, _ := GenerateRefreshToken()
		if len(token) != 64 {
			t.Errorf("expected 64 chars, got %d", len(token))
		}
	})

	t.Run("tokens are unique", func(t *testing.T) {
		t1, _ := GenerateRefreshToken()
		t2, _ := GenerateRefreshToken()
		if t1 == t2 {
			t.Error("two generated tokens should not be equal")
		}
	})
}

func TestHashRefreshToken(t *testing.T) {
	t.Run("same input produces same hash", func(t *testing.T) {
		raw := "some-raw-token-value"
		h1 := HashRefreshToken(raw)
		h2 := HashRefreshToken(raw)
		if h1 != h2 {
			t.Error("same input should produce same hash")
		}
	})

	t.Run("different inputs produce different hashes", func(t *testing.T) {
		h1 := HashRefreshToken("token-a")
		h2 := HashRefreshToken("token-b")
		if h1 == h2 {
			t.Error("different inputs should produce different hashes")
		}
	})

	t.Run("hash is valid hex", func(t *testing.T) {
		hash := HashRefreshToken("any-token")
		_, err := hex.DecodeString(hash)
		if err != nil {
			t.Errorf("hash is not valid hex: %v", err)
		}
	})

	t.Run("hash is 64 chars (SHA-256 hex)", func(t *testing.T) {
		hash := HashRefreshToken("any-token")
		if len(hash) != 64 {
			t.Errorf("expected 64 chars, got %d", len(hash))
		}
	})

	t.Run("hash is lowercase hex", func(t *testing.T) {
		hash := HashRefreshToken("any-token")
		if hash != strings.ToLower(hash) {
			t.Error("hash should be lowercase hex")
		}
	})
}

func TestVerifyRefreshToken(t *testing.T) {
	t.Run("correct raw token verifies", func(t *testing.T) {
		raw, _ := GenerateRefreshToken()
		stored := HashRefreshToken(raw)

		if !VerifyRefreshToken(raw, stored) {
			t.Fatal("correct token should verify")
		}
	})

	t.Run("wrong token does not verify", func(t *testing.T) {
		raw, _ := GenerateRefreshToken()
		stored := HashRefreshToken(raw)

		other, _ := GenerateRefreshToken()
		if VerifyRefreshToken(other, stored) {
			t.Fatal("wrong token should not verify")
		}
	})

	t.Run("empty raw token does not verify", func(t *testing.T) {
		raw, _ := GenerateRefreshToken()
		stored := HashRefreshToken(raw)

		if VerifyRefreshToken("", stored) {
			t.Fatal("empty token should not verify")
		}
	})

	t.Run("empty stored hash does not verify", func(t *testing.T) {
		raw, _ := GenerateRefreshToken()
		if VerifyRefreshToken(raw, "") {
			t.Fatal("empty stored hash should not verify")
		}
	})

	t.Run("tampered stored hash does not verify", func(t *testing.T) {
		raw, _ := GenerateRefreshToken()
		stored := HashRefreshToken(raw)
		tampered := stored[:len(stored)-1] + "f"

		if VerifyRefreshToken(raw, tampered) {
			t.Fatal("tampered hash should not verify")
		}
	})

	t.Run("full round-trip with generated token", func(t *testing.T) {
		raw, err := GenerateRefreshToken()
		if err != nil {
			t.Fatalf("generate failed: %v", err)
		}

		stored := HashRefreshToken(raw)

		// What gets stored in DB is the hash — not the raw token
		if raw == stored {
			t.Error("raw token and hash should differ")
		}

		// Verify with correct token
		if !VerifyRefreshToken(raw, stored) {
			t.Error("round-trip verify failed")
		}

		// Verify with wrong token
		wrong, _ := GenerateRefreshToken()
		if VerifyRefreshToken(wrong, stored) {
			t.Error("wrong token should not pass round-trip")
		}
	})
}
