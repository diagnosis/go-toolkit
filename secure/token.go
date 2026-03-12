package secure

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims is the base JWT payload the toolkit signs and verifies.
// Projects that need additional fields (Role, Email, etc.) should embed this
// in their own claims struct and call jwt.ParseWithClaims directly.
type Claims struct {
	Sub string `json:"sub"`
	jwt.RegisteredClaims
}

// JWTConfig holds all parameters for a JWTSigner.
// Leeway is optional clock-skew tolerance applied during verification.
// If zero, defaults to 30 seconds. Set explicitly to 1 to disable (e.g. in tests).
type JWTConfig struct {
	AccessSecret       string
	RefreshSecret      string
	AccessTokenExpiry  time.Duration
	RefreshTokenExpiry time.Duration
	Issuer             string
	Audience           string
	Leeway             time.Duration
}

// JWTSigner manages access and refresh token signing with separate secrets.
// One instance handles both token types for a given role/service.
type JWTSigner struct {
	cfg    JWTConfig
	leeway time.Duration
}

// NewJWTSigner validates config and returns a JWTSigner.
func NewJWTSigner(cfg JWTConfig) (*JWTSigner, error) {
	if cfg.AccessSecret == "" {
		return nil, fmt.Errorf("access secret cannot be empty")
	}
	if cfg.RefreshSecret == "" {
		return nil, fmt.Errorf("refresh secret cannot be empty")
	}
	if cfg.AccessSecret == cfg.RefreshSecret {
		return nil, fmt.Errorf("access and refresh secrets must differ")
	}
	if cfg.AccessTokenExpiry <= 0 {
		return nil, fmt.Errorf("access token expiry must be positive")
	}
	if cfg.RefreshTokenExpiry <= 0 {
		return nil, fmt.Errorf("refresh token expiry must be positive")
	}
	if cfg.Issuer == "" {
		return nil, fmt.Errorf("issuer cannot be empty")
	}
	if cfg.Audience == "" {
		return nil, fmt.Errorf("audience cannot be empty")
	}

	leeway := cfg.Leeway
	if leeway == 0 {
		leeway = 30 * time.Second
	}

	return &JWTSigner{cfg: cfg, leeway: leeway}, nil
}

// SignAccess creates a signed access token for the given subject.
// sub is typically a UUID string.
func (s *JWTSigner) SignAccess(sub string) (string, error) {
	return s.sign(sub, s.cfg.AccessSecret, s.cfg.AccessTokenExpiry)
}

// SignRefresh creates a signed refresh token for the given subject.
func (s *JWTSigner) SignRefresh(sub string) (string, error) {
	return s.sign(sub, s.cfg.RefreshSecret, s.cfg.RefreshTokenExpiry)
}

func (s *JWTSigner) sign(sub, secret string, expiry time.Duration) (string, error) {
	if sub == "" {
		return "", fmt.Errorf("sub cannot be empty")
	}

	now := time.Now().UTC()
	claims := Claims{
		Sub: sub,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   sub,
			Issuer:    s.cfg.Issuer,
			Audience:  jwt.ClaimStrings{s.cfg.Audience},
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
		},
	}

	tok := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := tok.SignedString([]byte(secret))
	if err != nil {
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// VerifyAccess parses and validates an access token.
// Returns Claims on success, a descriptive error on any failure.
func (s *JWTSigner) VerifyAccess(tokenStr string) (*Claims, error) {
	return s.verify(tokenStr, s.cfg.AccessSecret)
}

// VerifyRefresh parses and validates a refresh token.
func (s *JWTSigner) VerifyRefresh(tokenStr string) (*Claims, error) {
	return s.verify(tokenStr, s.cfg.RefreshSecret)
}

func (s *JWTSigner) verify(tokenStr, secret string) (*Claims, error) {
	if tokenStr == "" {
		return nil, fmt.Errorf("token cannot be empty")
	}

	parser := jwt.NewParser(
		jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}),
		jwt.WithIssuedAt(),
		jwt.WithIssuer(s.cfg.Issuer),
		jwt.WithAudience(s.cfg.Audience),
		jwt.WithExpirationRequired(),
		jwt.WithLeeway(s.leeway),
	)

	var claims Claims
	token, err := parser.ParseWithClaims(tokenStr, &claims, func(t *jwt.Token) (any, error) {
		return []byte(secret), nil
	})

	if err != nil {
		switch {
		case errors.Is(err, jwt.ErrTokenExpired):
			return nil, fmt.Errorf("token expired: %w", err)
		case errors.Is(err, jwt.ErrTokenNotValidYet):
			return nil, fmt.Errorf("token not valid yet: %w", err)
		case errors.Is(err, jwt.ErrTokenMalformed):
			return nil, fmt.Errorf("token malformed: %w", err)
		case errors.Is(err, jwt.ErrTokenSignatureInvalid):
			return nil, fmt.Errorf("token signature invalid: %w", err)
		default:
			return nil, fmt.Errorf("token invalid: %w", err)
		}
	}

	if !token.Valid {
		return nil, fmt.Errorf("token is not valid")
	}

	return &claims, nil
}

// AccessExpiry returns the configured access token lifetime.
func (s *JWTSigner) AccessExpiry() time.Duration {
	return s.cfg.AccessTokenExpiry
}

// RefreshExpiry returns the configured refresh token lifetime.
func (s *JWTSigner) RefreshExpiry() time.Duration {
	return s.cfg.RefreshTokenExpiry
}
