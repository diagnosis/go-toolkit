package middleware

import (
	"fmt"
	"net/http"

	"github.com/diagnosis/go-toolkit/v2/apperr"
)

// SSEWriter writes server-sent events to an underlying http.ResponseWriter,
// flushing after each event so it reaches the client immediately.
type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

// NewSSEWriter sets the required SSE response headers on w and returns an
// SSEWriter. It fails if w does not support http.Flusher, which streaming
// requires.
func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	// 1. Set required SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 2. Get The Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, apperr.Internal("streaming not supported", "streaming not supported")
	}
	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}, nil

}

// Send writes a single server-sent event with the given event name and data
// payload, then flushes it to the client.
func (s *SSEWriter) Send(event, data string) error {
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil

}
