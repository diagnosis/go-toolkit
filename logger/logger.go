package logger

import (
	"context"
	"log/slog"
	"os"
	"sync"
)

var (
	rootLogger *slog.Logger
	once       sync.Once
)

func Init() {
	var handler slog.Handler
	env := os.Getenv("APP_ENV")
	if env == "" {
		env = "dev"
	}
	if env == "dev" {
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelInfo,
		})
	}
	rootLogger = slog.New(handler)

}

func Get() *slog.Logger {
	once.Do(Init)
	return rootLogger
}
func GetWithCorrelationID(ctx context.Context) *slog.Logger {
	logger := Get()
	if id, ok := GetCorrelationID(ctx); ok {
		logger = logger.With("correlation_id", id)
	}
	return logger
}
func Info(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).InfoContext(ctx, msg, args...)
}
func Debug(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).DebugContext(ctx, msg, args...)
}
func Error(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).ErrorContext(ctx, msg, args...)
}
func Warn(ctx context.Context, msg string, args ...any) {
	GetWithCorrelationID(ctx).WarnContext(ctx, msg, args...)
}
func Fatal(ctx context.Context, msg string, args ...any) {
	Error(ctx, msg, args...)
	os.Exit(1)
}

// helpers
type ctxKey string

const correlationIDKey ctxKey = "correlation_id"

func WithCorrelationID(ctx context.Context, correlationID string) context.Context {
	return context.WithValue(ctx, correlationIDKey, correlationID)
}
func GetCorrelationID(ctx context.Context) (string, bool) {
	if id, ok := ctx.Value(correlationIDKey).(string); ok {
		return id, true
	}
	return "", false
}
