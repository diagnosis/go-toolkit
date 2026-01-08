package logger

import (
	"bytes"
	"context"
	"log/slog"
	"strings"
	"testing"
)

func TestInit(t *testing.T) {
	Init()
	log := Get()
	if log == nil {
		t.Error("it should initialize logger")
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
