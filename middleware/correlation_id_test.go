package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/diagnosis/go-toolkit/logger"
)

func TestCorrelationID(t *testing.T) {
	var capturedID string
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		id, ok := logger.GetCorrelationID(ctx)
		if !ok {
			t.Error("failed to get correlation id from context")
		}
		capturedID = id
		w.WriteHeader(200)
	})

	corrHandler := CorrelationID()(handler)

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	corrHandler.ServeHTTP(w, req)

	if capturedID == "" {
		t.Error("correlation ID should be generated")
	}

	if w.Header().Get("X-Correlation-ID") == "" {
		t.Errorf("no correlation id set")
	}

	req2 := httptest.NewRequest("GET", "/test", nil)
	req2.Header.Set("X-Correlation-ID", "existing-id-123")
	w2 := httptest.NewRecorder()
	corrHandler.ServeHTTP(w2, req2)
	if capturedID != "existing-id-123" {
		t.Errorf("expected captured correlationID existing-id-123, got: %s", capturedID)
	}
	if w2.Header().Get("X-Correlation-ID") != "existing-id-123" {
		t.Errorf("response should echo correlation ID")
	}

}
