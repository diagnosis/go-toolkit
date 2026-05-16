package secure

import (
	"context"
	"net/http"
	"testing"
)

func TestFetchGitHubUserInfo_Valid(t *testing.T) {
	body := `{
	"id": 34,
	"login": "diagnosis",
	"name": "demirks",
	"avatar_url": "https://hithub.com/avatar",
	"email": "safa@test.com"
}`
	client := mockClient(http.StatusOK, body)
	info, err := FetchGitHubUserInfo(context.Background(), client)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.ID != 34 {
		t.Errorf("expected %d, got %d", 34, info.ID)
	}
	if info.Login != "diagnosis" {
		t.Errorf("expected %s, got %s", "safa@test.com", info.Login)
	}
	if info.Name != "demirks" {
		t.Errorf("expected %s, got %s", "demirks", info.Name)
	}
	if info.AvatarURL != "https://hithub.com/avatar" {
		t.Errorf("expected %s, got %s", "https://hithub.com/avatar", info.AvatarURL)
	}
	if info.Email != "safa@test.com" {
		t.Errorf("expected %s, got %s", "safa@test.com", info.Email)
	}
}

func TestFetchGitHubUserInfo_NonOkStatus(t *testing.T) {
	client := mockClient(http.StatusUnauthorized, `{"error": "unauthorized"}`)
	_, err := FetchGitHubUserInfo(context.Background(), client)
	if err == nil {
		t.Fatal("expected error, got: no error")
	}
}

func TestFetchGitHubUserInfo_InvalidJSON(t *testing.T) {
	body := `not json at all`
	client := mockClient(http.StatusOK, body)
	_, err := FetchGitHubUserInfo(context.Background(), client)
	if err == nil {
		t.Fatal("expected error, got: no error")
	}

}

func TestFetchGithubUserInfo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	body := `{
	"id": 34,
	"login": "diagnosis",
	"name": "demirks",
	"avatar_url": "https://hithub.com/avatar",
	"email": "safa@test.com"
}`
	client := mockClient(http.StatusOK, body)
	_, err := FetchGitHubUserInfo(ctx, client)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
