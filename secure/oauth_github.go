package secure

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

// GithubUserInfo is the subset of the GitHub /user endpoint response used
// for sign-in.
type GithubUserInfo struct {
	ID        int64  `json:"id"`
	Login     string `json:"login"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Email     string `json:"email"`
}

// FetchGitHubUserInfo fetches the authenticated user's profile from the
// GitHub API. client must already carry OAuth credentials (e.g. one from
// oauth2.Config.Client).
func FetchGitHubUserInfo(ctx context.Context, client *http.Client) (*GithubUserInfo, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", "https://api.github.com/user", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data")
	}

	req.Header.Set("Accept", "application/vnd.github.v3+json")
	req.Header.Set("User-Agent", "go-toolkit")
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GitHub API returned non-200 status: %d", resp.StatusCode)
	}

	var userInfo GithubUserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &userInfo, nil
}
