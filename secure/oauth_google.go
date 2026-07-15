package secure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/diagnosis/go-toolkit/v2/apperr"
)

const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

// GoogleUserInfo is the subset of the Google OpenID Connect userinfo
// response used for sign-in.
type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

// GenerateStateToken returns a cryptographically random hex string for
// use as the OAuth state parameter.
func GenerateStateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

// FetchGoogleUserInfo fetches the authenticated user's profile from the
// Google userinfo endpoint and rejects accounts whose email is not
// verified. client must already carry OAuth credentials.
func FetchGoogleUserInfo(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", res.StatusCode)
	}

	var userinfo GoogleUserInfo
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&userinfo)
	if err != nil {
		return nil, apperr.BadRequest("bad request", "bad request", err)
	}

	// is email verified
	if !userinfo.EmailVerified {
		return nil, apperr.Validation("email is not verified", "email is not verified", err)
	}

	return &userinfo, nil

}
