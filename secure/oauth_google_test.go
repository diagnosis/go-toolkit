package secure

import (
	"context"
	"io"
	"net/http"
	"strings"
	"testing"
)

// mockTransport implements http.RoundTripper for testing
// without making real HTTP calls
type mockTransport struct {
	statusCode int
	body       string
}

func (m *mockTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	if err := r.Context().Err(); err != nil {
		return nil, err
	}
	return &http.Response{
		StatusCode: m.statusCode,
		Body:       io.NopCloser(strings.NewReader(m.body)),
		Header:     make(http.Header),
	}, nil
}

func mockClient(statusCode int, body string) *http.Client {

	return &http.Client{Transport: &mockTransport{statusCode: statusCode, body: body}}
}

// ── GenerateStateToken ────────────────────────────────────────────────────────

func TestGenerateStateToken_NonEmpty(t *testing.T) {
	token, err := GenerateStateToken()
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestGenerateStateToken_IsHex(t *testing.T) {
	token, _ := GenerateStateToken()
	// 16 bytes hex encoded = 32 chars
	if len(token) != 32 {
		t.Errorf("expected 32 chars, got %d", len(token))
	}
}

func TestGenerateStateToken_Unique(t *testing.T) {
	t1, _ := GenerateStateToken()
	t2, _ := GenerateStateToken()
	if t1 == t2 {
		t.Error("two state tokens should not be equal")
	}
}

// ── FetchGoogleUserInfo ───────────────────────────────────────────────────────

func TestFetchGoogleUserInfo_Valid(t *testing.T) {
	body := `{
		"sub": "1234567890",
		"email": "safa@example.com",
		"name": "Safa D",
		"email_verified": true
	}`
	client := mockClient(http.StatusOK, body)

	info, err := FetchGoogleUserInfo(context.Background(), client)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if info.Sub != "1234567890" {
		t.Errorf("expected sub 1234567890, got %s", info.Sub)
	}
	if info.Email != "safa@example.com" {
		t.Errorf("expected email safa@example.com, got %s", info.Email)
	}
	if info.Name != "Safa D" {
		t.Errorf("expected name Safa D, got %s", info.Name)
	}
	if !info.EmailVerified {
		t.Error("expected email to be verified")
	}
}

func TestFetchGoogleUserInfo_UnverifiedEmail(t *testing.T) {
	body := `{
		"sub": "1234567890",
		"email": "safa@example.com",
		"name": "Safa D",
		"email_verified": false
	}`
	client := mockClient(http.StatusOK, body)

	_, err := FetchGoogleUserInfo(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for unverified email")
	}
}

func TestFetchGoogleUserInfo_NonOKStatus(t *testing.T) {
	client := mockClient(http.StatusUnauthorized, `{"error": "unauthorized"}`)

	_, err := FetchGoogleUserInfo(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for non-200 status")
	}
}

func TestFetchGoogleUserInfo_InvalidJSON(t *testing.T) {
	client := mockClient(http.StatusOK, `not json at all`)

	_, err := FetchGoogleUserInfo(context.Background(), client)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestFetchGoogleUserInfo_ContextCancelled(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	body := `{"sub":"123","email":"a@b.com","name":"A","email_verified":true}`
	client := mockClient(http.StatusOK, body)

	_, err := FetchGoogleUserInfo(ctx, client)
	if err == nil {
		t.Fatal("expected error for cancelled context")
	}
}
