package simulation

import (
	"context"
	"errors"
	"testing"
)

func TestScenarioRunBaselineCompareFlow(t *testing.T) {
	svc := New()
	ctx := context.Background()

	scenario, err := svc.CreateScenario(ctx, CreateScenarioRequest{
		TenantID:   "ten_01",
		ScenarioID: "scheduler_starvation_v1",
		Scope:      ScopeTenant,
		Name:       "Scheduler Starvation",
		Goal:       "Detect starvation",
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}

	version, err := svc.CreateScenarioVersion(ctx, "ten_01", scenario.ScenarioID, CreateScenarioVersionRequest{
		Version: 1,
		Spec: map[string]any{
			"workload_mix":       map[string]any{"code.patch": 0.7},
			"population_model":   map[string]any{"workers": 12},
			"seed_strategy":      "fixed",
			"success_criteria":   []string{"no_starvation"},
			"failure_thresholds": map[string]any{"starvation_ratio": 0.15},
		},
	})
	if err != nil {
		t.Fatalf("create scenario version: %v", err)
	}
	if err := svc.PublishScenario(ctx, "ten_01", scenario.ScenarioID, version.Version); err != nil {
		t.Fatalf("publish scenario: %v", err)
	}

	run, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      scenario.ScenarioID,
		ScenarioVersion: version.Version,
		RunMode:         RunModeDeterministicStub,
		Seed:            "seed_123",
		ScaleProfile:    "ci",
		BudgetLimitJW:   25,
		TimeboxSeconds:  60,
		PolicyOverrides: map[string]any{"scheduler_profile": "fair_v2"},
	})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	if run.Status != RunStatusCompleted {
		t.Fatalf("run status=%q want %q", run.Status, RunStatusCompleted)
	}

	metrics, err := svc.ListRunMetrics(ctx, "ten_01", run.RunID)
	if err != nil {
		t.Fatalf("list run metrics: %v", err)
	}
	if len(metrics) == 0 {
		t.Fatalf("expected metrics")
	}

	findings, err := svc.ListRunFindings(ctx, "ten_01", run.RunID)
	if err != nil {
		t.Fatalf("list run findings: %v", err)
	}
	if len(findings) == 0 {
		t.Fatalf("expected findings")
	}

	baseline, err := svc.PromoteBaseline(ctx, PromoteBaselineRequest{
		TenantID:   "ten_01",
		RunID:      run.RunID,
		Reason:     "golden_ci",
		PromotedBy: "user_admin",
	})
	if err != nil {
		t.Fatalf("promote baseline: %v", err)
	}

	compare, err := svc.Compare(ctx, CompareRequest{
		TenantID:         "ten_01",
		CandidateRunID:   run.RunID,
		BaselineID:       baseline.BaselineID,
		FailOnSeverities: []string{"high", "critical"},
	})
	if err != nil {
		t.Fatalf("compare: %v", err)
	}
	if compare.Verdict == "" {
		t.Fatalf("expected verdict")
	}
}

func TestReplayApprovalGuardrail(t *testing.T) {
	svc := New()
	ctx := context.Background()

	replay, err := svc.RequestReplay(ctx, RequestReplayRequest{
		TenantID:        "ten_01",
		SourceType:      "incident_trace",
		SourceRef:       "inc_001",
		Sensitivity:     "sensitive",
		RequestedBy:     "user_approver",
		ApprovalGranted: false,
	})
	if err != nil {
		t.Fatalf("request replay should return pending approval record: %v", err)
	}
	if replay.Status != ReplayStatusPendingApproval {
		t.Fatalf("status=%q want %q", replay.Status, ReplayStatusPendingApproval)
	}

	replay2, err := svc.RequestReplay(ctx, RequestReplayRequest{
		TenantID:        "ten_01",
		SourceType:      "incident_trace",
		SourceRef:       "inc_002",
		Sensitivity:     "sensitive",
		RequestedBy:     "user_approver",
		ApprovalGranted: true,
	})
	if err != nil {
		t.Fatalf("request replay with approval: %v", err)
	}
	if replay2.Status != ReplayStatusApproved {
		t.Fatalf("status=%q want %q", replay2.Status, ReplayStatusApproved)
	}
}

