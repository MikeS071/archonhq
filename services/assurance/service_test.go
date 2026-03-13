package assurance

import (
	"context"
	"errors"
	"testing"
)

func TestTemplateVersionPublishAndSnapshot(t *testing.T) {
	svc := New()
	ctx := context.Background()

	tpl, err := svc.CreateTemplate(ctx, CreateTemplateRequest{
		TenantID:   "ten_01",
		TemplateID: "act_code_patch",
		TaskFamily: "code.patch",
		Name:       "Code Patch Contract",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if tpl.Status != StatusDraft {
		t.Fatalf("status=%q want %q", tpl.Status, StatusDraft)
	}

	ver, err := svc.CreateTemplateVersion(ctx, "ten_01", tpl.TemplateID, CreateTemplateVersionRequest{
		Version: 1,
		Contract: map[string]any{
			"validation_tier":         ValidationTierStandard,
			"required_critic_classes": []string{"output_correctness", "policy_compliance"},
		},
	})
	if err != nil {
		t.Fatalf("create template version: %v", err)
	}
	if ver.Version != 1 {
		t.Fatalf("version=%d want 1", ver.Version)
	}

	if err := svc.PublishTemplate(ctx, "ten_01", tpl.TemplateID, 1); err != nil {
		t.Fatalf("publish template: %v", err)
	}

	snap1, err := svc.AttachTaskContract(ctx, AttachTaskContractRequest{
		TenantID:       "ten_01",
		TaskID:         "task_01",
		TemplateID:     tpl.TemplateID,
		TemplateVer:    1,
		ValidationTier: ValidationTierStandard,
	})
	if err != nil {
		t.Fatalf("attach task contract: %v", err)
	}
	if snap1.ContractSource != ContractSourceTemplateRef {
		t.Fatalf("contract source=%q want %q", snap1.ContractSource, ContractSourceTemplateRef)
	}

	snap2, err := svc.AttachTaskContract(ctx, AttachTaskContractRequest{
		TenantID:       "ten_01",
		TaskID:         "task_01",
		ValidationTier: ValidationTierFast,
		InlineContract: map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("reattach should return immutable snapshot: %v", err)
	}
	if snap2.ValidationTier != ValidationTierStandard {
		t.Fatalf("validation tier changed unexpectedly: %q", snap2.ValidationTier)
	}
}

func TestValidationRunDiversityAndEscalation(t *testing.T) {
	svc := New()
	ctx := context.Background()

	_, err := svc.CreateCritic(ctx, CreateCriticRequest{
		TenantID:         "ten_01",
		CriticID:         "critic_plan_a",
		Name:             "Plan Critic A",
		StageName:        StagePlan,
		TaskFamily:       "code.patch",
		CriticClass:      "plan_soundness",
		ProviderFamily:   "provider_a",
		FailureModeClass: "logic",
	})
	if err != nil {
		t.Fatalf("create critic a: %v", err)
	}
	if _, err := svc.CreateCriticVersion(ctx, "ten_01", "critic_plan_a", CreateCriticVersionRequest{Version: 1, Config: map[string]any{"threshold": 0.8}}); err != nil {
		t.Fatalf("create critic version a: %v", err)
	}
	if err := svc.PublishCritic(ctx, "ten_01", "critic_plan_a", 1); err != nil {
		t.Fatalf("publish critic a: %v", err)
	}

	_, err = svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:        "ten_01",
		TaskID:          "task_01",
		TaskFamily:      "code.patch",
		ValidationTier:  ValidationTierHighAssurance,
		InlineContract:  map[string]any{"required_critic_classes": []string{"plan_soundness"}},
		ResultReference: "res_01",
	})
	if err == nil {
		t.Fatalf("expected critic diversity error")
	}

	_, err = svc.CreateCritic(ctx, CreateCriticRequest{
		TenantID:         "ten_01",
		CriticID:         "critic_plan_b",
		Name:             "Plan Critic B",
		StageName:        StagePlan,
		TaskFamily:       "code.patch",
		CriticClass:      "plan_soundness",
		ProviderFamily:   "provider_b",
		FailureModeClass: "safety",
	})
	if err != nil {
		t.Fatalf("create critic b: %v", err)
	}
	if _, err := svc.CreateCriticVersion(ctx, "ten_01", "critic_plan_b", CreateCriticVersionRequest{Version: 1, Config: map[string]any{"threshold": 0.82}}); err != nil {
		t.Fatalf("create critic version b: %v", err)
	}
	if err := svc.PublishCritic(ctx, "ten_01", "critic_plan_b", 1); err != nil {
		t.Fatalf("publish critic b: %v", err)
	}

	run, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:        "ten_01",
		TaskID:          "task_01",
		TaskFamily:      "code.patch",
		ValidationTier:  ValidationTierHighAssurance,
		InlineContract:  map[string]any{"required_critic_classes": []string{"plan_soundness"}},
		ResultReference: "res_01",
	})
	if err != nil {
		t.Fatalf("start validation run: %v", err)
	}
	if len(run.StageResults) == 0 {
		t.Fatalf("expected stage results")
	}

	escalated, err := svc.EscalateValidationRun(ctx, EscalateValidationRunRequest{
		TenantID:         "ten_01",
		ValidationRunID:  run.ValidationRunID,
		Reason:           "manual_review",
		EscalatedByActor: "user_approver",
	})
	if err != nil {
		t.Fatalf("escalate validation run: %v", err)
	}
	if escalated.Status != RunStatusEscalated {
		t.Fatalf("status=%q want %q", escalated.Status, RunStatusEscalated)
	}
}

