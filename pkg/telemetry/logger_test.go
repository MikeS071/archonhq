package telemetry

import (
	"context"
	"testing"
)

func TestCorrelationContextHelpers(t *testing.T) {
	logger := NewLogger()
	if logger == nil {
		t.Fatalf("expected logger")
	}
	ctx := WithCorrelationID(context.Background(), "corr_123")
	if got := CorrelationIDFromContext(ctx); got != "corr_123" {
		t.Fatalf("unexpected correlation id: %s", got)
	}
}