func TestSimulationListGetCancelCompareAndErrorPaths(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.CreateScenario(ctx, CreateScenarioRequest{TenantID: "", ScenarioID: "s1", Name: "x"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for missing tenant, got %v", err)
	}

	scenario, err := svc.CreateScenario(ctx, CreateScenarioRequest{
		TenantID:   "ten_01",
		ScenarioID: "approval_backlog_v1",
		Scope:      ScopeTenant,
		Name:       "Approval Backlog",
		Goal:       "Stress queue",
	})
	if err != nil {
		t.Fatalf("create scenario: %v", err)
	}
	if _, err := svc.CreateScenario(ctx, CreateScenarioRequest{
		TenantID:   "ten_01",
		ScenarioID: "approval_backlog_v1",
		Scope:      ScopeTenant,
		Name:       "Approval Backlog",
	}); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("expected duplicate scenario error, got %v", err)
	}
	if len(svc.ListScenarios(ctx, "ten_01")) != 1 {
		t.Fatalf("expected 1 scenario in list")
	}
	if _, err := svc.GetScenario(ctx, "ten_01", scenario.ScenarioID); err != nil {
		t.Fatalf("get scenario: %v", err)
	}

	if _, err := svc.CreateScenarioVersion(ctx, "ten_01", "missing", CreateScenarioVersionRequest{Version: 1, Spec: map[string]any{}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing scenario for version create, got %v", err)
	}
	if _, err := svc.CreateScenarioVersion(ctx, "ten_01", scenario.ScenarioID, CreateScenarioVersionRequest{Version: 1, Spec: map[string]any{"k": "v"}}); err != nil {
		t.Fatalf("create scenario version: %v", err)
	}
	if err := svc.PublishScenario(ctx, "ten_01", scenario.ScenarioID, 1); err != nil {
		t.Fatalf("publish scenario: %v", err)
	}

	if _, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      scenario.ScenarioID,
		ScenarioVersion: 1,
		RunMode:         "invalid_mode",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid run mode, got %v", err)
	}

	run, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      scenario.ScenarioID,
		ScenarioVersion: 1,
		RunMode:         RunModeDeterministicStub,
		Seed:            "seed_case_2",
	})
	if err != nil {
		t.Fatalf("start run: %v", err)
	}
	if len(svc.ListRuns(ctx, "ten_01", scenario.ScenarioID)) != 1 {
		t.Fatalf("expected 1 run in list")
	}
	if _, err := svc.GetRun(ctx, "ten_01", run.RunID); err != nil {
		t.Fatalf("get run: %v", err)
	}
	if _, err := svc.ListRunEvents(ctx, "ten_01", run.RunID); err != nil {
		t.Fatalf("list run events: %v", err)
	}
	if _, err := svc.ListRunArtifacts(ctx, "ten_01", run.RunID); err != nil {
		t.Fatalf("list run artifacts: %v", err)
	}

	// Cover cancel branch for non-completed run.
	k := tenantScoped("ten_01", run.RunID)
	tmp := svc.runs[k]
	tmp.Status = RunStatusRunning
	svc.runs[k] = tmp
	cancelled, err := svc.CancelRun(ctx, "ten_01", run.RunID)
	if err != nil {
		t.Fatalf("cancel run: %v", err)
	}
	if cancelled.Status != RunStatusCancelled {
		t.Fatalf("status=%q want %q", cancelled.Status, RunStatusCancelled)
	}

	// Restore completion status for baseline promotion.
	tmp = svc.runs[k]
	tmp.Status = RunStatusCompleted
	svc.runs[k] = tmp
	baseline, err := svc.PromoteBaseline(ctx, PromoteBaselineRequest{
		TenantID:   "ten_01",
		RunID:      run.RunID,
		Reason:     "golden",
		PromotedBy: "user_admin",
	})
	if err != nil {
		t.Fatalf("promote baseline: %v", err)
	}
	if len(svc.ListBaselines(ctx, "ten_01", scenario.ScenarioID)) != 1 {
		t.Fatalf("expected 1 baseline in list")
	}
	if _, err := svc.GetBaseline(ctx, "ten_01", baseline.BaselineID); err != nil {
		t.Fatalf("get baseline: %v", err)
	}

	compare, err := svc.Compare(ctx, CompareRequest{
		TenantID:         "ten_01",
		CandidateRunID:   run.RunID,
		BaselineID:       baseline.BaselineID,
		FailOnSeverities: []string{SeverityLow},
	})
	if err != nil {
		t.Fatalf("compare runs: %v", err)
	}
	if compare.Verdict == "" {
		t.Fatalf("expected compare verdict")
	}

	if _, err := svc.GetReplay(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found replay, got %v", err)
	}
	replay, err := svc.RequestReplay(ctx, RequestReplayRequest{
		TenantID:        "ten_01",
		SourceType:      "incident_trace",
		SourceRef:       "inc_003",
		Sensitivity:     "low",
		RequestedBy:     "user_admin",
		ApprovalGranted: true,
	})
	if err != nil {
		t.Fatalf("request replay approved: %v", err)
	}
	if _, err := svc.GetReplay(ctx, "ten_01", replay.ReplayID); err != nil {
		t.Fatalf("get replay: %v", err)
	}
}