func TestAssuranceListGetAndErrorPaths(t *testing.T) {
	svc := New()
	ctx := context.Background()

	if _, err := svc.CreateTemplate(ctx, CreateTemplateRequest{TenantID: "", TemplateID: "x", TaskFamily: "code.patch", Name: "x"}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for empty tenant, got %v", err)
	}

	tpl, err := svc.CreateTemplate(ctx, CreateTemplateRequest{
		TenantID:   "ten_01",
		TemplateID: "act_01",
		TaskFamily: "code.patch",
		Name:       "Contract",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if _, err := svc.CreateTemplateVersion(ctx, "ten_01", "missing", CreateTemplateVersionRequest{Version: 1, Contract: map[string]any{}}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found for missing template version create, got %v", err)
	}
	if _, err := svc.CreateTemplateVersion(ctx, "ten_01", tpl.TemplateID, CreateTemplateVersionRequest{Version: 1, Contract: map[string]any{"a": 1}}); err != nil {
		t.Fatalf("create template version: %v", err)
	}
	if err := svc.PublishTemplate(ctx, "ten_01", tpl.TemplateID, 1); err != nil {
		t.Fatalf("publish template: %v", err)
	}
	templates := svc.ListTemplates(ctx, "ten_01")
	if len(templates) != 1 {
		t.Fatalf("expected 1 template, got %d", len(templates))
	}
	if _, err := svc.GetTemplate(ctx, "ten_01", tpl.TemplateID); err != nil {
		t.Fatalf("get template: %v", err)
	}

	snap, err := svc.AttachTaskContract(ctx, AttachTaskContractRequest{
		TenantID:       "ten_01",
		TaskID:         "task_01",
		TemplateID:     tpl.TemplateID,
		TemplateVer:    1,
		ValidationTier: ValidationTierStandard,
		InlineContract: map[string]any{"override": true},
	})
	if err != nil {
		t.Fatalf("attach with override: %v", err)
	}
	if snap.ContractSource != ContractSourceTemplateRefWithOverride {
		t.Fatalf("contract source=%q want %q", snap.ContractSource, ContractSourceTemplateRefWithOverride)
	}
	if HashContract(snap.ContractSnapshot) == "" {
		t.Fatalf("hash contract should not be empty")
	}

	if _, err := svc.CreateCritic(ctx, CreateCriticRequest{
		TenantID:         "ten_01",
		CriticID:         "critic_1",
		Name:             "Critic 1",
		StageName:        "bad_stage",
		TaskFamily:       "code.patch",
		CriticClass:      "plan_soundness",
		ProviderFamily:   "provider_a",
		FailureModeClass: "logic",
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for bad stage, got %v", err)
	}

	critic, err := svc.CreateCritic(ctx, CreateCriticRequest{
		TenantID:         "ten_01",
		CriticID:         "critic_2",
		Name:             "Critic 2",
		StageName:        StagePlan,
		TaskFamily:       "code.patch",
		CriticClass:      "plan_soundness",
		ProviderFamily:   "provider_a",
		FailureModeClass: "logic",
	})
	if err != nil {
		t.Fatalf("create critic: %v", err)
	}
	if _, err := svc.CreateCriticVersion(ctx, "ten_01", critic.CriticID, CreateCriticVersionRequest{Version: 1, Config: map[string]any{}}); err != nil {
		t.Fatalf("create critic version: %v", err)
	}
	if err := svc.PublishCritic(ctx, "ten_01", critic.CriticID, 1); err != nil {
		t.Fatalf("publish critic: %v", err)
	}
	if len(svc.ListCritics(ctx, "ten_01")) != 1 {
		t.Fatalf("expected 1 critic")
	}
	if _, err := svc.GetCritic(ctx, "ten_01", critic.CriticID); err != nil {
		t.Fatalf("get critic: %v", err)
	}

	runFast, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:       "ten_01",
		TaskID:         "task_fast",
		TaskFamily:     "code.patch",
		ValidationTier: ValidationTierFast,
		InlineContract: map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("start fast run: %v", err)
	}
	if len(runFast.StageResults) == 0 {
		t.Fatalf("expected stage results for fast run")
	}

	runStandard, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:       "ten_01",
		TaskID:         "task_standard",
		TaskFamily:     "code.patch",
		ValidationTier: ValidationTierStandard,
		InlineContract: map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("start standard run: %v", err)
	}

	runs := svc.ListTaskValidationRuns(ctx, "ten_01", "task_standard")
	if len(runs) != 1 || runs[0].ValidationRunID != runStandard.ValidationRunID {
		t.Fatalf("unexpected task run listing: %+v", runs)
	}
	if _, err := svc.GetValidationRun(ctx, "ten_01", runStandard.ValidationRunID); err != nil {
		t.Fatalf("get validation run: %v", err)
	}
	if stages, err := svc.ListValidationStages(ctx, "ten_01", runStandard.ValidationRunID); err != nil || len(stages) == 0 {
		t.Fatalf("list validation stages err=%v len=%d", err, len(stages))
	}

	if _, err := svc.GetValidationRun(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found for missing run, got %v", err)
	}
	if _, err := svc.ListValidationStages(ctx, "ten_01", "missing"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected not found for missing stage list, got %v", err)
	}
}

func TestAssuranceAdditionalBranches(t *testing.T) {
	svc := New()
	ctx := context.Background()

	_, err := svc.AttachTaskContract(ctx, AttachTaskContractRequest{
		TenantID:       "ten_01",
		TaskID:         "task_invalid_tier",
		ValidationTier: "invalid",
	})
	if !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid tier error, got %v", err)
	}

	if _, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:       "ten_01",
		TaskID:         "task_missing_family",
		ValidationTier: ValidationTierStandard,
	}); !errors.Is(err, ErrInvalidRequest) {
		t.Fatalf("expected invalid request for missing task family, got %v", err)
	}

	tpl, err := svc.CreateTemplate(ctx, CreateTemplateRequest{
		TenantID:   "ten_01",
		TemplateID: "act_branch",
		TaskFamily: "code.patch",
		Name:       "Branch Contract",
	})
	if err != nil {
		t.Fatalf("create template: %v", err)
	}
	if _, err := svc.CreateTemplateVersion(ctx, "ten_01", tpl.TemplateID, CreateTemplateVersionRequest{Version: 1, Contract: map[string]any{"base": true}}); err != nil {
		t.Fatalf("create template version: %v", err)
	}

	_, err = svc.AttachTaskContract(ctx, AttachTaskContractRequest{
		TenantID:    "ten_01",
		TaskID:      "task_not_ready",
		TemplateID:  tpl.TemplateID,
		TemplateVer: 0,
	})
	if !errors.Is(err, ErrTemplateNotReady) {
		t.Fatalf("expected template not ready, got %v", err)
	}

	if err := svc.PublishTemplate(ctx, "ten_01", tpl.TemplateID, 1); err != nil {
		t.Fatalf("publish template: %v", err)
	}

	if _, err := svc.CreateCriticVersion(ctx, "ten_01", "missing_critic", CreateCriticVersionRequest{Version: 1}); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing critic on version create, got %v", err)
	}
	if err := svc.PublishCritic(ctx, "ten_01", "missing_critic", 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected missing critic on publish, got %v", err)
	}

	if _, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:        "ten_01",
		TaskID:          "task_tpl_branch",
		TaskFamily:      "code.patch",
		ValidationTier:  ValidationTierStandard,
		TemplateID:      tpl.TemplateID,
		TemplateVersion: 1,
		InlineContract:  map[string]any{"override": true},
	}); err != nil {
		t.Fatalf("start validation with template and override: %v", err)
	}
}

