package middleware

import (
	"fmt"
	"net/http"

	"github.com/diagnosis/go-toolkit/errors"
)

type SSEWriter struct {
	w       http.ResponseWriter
	flusher http.Flusher
}

func NewSSEWriter(w http.ResponseWriter) (*SSEWriter, error) {
	// 1. Set required SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	// 2. Get The Flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		return nil, errors.Internal("streaming not supported", "streaming not supported")
	}
	return &SSEWriter{
		w:       w,
		flusher: flusher,
	}, nil

}

func (s *SSEWriter) Send(event, data string) error {
	_, err := fmt.Fprintf(s.w, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	s.flusher.Flush()
	return nil

}
