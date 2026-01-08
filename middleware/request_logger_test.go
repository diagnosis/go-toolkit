package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRequestLogger(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write([]byte("ok"))
	})
	loggerHandler := RequestLogger()(handler)
	req := httptest.NewRequest("GET", "/test-path", nil)
	w := httptest.NewRecorder()

	loggerHandler.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Errorf("expected 200, got %d", w.Code)
	}
	if w.Body.String() != "ok" {
		t.Errorf("failed to get response body")
	}
}
