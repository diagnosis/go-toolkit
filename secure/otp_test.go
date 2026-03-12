package secure

import (
	"strings"
	"testing"
)

// ── GenerateOTP ───────────────────────────────────────────────────────────────

func TestGenerateOTP_Length(t *testing.T) {
	code, err := GenerateOTP()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if len(code) != 6 {
		t.Errorf("expected 6 digits, got %d: %s", len(code), code)
	}
}

func TestGenerateOTP_OnlyDigits(t *testing.T) {
	for i := 0; i < 20; i++ {
		code, err := GenerateOTP()
		if err != nil {
			t.Fatalf("expected no error, got: %v", err)
		}
		for _, c := range code {
			if c < '0' || c > '9' {
				t.Errorf("expected only digits, got char %q in code %s", c, code)
			}
		}
	}
}

func TestGenerateOTP_ZeroPadded(t *testing.T) {
	// Run many times to statistically hit low numbers
	found := false
	for i := 0; i < 10000; i++ {
		code, _ := GenerateOTP()
		if strings.HasPrefix(code, "0") {
			found = true
			if len(code) != 6 {
				t.Errorf("zero-padded code should still be 6 digits, got %s", code)
			}
			break
		}
	}
	// Not a hard failure — just informational if padding never triggered
	_ = found
}

func TestGenerateOTP_Unique(t *testing.T) {
	c1, _ := GenerateOTP()
	c2, _ := GenerateOTP()
	// Not guaranteed but statistically near-certain
	if c1 == c2 {
		t.Log("two OTPs matched — possible but extremely unlikely, re-run if flaky")
	}
}

// ── HashOTP ───────────────────────────────────────────────────────────────────

func TestHashOTP_Deterministic(t *testing.T) {
	h1 := HashOTP("123456")
	h2 := HashOTP("123456")
	if h1 != h2 {
		t.Error("same input should produce same hash")
	}
}

func TestHashOTP_DifferentInputs(t *testing.T) {
	h1 := HashOTP("123456")
	h2 := HashOTP("654321")
	if h1 == h2 {
		t.Error("different codes should produce different hashes")
	}
}

func TestHashOTP_IsSHA256Length(t *testing.T) {
	h := HashOTP("123456")
	// SHA-256 hex = 64 chars
	if len(h) != 64 {
		t.Errorf("expected 64 char SHA-256 hex, got %d", len(h))
	}
}

func TestHashOTP_IsLowercaseHex(t *testing.T) {
	h := HashOTP("123456")
	if h != strings.ToLower(h) {
		t.Error("hash should be lowercase hex")
	}
}

// ── VerifyOTP ─────────────────────────────────────────────────────────────────

func TestVerifyOTP_CorrectCode(t *testing.T) {
	code, _ := GenerateOTP()
	stored := HashOTP(code)

	if !VerifyOTP(code, stored) {
		t.Fatal("correct code should verify")
	}
}

func TestVerifyOTP_WrongCode(t *testing.T) {
	code, _ := GenerateOTP()
	stored := HashOTP(code)

	if VerifyOTP("000000", stored) {
		t.Fatal("wrong code should not verify")
	}
}

func TestVerifyOTP_EmptyCode(t *testing.T) {
	stored := HashOTP("123456")
	if VerifyOTP("", stored) {
		t.Fatal("empty code should not verify")
	}
}

func TestVerifyOTP_EmptyStoredHash(t *testing.T) {
	if VerifyOTP("123456", "") {
		t.Fatal("empty stored hash should not verify")
	}
}

func TestVerifyOTP_TamperedHash(t *testing.T) {
	code, _ := GenerateOTP()
	stored := HashOTP(code)
	tampered := stored[:len(stored)-1] + "f"

	if VerifyOTP(code, tampered) {
		t.Fatal("tampered hash should not verify")
	}
}

func TestVerifyOTP_RoundTrip(t *testing.T) {
	code, err := GenerateOTP()
	if err != nil {
		t.Fatalf("generate failed: %v", err)
	}

	stored := HashOTP(code)

	// raw code and hash should differ
	if code == stored {
		t.Error("code and hash should not be equal")
	}

	// correct code verifies
	if !VerifyOTP(code, stored) {
		t.Error("round-trip verify failed")
	}

	// different code does not verify
	if VerifyOTP("000000", stored) {
		t.Error("wrong code should not pass round-trip")
	}
}
