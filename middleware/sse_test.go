package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNewSSEWriter_SetsHeaders(t *testing.T) {
	w := httptest.NewRecorder()
	_, err := NewSSEWriter(w)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if w.Header().Get("Content-Type") != "text/event-stream" {
		t.Error("expected text/event-stream content type")
	}
	if w.Header().Get("Cache-Control") != "no-cache" {
		t.Error("expected no-cache Cache-Control")
	}
	if w.Header().Get("Connection") != "keep-alive" {
		t.Error("expected keep-alive Connection")
	}
}

func TestSSEWriter_Send_Format(t *testing.T) {
	expected := `event: push
data: {"repo":"test"}

`
	w := httptest.NewRecorder()
	sse, _ := NewSSEWriter(w)
	if err := sse.Send("push", `{"repo":"test"}`); err != nil {
		t.Error("expected no error")
	}
	actual := w.Body.String()
	if actual != expected {
		t.Errorf("expected: %s, got: %s", expected, actual)
	}
}

type nonFlusher struct {
	http.ResponseWriter
}

func TestNewSSEWriter_NonFlusher(t *testing.T) {
	w := httptest.NewRecorder()
	sse := &nonFlusher{w}
	_, err := NewSSEWriter(sse)
	if err == nil {
		t.Error("expected error for non-flusher writer")
	}
}
