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
	if _, err := a.StartLease(context.Background(), workeradapter.Lease{}, workeradapter.TaskSpec{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
	if _, err := a.PollRun(context.Background(), workeradapter.RunHandle{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
	if _, err := a.CollectResult(context.Background(), workeradapter.RunHandle{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
	if err := a.CancelRun(context.Background(), workeradapter.RunHandle{}); err == nil {
		t.Fatalf("expected not implemented error")
	}
}
