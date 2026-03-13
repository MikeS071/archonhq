package paperclipconnector

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MikeS071/archonhq/integrations/paperclip"
)

type fakeConnector struct {
	externalRef string
	err         error
}

func (f fakeConnector) PushProjection(_ context.Context, _ paperclip.ProjectionPayload) (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.externalRef, nil
}

func TestServiceSyncDryRun(t *testing.T) {
	svc := New(fakeConnector{externalRef: "ignored"})
	svc.now = func() time.Time { return time.Unix(1700000000, 0).UTC() }

	got, err := svc.Sync(context.Background(), SyncRequest{
		SyncID:      "pcsync_1",
		InitiatedBy: "user_1",
		DryRun:      true,
		Payload: paperclip.ProjectionPayload{
			TenantID:      "ten_01",
			SourceOfTruth: "postgres",
		},
	})
	if err != nil {
		t.Fatalf("sync dry run failed: %v", err)
	}
	if got.Status != "dry_run" {
		t.Fatalf("expected dry_run got %s", got.Status)
	}
	if got.ExternalRef != "paperclip://dry-run/pcsync_1" {
		t.Fatalf("unexpected external ref: %s", got.ExternalRef)
	}
	if got.SurfaceCounts["approvals"] != 0 {
		t.Fatalf("unexpected approvals count: %d", got.SurfaceCounts["approvals"])
	}
}

func TestServiceSyncCompleted(t *testing.T) {
	svc := New(fakeConnector{externalRef: "paperclip_ref_1"})
	got, err := svc.Sync(context.Background(), SyncRequest{
		SyncID:      "pcsync_2",
		InitiatedBy: "user_1",
		Payload: paperclip.ProjectionPayload{
			TenantID:           "ten_01",
			SourceOfTruth:      "postgres",
			WorkspaceSummaries: []map[string]any{{"workspace_id": "ws_1"}},
			Approvals:          []map[string]any{{"approval_id": "apr_1"}},
			Fleet:              []map[string]any{{"node_id": "node_1"}},
		},
	})
	if err != nil {
		t.Fatalf("sync failed: %v", err)
	}
	if got.Status != "completed" {
		t.Fatalf("expected completed got %s", got.Status)
	}
	if got.ExternalRef != "paperclip_ref_1" {
		t.Fatalf("unexpected external ref: %s", got.ExternalRef)
	}
	if got.SurfaceCounts["workspace_summaries"] != 1 {
		t.Fatalf("unexpected workspace count: %d", got.SurfaceCounts["workspace_summaries"])
	}
}

func TestServiceSyncConnectorFailure(t *testing.T) {
	svc := New(fakeConnector{err: errors.New("boom")})
	_, err := svc.Sync(context.Background(), SyncRequest{
		SyncID:      "pcsync_3",
		InitiatedBy: "user_1",
		Payload: paperclip.ProjectionPayload{
			TenantID:      "ten_01",
			SourceOfTruth: "postgres",
		},
	})
	if err == nil {
		t.Fatalf("expected connector error")
	}
}
