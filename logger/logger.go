// Package logger provides a process-wide structured logger built on log/slog,
// with helpers for carrying a correlation ID through a context.Context.
package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

var (
	mu         sync.Mutex
	rootLogger *slog.Logger
)

// Init configures the process-wide logger for the given environment.
// Pass "dev" for a human-readable text handler at debug level; any other
// value selects a JSON handler at info level, suitable for production.
// Calling Init again replaces the previously configured logger.
func Init(env string) {
	var handler slog.Handler
	if env == "dev" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}

	mu.Lock()
	defer mu.Unlock()
	rootLogger = slog.New(handler)
}

// Get returns the process-wide logger. If Init was never called, it lazily
// initializes a production-safe JSON handler at info level.
func Get() *slog.Logger {
	mu.Lock()
	defer mu.Unlock()
	if rootLogger == nil {
		rootLogger = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		}))
	}
	return rootLogger
}

// GetWithCorrelationID returns the process-wide logger, enriched with the
// correlation ID from ctx (as the "correlation_id" attribute) when one is set.
func GetWithCorrelationID(ctx context.Context) *slog.Logger {
	logger := Get()
	if id, ok := GetCorrelationID(ctx); ok {
		logger = logger.With("correlation_id", id)
	}
	return logger
}

// Info logs msg at info level, including the correlation ID from ctx if present.
func Info(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).InfoContext(ctx, msg, args...)
}

// Debug logs msg at debug level, including the correlation ID from ctx if present.
func Debug(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).DebugContext(ctx, msg, args...)
}

// Error logs msg at error level, including the correlation ID from ctx if present.
func Error(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).ErrorContext(ctx, msg, args...)
}

// Warn logs msg at warn level, including the correlation ID from ctx if present.
func Warn(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).WarnContext(ctx, msg, args...)
}

// Fatal logs msg at error level and then terminates the process with exit
// code 1. WARNING: it calls os.Exit, which skips deferred function calls —
// open resources are not flushed or closed. Use it only at the top level of
// a program, never in library code or below deferred cleanup.
func Fatal(ctx context.Context, msg string, args ...any) {
	Error(ctx, msg, args...)
	os.Exit(1)
}

// ctxKey is the private type for context values owned by this package.
type ctxKey string

const correlationIDKey ctxKey = "correlation_id"

// WithCorrelationID returns a copy of ctx carrying the given correlation ID.
func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}

// GetCorrelationID returns the correlation ID stored in ctx by
// WithCorrelationID, and whether one was present.
func GetCorrelationID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id, true
	}
	return "", false
}
