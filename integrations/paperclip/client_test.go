package paperclip

import (
	"context"
	"strings"
	"testing"
	"time"
)

func TestNoopConnectorPushProjection(t *testing.T) {
	connector := NoopConnector{}

	t.Run("requires tenant id", func(t *testing.T) {
		_, err := connector.PushProjection(context.Background(), ProjectionPayload{
			TenantID:      "   ",
			GeneratedAt:   time.Unix(1700000000, 0),
			SourceOfTruth: "postgres",
		})
		if err == nil {
			t.Fatalf("expected tenant_id validation error")
		}
	})

	t.Run("requires postgres source of truth", func(t *testing.T) {
		_, err := connector.PushProjection(context.Background(), ProjectionPayload{
			TenantID:      "ten_01",
			GeneratedAt:   time.Unix(1700000000, 0),
			SourceOfTruth: "paperclip",
		})
		if err == nil {
			t.Fatalf("expected source_of_truth validation error")
		}
	})

	t.Run("returns deterministic external ref shape", func(t *testing.T) {
		ref, err := connector.PushProjection(context.Background(), ProjectionPayload{
			TenantID:      "ten_01",
			GeneratedAt:   time.Date(2026, time.January, 2, 3, 4, 5, 0, time.FixedZone("AEST", 10*3600)),
			SourceOfTruth: "postgres",
		})
		if err != nil {
			t.Fatalf("push projection: %v", err)
		}
		if !strings.HasPrefix(ref, "paperclip_stub_") {
			t.Fatalf("expected paperclip_stub_ prefix got %q", ref)
		}
		if ref != "paperclip_stub_20260101T170405Z" {
			t.Fatalf("unexpected external ref %q", ref)
		}
	})
}