func TestEnsureV1ScenarioLibraryAndModes(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if err := svc.EnsureV1ScenarioLibrary(ctx, ""); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for empty tenant, got %v", err)
	}
	if err := svc.EnsureV1ScenarioLibrary(ctx, "ten_01"); err != nil {
		t.Fatalf("seed v1 library: %v", err)
	}
	if err := svc.EnsureV1ScenarioLibrary(ctx, "ten_01"); err != nil {
		t.Fatalf("seed v1 library idempotent call: %v", err)
	}
	scenarios := svc.ListScenarios(ctx, "ten_01")
	if len(scenarios) < 11 {
		t.Fatalf("expected seeded scenarios, got %d", len(scenarios))
	}
	seeded, err := svc.GetScenario(ctx, "ten_01", "scheduler_starvation_v1")
	if err != nil {
		t.Fatalf("get seeded scenario: %v", err)
	}
	if seeded.Status != ScenarioStatusPublished || seeded.PublishedVersion != 1 {
		t.Fatalf("seeded scenario status/version unexpected: %+v", seeded)
	}

	custom, err := svc.CreateScenario(ctx, CreateScenarioRequest{
		TenantID:   "ten_01",
		ScenarioID: "mode_coverage_scn",
		Scope:      ScopeTenant,
		Name:       "Mode Coverage",
		Goal:       "Cover sampled/runtime modes",
	})
	if err != nil {
		t.Fatalf("create custom scenario: %v", err)
	}
	if _, err := svc.CreateScenarioVersion(ctx, "ten_01", custom.ScenarioID, CreateScenarioVersionRequest{Version: 1, Spec: map[string]any{"k": "v"}}); err != nil {
		t.Fatalf("create custom scenario version: %v", err)
	}
	if err := svc.PublishScenario(ctx, "ten_01", custom.ScenarioID, 1); err != nil {
		t.Fatalf("publish custom scenario: %v", err)
	}

	sampledRun, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      custom.ScenarioID,
		ScenarioVersion: 1,
		RunMode:         RunModeSampledSynthetic,
		Seed:            "sampled_seed",
	})
	if err != nil {
		t.Fatalf("start sampled synthetic run: %v", err)
	}
	if len(sampledRun.Metrics) < 4 {
		t.Fatalf("expected sampled mode metric set, got %d", len(sampledRun.Metrics))
	}
	foundSampledArtifact := false
	for _, artifact := range sampledRun.Artifacts {
		if artifact.Kind == "sampled_trace_json" {
			foundSampledArtifact = true
			break
		}
	}
	if !foundSampledArtifact {
		t.Fatalf("expected sampled_trace_json artifact")
	}

	runtimeRun, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      custom.ScenarioID,
		ScenarioVersion: 1,
		RunMode:         RunModeRuntimeBacked,
		Seed:            "runtime_seed",
	})
	if err != nil {
		t.Fatalf("start runtime backed run: %v", err)
	}
	foundRuntimeArtifact := false
	for _, artifact := range runtimeRun.Artifacts {
		if artifact.Kind == "runtime_trace_json" {
			foundRuntimeArtifact = true
			break
		}
	}
	if !foundRuntimeArtifact {
		t.Fatalf("expected runtime_trace_json artifact")
	}
}

func TestSimulationDashboard(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if err := svc.EnsureV1ScenarioLibrary(ctx, "ten_01"); err != nil {
		t.Fatalf("seed scenarios: %v", err)
	}

	runDeterministic, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      "scheduler_starvation_v1",
		ScenarioVersion: 1,
		RunMode:         RunModeDeterministicStub,
		Seed:            "dash_seed_1",
	})
	if err != nil {
		t.Fatalf("start deterministic run: %v", err)
	}
	runSampled, err := svc.StartRun(ctx, StartRunRequest{
		TenantID:        "ten_01",
		ScenarioID:      "verifier_collusion_v1",
		ScenarioVersion: 1,
		RunMode:         RunModeSampledSynthetic,
		Seed:            "dash_seed_2",
	})
	if err != nil {
		t.Fatalf("start sampled run: %v", err)
	}

	if _, err := svc.PromoteBaseline(ctx, PromoteBaselineRequest{
		TenantID:   "ten_01",
		RunID:      runDeterministic.RunID,
		Reason:     "dashboard_baseline",
		PromotedBy: "user_admin",
	}); err != nil {
		t.Fatalf("promote baseline: %v", err)
	}

	dashboard := svc.Dashboard(ctx, "ten_01", 10)
	if dashboard.TotalRuns != 2 {
		t.Fatalf("total runs=%d want 2", dashboard.TotalRuns)
	}
	if dashboard.RunModeCounts[RunModeDeterministicStub] != 1 || dashboard.RunModeCounts[RunModeSampledSynthetic] != 1 {
		t.Fatalf("unexpected run mode counts: %+v", dashboard.RunModeCounts)
	}
	if dashboard.BaselinesTotal != 1 {
		t.Fatalf("baselines total=%d want 1", dashboard.BaselinesTotal)
	}
	if len(dashboard.RiskHeatmap) == 0 {
		t.Fatalf("expected non-empty risk heatmap")
	}
	if dashboard.ScenarioCounts[runSampled.ScenarioID] == 0 {
		t.Fatalf("expected sampled scenario count")
	}
}
