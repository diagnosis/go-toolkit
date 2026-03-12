package secure

import (
	"strings"
	"testing"
	"time"
)

// validConfig returns a production-like config with default leeway (30s).
func validConfig() JWTConfig {
	return JWTConfig{
		AccessSecret:       "shopper-access-secret-32-chars!!",
		RefreshSecret:      "shopper-refresh-secret-32-chars!",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "pazar",
		Audience:           "pazar-api",
	}
}

// noLeewayConfig returns a config with leeway=1ns for expiry tests.
// Without this, the 30s default leeway swallows short-lived test tokens.
func noLeewayConfig() JWTConfig {
	cfg := validConfig()
	cfg.Leeway = 1 * time.Nanosecond
	return cfg
}

// ── NewJWTSigner ─────────────────────────────────────────────────────────────

func TestNewJWTSigner_Valid(t *testing.T) {
	s, err := NewJWTSigner(validConfig())
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if s == nil {
		t.Fatal("expected signer, got nil")
	}
}

func TestNewJWTSigner_DefaultLeeway(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())
	if s.leeway != 30*time.Second {
		t.Errorf("expected default leeway 30s, got %v", s.leeway)
	}
}

func TestNewJWTSigner_CustomLeeway(t *testing.T) {
	cfg := validConfig()
	cfg.Leeway = 5 * time.Second
	s, _ := NewJWTSigner(cfg)
	if s.leeway != 5*time.Second {
		t.Errorf("expected leeway 5s, got %v", s.leeway)
	}
}

func TestNewJWTSigner_InvalidConfigs(t *testing.T) {
	base := validConfig()
	tests := []struct {
		name string
		cfg  JWTConfig
	}{
		{"empty access secret", func() JWTConfig { c := base; c.AccessSecret = ""; return c }()},
		{"empty refresh secret", func() JWTConfig { c := base; c.RefreshSecret = ""; return c }()},
		{"same secrets", func() JWTConfig { c := base; c.RefreshSecret = c.AccessSecret; return c }()},
		{"zero access expiry", func() JWTConfig { c := base; c.AccessTokenExpiry = 0; return c }()},
		{"zero refresh expiry", func() JWTConfig { c := base; c.RefreshTokenExpiry = 0; return c }()},
		{"empty issuer", func() JWTConfig { c := base; c.Issuer = ""; return c }()},
		{"empty audience", func() JWTConfig { c := base; c.Audience = ""; return c }()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := NewJWTSigner(tt.cfg)
			if err == nil {
				t.Errorf("expected error for config: %s", tt.name)
			}
		})
	}
}

// ── SignAccess / SignRefresh ──────────────────────────────────────────────────

func TestSign_ValidSub(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())

	for _, fn := range []struct {
		name string
		sign func(string) (string, error)
	}{
		{"access", s.SignAccess},
		{"refresh", s.SignRefresh},
	} {
		t.Run(fn.name, func(t *testing.T) {
			tok, err := fn.sign("user-uuid-123")
			if err != nil {
				t.Fatalf("expected no error, got: %v", err)
			}
			if tok == "" {
				t.Fatal("expected non-empty token")
			}
			if len(strings.Split(tok, ".")) != 3 {
				t.Error("expected 3-segment JWT")
			}
		})
	}
}

func TestSign_EmptySub(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())

	if _, err := s.SignAccess(""); err == nil {
		t.Error("expected error for empty sub on access token")
	}
	if _, err := s.SignRefresh(""); err == nil {
		t.Error("expected error for empty sub on refresh token")
	}
}

// ── VerifyAccess ─────────────────────────────────────────────────────────────

func TestVerifyAccess_RoundTrip(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())

	tok, _ := s.SignAccess("customer-uuid-abc")
	claims, err := s.VerifyAccess(tok)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Sub != "customer-uuid-abc" {
		t.Errorf("expected sub customer-uuid-abc, got %s", claims.Sub)
	}
}

