package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	Init("dev")
	log := Get()
	if log == nil {
		t.Error("it should initialize logger")
	}

	Init("prod")
	log = Get()
	if log == nil {
		t.Error("it should initialize logger for prod")
	}
}

func TestGet_LazyDefault(t *testing.T) {
	mu.Lock()
	rootLogger = nil
	mu.Unlock()

	log := Get()
	if log == nil {
		t.Error("Get should lazily initialize a default logger")
	}
	if !log.Enabled(context.Background(), slog.LevelInfo) {
		t.Error("default logger should log at info level")
	}
	if log.Enabled(context.Background(), slog.LevelDebug) {
		t.Error("default logger should not log at debug level")
	}
}

func TestGetCorrelationID(t *testing.T) {

	var buff bytes.Buffer

	handler := slog.NewTextHandler(&buff, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	rootLogger = slog.New(handler)

	ctx := context.Background()
	ctx = WithCorrelationID(ctx, "test-correlation-123")

	Info(ctx, "test message")

	output := buff.String()
	if !strings.Contains(output, "test-correlation-123") {
		t.Error("log should contain correlationID", "output:", output)
	}
}
