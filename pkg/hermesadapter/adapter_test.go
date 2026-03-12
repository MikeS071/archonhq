package hermesadapter

import (
	"context"
	"testing"

	"github.com/MikeS071/archonhq/pkg/workeradapter"
)

func TestAdapterBasics(t *testing.T) {
	a := New()
	if a.Name() != "hermes" {
		t.Fatalf("unexpected name")
	}
	if _, err := a.Capabilities(context.Background()); err != nil {
		t.Fatalf("capabilities: %v", err)
	}
	handle, err := a.StartLease(context.Background(), workeradapter.Lease{
		LeaseID:          "lease_01",
		TenantID:         "ten_01",
		NodeID:           "node_01",
		ExecutionPolicy:  map[string]any{"allowed_backends": []any{"docker"}},
		ByokProviderRef:  "cred_abc",
		WorkspaceRootRef: "ws_01",
	}, workeradapter.TaskSpec{
		TaskID:      "task_01",
		WorkspaceID: "ws_01",
		TaskFamily:  "research.extract",
		Title:       "Test task",
	})
	if err != nil {
		t.Fatalf("start lease: %v", err)
	}
	if handle.RunID == "" || handle.WorkspacePath == "" {
		t.Fatalf("expected run handle details")
	}
	status, err := a.PollRun(context.Background(), handle)
	if err != nil {
		t.Fatalf("poll run: %v", err)
	}
	if status.State != "running" {
		t.Fatalf("expected running state")
	}
	result, err := a.CollectResult(context.Background(), handle)
	if err != nil {
		t.Fatalf("collect result: %v", err)
	}
	if result.Signature == "" {
		t.Fatalf("expected result signature")
	}
	if err := a.CancelRun(context.Background(), handle); err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if _, err := a.PollRun(context.Background(), workeradapter.RunHandle{RunID: "missing"}); err == nil {
		t.Fatalf("expected missing run error")
	}
}