func TestVerifyAccess_WrongSecret(t *testing.T) {
	s1, _ := NewJWTSigner(validConfig())
	cfg2 := validConfig()
	cfg2.AccessSecret = "completely-different-access-secret"
	cfg2.RefreshSecret = "completely-different-refresh-secret"
	s2, _ := NewJWTSigner(cfg2)

	tok, _ := s1.SignAccess("user-uuid")
	_, err := s2.VerifyAccess(tok)
	if err == nil {
		t.Fatal("expected error when verifying with wrong secret")
	}
}

func TestVerifyAccess_Expired(t *testing.T) {
	// Use noLeewayConfig so the 30s default doesn't absorb the short expiry
	cfg := noLeewayConfig()
	cfg.AccessTokenExpiry = 1 * time.Millisecond
	s, _ := NewJWTSigner(cfg)

	tok, _ := s.SignAccess("user-uuid")
	time.Sleep(5 * time.Millisecond)

	_, err := s.VerifyAccess(tok)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if !strings.Contains(err.Error(), "expired") {
		t.Errorf("expected 'expired' in error message, got: %v", err)
	}
}

func TestVerifyAccess_Tampered(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())
	tok, _ := s.SignAccess("user-uuid")

	_, err := s.VerifyAccess(tok + "x")
	if err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestVerifyAccess_Empty(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())
	_, err := s.VerifyAccess("")
	if err == nil {
		t.Fatal("expected error for empty token")
	}
}

func TestVerifyAccess_Malformed(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())
	_, err := s.VerifyAccess("not.a.jwt")
	if err == nil {
		t.Fatal("expected error for malformed token")
	}
}

// ── VerifyRefresh ────────────────────────────────────────────────────────────

func TestVerifyRefresh_RoundTrip(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())

	tok, _ := s.SignRefresh("customer-uuid-abc")
	claims, err := s.VerifyRefresh(tok)
	if err != nil {
		t.Fatalf("verify failed: %v", err)
	}
	if claims.Sub != "customer-uuid-abc" {
		t.Errorf("expected sub customer-uuid-abc, got %s", claims.Sub)
	}
}

func TestVerifyRefresh_CannotVerifyAccessToken(t *testing.T) {
	// Access tokens must not be accepted as refresh tokens
	s, _ := NewJWTSigner(validConfig())
	accessTok, _ := s.SignAccess("user-uuid")

	_, err := s.VerifyRefresh(accessTok)
	if err == nil {
		t.Fatal("refresh verify should reject access tokens (different secret)")
	}
}

// ── Cross-role isolation ─────────────────────────────────────────────────────

func TestCrossRoleIsolation(t *testing.T) {
	// Simulates Pazar's two-signer setup: shopper vs admin
	shopperSigner, _ := NewJWTSigner(JWTConfig{
		AccessSecret:       "shopper-access-secret-unique-abc",
		RefreshSecret:      "shopper-refresh-secret-unique-xyz",
		AccessTokenExpiry:  15 * time.Minute,
		RefreshTokenExpiry: 7 * 24 * time.Hour,
		Issuer:             "pazar",
		Audience:           "pazar-api",
	})
	adminSigner, _ := NewJWTSigner(JWTConfig{
		AccessSecret:       "admin-access-secret-unique-abc!!",
		RefreshSecret:      "admin-refresh-secret-unique-xyz!!",
		AccessTokenExpiry:  8 * time.Hour,
		RefreshTokenExpiry: 24 * time.Hour,
		Issuer:             "pazar",
		Audience:           "pazar-api",
	})

	customerTok, _ := shopperSigner.SignAccess("customer-uuid")
	if _, err := adminSigner.VerifyAccess(customerTok); err == nil {
		t.Fatal("admin signer must not verify customer token")
	}

	adminTok, _ := adminSigner.SignAccess("admin-uuid")
	if _, err := shopperSigner.VerifyAccess(adminTok); err == nil {
		t.Fatal("shopper signer must not verify admin token")
	}
}

// ── Expiry helpers ───────────────────────────────────────────────────────────

func TestExpiry(t *testing.T) {
	s, _ := NewJWTSigner(validConfig())
	if s.AccessExpiry() != 15*time.Minute {
		t.Errorf("expected 15m access expiry, got %v", s.AccessExpiry())
	}
	if s.RefreshExpiry() != 7*24*time.Hour {
		t.Errorf("expected 7d refresh expiry, got %v", s.RefreshExpiry())
	}
}
