package secure

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/diagnosis/go-toolkit/errors"
)

const googleUserInfoURL = "https://openidconnect.googleapis.com/v1/userinfo"

type GoogleUserInfo struct {
	Sub           string `json:"sub"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	EmailVerified bool   `json:"email_verified"`
}

func GenerateStateToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate state token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func FetchGoogleUserInfo(ctx context.Context, client *http.Client) (*GoogleUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, googleUserInfoURL, nil)
	if err != nil {
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("google userinfo returned status %d", res.StatusCode)
	}

	var userinfo GoogleUserInfo
	dec := json.NewDecoder(res.Body)
	err = dec.Decode(&userinfo)
	if err != nil {
		return nil, errors.BadRequest("bad request", "bad request", err)
	}

	//is email verified
	if !userinfo.EmailVerified {
		return nil, errors.Validation("email is not verified", "email is not verified", err)
	}

	return &userinfo, nil

}
