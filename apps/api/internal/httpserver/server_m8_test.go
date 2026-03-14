package httpserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestM8AssuranceAndSimulationEndpoints(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	createTemplateReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates", "human:ten_01:user_admin:tenant_admin,approver", "idem_m8_tpl_create", map[string]any{
		"template_id": "act_code_patch_v1",
		"task_family": "code.patch",
		"name":        "Code Patch Contract",
	})
	rrTemplate := httptest.NewRecorder()
	h.ServeHTTP(rrTemplate, createTemplateReq)
	if rrTemplate.Code != http.StatusOK {
		t.Fatalf("create template expected 200 got %d body=%s", rrTemplate.Code, rrTemplate.Body.String())
	}

	createTemplateVersionReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates/act_code_patch_v1/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_tpl_version", map[string]any{
		"version": 1,
		"contract": map[string]any{
			"validation_tier":         "high_assurance",
			"required_critic_classes": []string{"plan_soundness"},
		},
	})
	rrTemplateVersion := httptest.NewRecorder()
	h.ServeHTTP(rrTemplateVersion, createTemplateVersionReq)
	if rrTemplateVersion.Code != http.StatusOK {
		t.Fatalf("create template version expected 200 got %d body=%s", rrTemplateVersion.Code, rrTemplateVersion.Body.String())
	}

	publishTemplateReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates/act_code_patch_v1/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_tpl_publish", map[string]any{"version": 1})
	rrTemplatePublish := httptest.NewRecorder()
	h.ServeHTTP(rrTemplatePublish, publishTemplateReq)
	if rrTemplatePublish.Code != http.StatusOK {
		t.Fatalf("publish template expected 200 got %d body=%s", rrTemplatePublish.Code, rrTemplatePublish.Body.String())
	}

	createCriticAReq := newJSONRequest(t, http.MethodPost, "/v1/critics", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_a", map[string]any{
		"critic_id":          "critic_plan_a",
		"name":               "Plan Critic A",
		"stage_name":         "plan",
		"task_family":        "code.patch",
		"critic_class":       "plan_soundness",
		"provider_family":    "provider_a",
		"failure_mode_class": "logic",
	})
	rrCriticA := httptest.NewRecorder()
	h.ServeHTTP(rrCriticA, createCriticAReq)
	if rrCriticA.Code != http.StatusOK {
		t.Fatalf("create critic A expected 200 got %d body=%s", rrCriticA.Code, rrCriticA.Body.String())
	}

	createCriticAVersionReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_plan_a/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_a_v", map[string]any{"version": 1, "config": map[string]any{"threshold": 0.8}})
	rrCriticAVersion := httptest.NewRecorder()
	h.ServeHTTP(rrCriticAVersion, createCriticAVersionReq)
	if rrCriticAVersion.Code != http.StatusOK {
		t.Fatalf("create critic A version expected 200 got %d body=%s", rrCriticAVersion.Code, rrCriticAVersion.Body.String())
	}

	publishCriticAReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_plan_a/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_a_publish", map[string]any{"version": 1})
	rrCriticAPublish := httptest.NewRecorder()
	h.ServeHTTP(rrCriticAPublish, publishCriticAReq)
	if rrCriticAPublish.Code != http.StatusOK {
		t.Fatalf("publish critic A expected 200 got %d body=%s", rrCriticAPublish.Code, rrCriticAPublish.Body.String())
	}

	createCriticBReq := newJSONRequest(t, http.MethodPost, "/v1/critics", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_b", map[string]any{
		"critic_id":          "critic_plan_b",
		"name":               "Plan Critic B",
		"stage_name":         "plan",
		"task_family":        "code.patch",
		"critic_class":       "plan_soundness",
		"provider_family":    "provider_b",
		"failure_mode_class": "safety",
	})
	rrCriticB := httptest.NewRecorder()
	h.ServeHTTP(rrCriticB, createCriticBReq)
	if rrCriticB.Code != http.StatusOK {
		t.Fatalf("create critic B expected 200 got %d body=%s", rrCriticB.Code, rrCriticB.Body.String())
	}

	createCriticBVersionReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_plan_b/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_b_v", map[string]any{"version": 1, "config": map[string]any{"threshold": 0.85}})
	rrCriticBVersion := httptest.NewRecorder()
	h.ServeHTTP(rrCriticBVersion, createCriticBVersionReq)
	if rrCriticBVersion.Code != http.StatusOK {
		t.Fatalf("create critic B version expected 200 got %d body=%s", rrCriticBVersion.Code, rrCriticBVersion.Body.String())
	}

	publishCriticBReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_plan_b/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_critic_b_publish", map[string]any{"version": 1})
	rrCriticBPublish := httptest.NewRecorder()
	h.ServeHTTP(rrCriticBPublish, publishCriticBReq)
	if rrCriticBPublish.Code != http.StatusOK {
		t.Fatalf("publish critic B expected 200 got %d body=%s", rrCriticBPublish.Code, rrCriticBPublish.Body.String())
	}

	validationRunReq := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_01/validation-runs", "human:ten_01:user_approver:approver,tenant_admin", "idem_m8_validation_run", map[string]any{
		"task_family":      "code.patch",
		"validation_tier":  "high_assurance",
		"template_id":      "act_code_patch_v1",
		"template_version": 1,
		"result_reference": "res_01",
	})
	rrValidationRun := httptest.NewRecorder()
	h.ServeHTTP(rrValidationRun, validationRunReq)
	if rrValidationRun.Code != http.StatusOK {
		t.Fatalf("create validation run expected 200 got %d body=%s", rrValidationRun.Code, rrValidationRun.Body.String())
	}

	validationRunID := jsonValue(t, rrValidationRun.Body.Bytes(), "validation_run_id")
	if validationRunID == "" {
		t.Fatalf("missing validation_run_id body=%s", rrValidationRun.Body.String())
	}

	getValidationReq := httptest.NewRequest(http.MethodGet, "/v1/validation-runs/"+validationRunID, nil)
	getValidationReq.Header.Set("Authorization", "Bearer human:ten_01:user_approver:approver")
	rrGetValidation := httptest.NewRecorder()
	h.ServeHTTP(rrGetValidation, getValidationReq)
	if rrGetValidation.Code != http.StatusOK {
		t.Fatalf("get validation run expected 200 got %d body=%s", rrGetValidation.Code, rrGetValidation.Body.String())
	}
	if !strings.Contains(rrGetValidation.Body.String(), "\"validation_tier\"") {
		t.Fatalf("expected validation_tier payload body=%s", rrGetValidation.Body.String())
	}

	stagesReq := httptest.NewRequest(http.MethodGet, "/v1/validation-runs/"+validationRunID+"/stages", nil)
	stagesReq.Header.Set("Authorization", "Bearer human:ten_01:user_approver:approver")
	rrStages := httptest.NewRecorder()
	h.ServeHTTP(rrStages, stagesReq)
	if rrStages.Code != http.StatusOK {
		t.Fatalf("list validation stages expected 200 got %d body=%s", rrStages.Code, rrStages.Body.String())
	}

	escalateReq := newJSONRequest(t, http.MethodPost, "/v1/validation-runs/"+validationRunID+"/escalate", "human:ten_01:user_approver:approver", "idem_m8_escalate", map[string]any{"reason": "manual_review"})
	rrEscalate := httptest.NewRecorder()
	h.ServeHTTP(rrEscalate, escalateReq)
	if rrEscalate.Code != http.StatusOK {
		t.Fatalf("escalate validation run expected 200 got %d body=%s", rrEscalate.Code, rrEscalate.Body.String())
	}

	createScenarioReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios", "human:ten_01:user_admin:tenant_admin", "idem_m8_scenario_create", map[string]any{
		"scenario_id": "scheduler_starvation_v1",
		"scope":       "tenant",
		"name":        "Scheduler Starvation",
		"goal":        "Detect starvation",
	})
	rrScenario := httptest.NewRecorder()
	h.ServeHTTP(rrScenario, createScenarioReq)
	if rrScenario.Code != http.StatusOK {
		t.Fatalf("create scenario expected 200 got %d body=%s", rrScenario.Code, rrScenario.Body.String())
	}

	createScenarioVersionReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios/scheduler_starvation_v1/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_scenario_version", map[string]any{
		"version": 1,
		"spec": map[string]any{
			"population_model": map[string]any{"workers": 20},
		},
	})
	rrScenarioVersion := httptest.NewRecorder()
	h.ServeHTTP(rrScenarioVersion, createScenarioVersionReq)
	if rrScenarioVersion.Code != http.StatusOK {
		t.Fatalf("create scenario version expected 200 got %d body=%s", rrScenarioVersion.Code, rrScenarioVersion.Body.String())
	}

	publishScenarioReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios/scheduler_starvation_v1/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_scenario_publish", map[string]any{"version": 1})
	rrScenarioPublish := httptest.NewRecorder()
	h.ServeHTTP(rrScenarioPublish, publishScenarioReq)
	if rrScenarioPublish.Code != http.StatusOK {
		t.Fatalf("publish scenario expected 200 got %d body=%s", rrScenarioPublish.Code, rrScenarioPublish.Body.String())
	}

	startRunReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/runs", "human:ten_01:user_admin:tenant_admin", "idem_m8_run_start", map[string]any{
		"scenario_id":      "scheduler_starvation_v1",
		"scenario_version": 1,
		"run_mode":         "deterministic_stub",
		"seed":             "seed_123",
		"scale_profile":    "ci",
		"budget_limit_jw":  10,
		"timebox_seconds":  60,
	})
	rrRun := httptest.NewRecorder()
	h.ServeHTTP(rrRun, startRunReq)
	if rrRun.Code != http.StatusAccepted {
		t.Fatalf("start simulation run expected 202 got %d body=%s", rrRun.Code, rrRun.Body.String())
	}
	runID := jsonValue(t, rrRun.Body.Bytes(), "run_id")
	if runID == "" {
		t.Fatalf("missing run_id body=%s", rrRun.Body.String())
	}

	getRunReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/runs/"+runID, nil)
	getRunReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetRun := httptest.NewRecorder()
	h.ServeHTTP(rrGetRun, getRunReq)
	if rrGetRun.Code != http.StatusOK {
		t.Fatalf("get simulation run expected 200 got %d body=%s", rrGetRun.Code, rrGetRun.Body.String())
	}

	for _, suffix := range []string{"events", "metrics", "findings", "artifacts"} {
		req := httptest.NewRequest(http.MethodGet, "/v1/simulation/runs/"+runID+"/"+suffix, nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("list run %s expected 200 got %d body=%s", suffix, rr.Code, rr.Body.String())
		}
	}

	promoteReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/runs/"+runID+"/promote-baseline", "human:ten_01:user_admin:tenant_admin", "idem_m8_promote", map[string]any{"reason": "ci_golden"})
	rrPromote := httptest.NewRecorder()
	h.ServeHTTP(rrPromote, promoteReq)
	if rrPromote.Code != http.StatusOK {
		t.Fatalf("promote baseline expected 200 got %d body=%s", rrPromote.Code, rrPromote.Body.String())
	}
	baselineID := jsonValue(t, rrPromote.Body.Bytes(), "baseline_id")
	if baselineID == "" {
		t.Fatalf("missing baseline_id body=%s", rrPromote.Body.String())
	}

	compareReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/compare", "human:ten_01:user_admin:tenant_admin", "idem_m8_compare", map[string]any{
		"candidate_run_id": runID,
		"baseline_id":      baselineID,
	})
	rrCompare := httptest.NewRecorder()
	h.ServeHTTP(rrCompare, compareReq)
	if rrCompare.Code != http.StatusOK {
		t.Fatalf("compare expected 200 got %d body=%s", rrCompare.Code, rrCompare.Body.String())
	}

	replayPendingReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/replays", "human:ten_01:user_admin:tenant_admin", "idem_m8_replay_pending", map[string]any{
		"source_type":      "incident_trace",
		"source_ref":       "inc_001",
		"sensitivity":      "sensitive",
		"approval_granted": false,
	})
	rrReplayPending := httptest.NewRecorder()
	h.ServeHTTP(rrReplayPending, replayPendingReq)
	if rrReplayPending.Code != http.StatusAccepted {
		t.Fatalf("pending replay expected 202 got %d body=%s", rrReplayPending.Code, rrReplayPending.Body.String())
	}
	if !strings.Contains(rrReplayPending.Body.String(), "pending_approval") {
		t.Fatalf("expected pending approval replay status body=%s", rrReplayPending.Body.String())
	}

	replayApprovedReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/replays", "human:ten_01:user_admin:tenant_admin", "idem_m8_replay_approved", map[string]any{
		"source_type":      "incident_trace",
		"source_ref":       "inc_002",
		"sensitivity":      "sensitive",
		"approval_granted": true,
	})
	rrReplayApproved := httptest.NewRecorder()
	h.ServeHTTP(rrReplayApproved, replayApprovedReq)
	if rrReplayApproved.Code != http.StatusAccepted {
		t.Fatalf("approved replay expected 202 got %d body=%s", rrReplayApproved.Code, rrReplayApproved.Body.String())
	}
	replayID := jsonValue(t, rrReplayApproved.Body.Bytes(), "replay_id")
	if replayID == "" {
		t.Fatalf("missing replay_id body=%s", rrReplayApproved.Body.String())
	}

	getReplayReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/replays/"+replayID, nil)
	getReplayReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetReplay := httptest.NewRecorder()
	h.ServeHTTP(rrGetReplay, getReplayReq)
	if rrGetReplay.Code != http.StatusOK {
		t.Fatalf("get replay expected 200 got %d body=%s", rrGetReplay.Code, rrGetReplay.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func jsonValue(t *testing.T, payload []byte, key string) string {
	t.Helper()
	obj := map[string]any{}
	if err := json.Unmarshal(payload, &obj); err != nil {
		t.Fatalf("unmarshal payload: %v body=%s", err, string(payload))
	}
	parts := strings.Split(key, ".")
	var current any = obj
	for _, part := range parts {
		next, ok := current.(map[string]any)
		if !ok {
			return ""
		}
		current, ok = next[part]
		if !ok {
			return ""
		}
	}
	if s, ok := current.(string); ok {
		return s
	}
	return ""
}
