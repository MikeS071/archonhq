package telemetry

import (
	"context"
	"log/slog"
	"os"
)

func NewLogger() *slog.Logger {
	return slog.New(slog.NewJSONHandler(os.Stdout, nil))
}

type ctxKey string

const correlationIDKey ctxKey = "correlation_id"

func WithCorrelationID(ctx context.Context, id string) context.Context {
	return context.WithValue(ctx, correlationIDKey, id)
}

func CorrelationIDFromContext(ctx context.Context) string {
	value, _ := ctx.Value(correlationIDKey).(string)
	return value
}
