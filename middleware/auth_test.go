package middleware

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diagnosis/go-toolkit/v3/apperr"
	"github.com/diagnosis/go-toolkit/v3/responder"
)

func TestRequireAuth(t *testing.T) {
	validAuthFunc := func(r *http.Request) (string, error) {
		return "user-123", nil
	}

	var capturedUserID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, ok := GetUserID(r.Context())
		if !ok {
			t.Error("UserID should be in context")
		}
		capturedUserID = userID
		w.WriteHeader(200)
	})

	authHandler := RequireAuth(validAuthFunc)(handler)
	req := httptest.NewRequest("GET", "/protecteed", nil)
	w := httptest.NewRecorder()

	authHandler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("valid auth should return 20, got: %d", w.Code)
	}
	if capturedUserID != "user-123" {
		t.Errorf("userID should be user-123, got: %s", capturedUserID)
	}

	invalidAuthFunc := func(r *http.Request) (string, error) {
		return "", fmt.Errorf("invalid token")
	}

	authHandler2 := RequireAuth(invalidAuthFunc)(handler)
	req2 := httptest.NewRequest("GET", "/protected", nil)
	w2 := httptest.NewRecorder()

	authHandler2.ServeHTTP(w2, req2)
	if w2.Code != 401 {
		t.Errorf("Invalid auth should return 401, got %d", w2.Code)
	}
	if ct := w2.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %s", ct)
	}

	var errRes responder.ErrorResponse
	if err := json.Unmarshal(w2.Body.Bytes(), &errRes); err != nil {
		t.Fatalf("401 body should be the JSON error envelope, got: %s", w2.Body.String())
	}
	if errRes.Error.Status != apperr.CodeUnauthorized {
		t.Errorf("expected status %v, got %v", apperr.CodeUnauthorized, errRes.Error.Status)
	}
	if errRes.Error.Message != "Unauthorized" {
		t.Errorf("expected message Unauthorized, got %q", errRes.Error.Message)
	}
	if errRes.Error.Timestamp.IsZero() {
		t.Error("expected envelope timestamp to be set")
	}
}