func TestValidationDashboard(t *testing.T) {
	svc := New()
	ctx := context.Background()

	runFast, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:       "ten_01",
		TaskID:         "task_dash_fast",
		TaskFamily:     "code.patch",
		ValidationTier: ValidationTierFast,
		InlineContract: map[string]any{"k": "v"},
	})
	if err != nil {
		t.Fatalf("start fast run: %v", err)
	}
	if _, err := svc.StartValidationRun(ctx, StartValidationRunRequest{
		TenantID:       "ten_01",
		TaskID:         "task_dash_std",
		TaskFamily:     "code.patch",
		ValidationTier: ValidationTierStandard,
		InlineContract: map[string]any{"k": "v"},
	}); err != nil {
		t.Fatalf("start standard run: %v", err)
	}

	if _, err := svc.EscalateValidationRun(ctx, EscalateValidationRunRequest{
		TenantID:         "ten_01",
		ValidationRunID:  runFast.ValidationRunID,
		Reason:           "manual_review",
		EscalatedByActor: "user_approver",
	}); err != nil {
		t.Fatalf("escalate run: %v", err)
	}

	dashboard := svc.ValidationDashboard(ctx, "ten_01", 10)
	if dashboard.TotalRuns != 2 {
		t.Fatalf("total runs=%d want 2", dashboard.TotalRuns)
	}
	if dashboard.StatusCounts[RunStatusEscalated] != 1 {
		t.Fatalf("expected one escalated run, got %d", dashboard.StatusCounts[RunStatusEscalated])
	}
	if dashboard.TierCounts[ValidationTierFast] != 1 || dashboard.TierCounts[ValidationTierStandard] != 1 {
		t.Fatalf("unexpected tier counts: %+v", dashboard.TierCounts)
	}
	if len(dashboard.EscalationQueue) != 1 {
		t.Fatalf("expected escalation queue length 1, got %d", len(dashboard.EscalationQueue))
	}
	if dashboard.EscalationResidualRate <= 0 {
		t.Fatalf("expected positive escalation residual rate, got %f", dashboard.EscalationResidualRate)
	}
}
