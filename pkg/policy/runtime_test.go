package policy

import "testing"

func TestNormalizeExecutionPolicyDefaults(t *testing.T) {
	got, err := NormalizeExecutionPolicy(nil)
	if err != nil {
		t.Fatalf("normalize policy: %v", err)
	}

	if got.WorkspaceMode != "ephemeral" {
		t.Fatalf("expected ephemeral workspace mode")
	}
	if got.HermesMemoryWriteback {
		t.Fatalf("expected hermes memory writeback to be disabled")
	}
	if got.InferenceMode != "byok_only" {
		t.Fatalf("expected byok_only inference mode")
	}
	if len(got.AllowedBackends) != 1 || got.AllowedBackends[0] != "docker" {
		t.Fatalf("unexpected default backends: %#v", got.AllowedBackends)
	}
}

func TestNormalizeExecutionPolicyRejectsInvalids(t *testing.T) {
	if _, err := NormalizeExecutionPolicy(map[string]any{
		"allowed_backends": []any{"kubernetes"},
	}); err == nil {
		t.Fatalf("expected backend validation error")
	}

	if _, err := NormalizeExecutionPolicy(map[string]any{
		"network_policy": "open",
	}); err == nil {
		t.Fatalf("expected network policy validation error")
	}

	if _, err := NormalizeExecutionPolicy(map[string]any{
		"hermes_memory_writeback": true,
	}); err == nil {
		t.Fatalf("expected memory writeback validation error")
	}

	if _, err := NormalizeExecutionPolicy(map[string]any{
		"inference_mode": "managed",
	}); err == nil {
		t.Fatalf("expected inference mode validation error")
	}
}

func TestNormalizeExecutionPolicySnapshotAndTypedSlices(t *testing.T) {
	got, err := NormalizeExecutionPolicy(map[string]any{
		"allowed_backends":        []string{"ssh", "docker", "ssh"},
		"allowed_toolsets":        []string{"file", "web"},
		"network_policy":          "deny_all",
		"workspace_mode":          "ephemeral",
		"hermes_memory_writeback": false,
		"inference_mode":          "byok_only",
		"byok_provider_ref":       "cred_123",
	})
	if err != nil {
		t.Fatalf("normalize policy: %v", err)
	}
	if len(got.AllowedBackends) != 2 {
		t.Fatalf("expected deduplicated backends, got %#v", got.AllowedBackends)
	}
	if got.ByokProviderRef != "cred_123" {
		t.Fatalf("expected byok provider ref")
	}
	snapshot := got.Snapshot()
	if snapshot["workspace_mode"] != "ephemeral" {
		t.Fatalf("expected snapshot workspace mode")
	}
}
