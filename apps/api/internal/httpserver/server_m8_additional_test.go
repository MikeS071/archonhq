package httpserver

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	assurancesvc "github.com/MikeS071/archonhq/services/assurance"
	simulationsvc "github.com/MikeS071/archonhq/services/simulation"
)

func TestM8AdditionalReadAndErrorEndpoints(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	listTemplatesReq := httptest.NewRequest(http.MethodGet, "/v1/acceptance-contract-templates", nil)
	listTemplatesReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListTemplates := httptest.NewRecorder()
	h.ServeHTTP(rrListTemplates, listTemplatesReq)
	if rrListTemplates.Code != http.StatusOK {
		t.Fatalf("list templates expected 200 got %d body=%s", rrListTemplates.Code, rrListTemplates.Body.String())
	}

	getMissingTemplateReq := httptest.NewRequest(http.MethodGet, "/v1/acceptance-contract-templates/missing", nil)
	getMissingTemplateReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrMissingTemplate := httptest.NewRecorder()
	h.ServeHTTP(rrMissingTemplate, getMissingTemplateReq)
	if rrMissingTemplate.Code != http.StatusNotFound {
		t.Fatalf("missing template expected 404 got %d body=%s", rrMissingTemplate.Code, rrMissingTemplate.Body.String())
	}

	createTemplateReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_tpl", map[string]any{
		"template_id": "act_extra",
		"task_family": "code.patch",
		"name":        "Extra",
	})
	rrCreateTemplate := httptest.NewRecorder()
	h.ServeHTTP(rrCreateTemplate, createTemplateReq)
	if rrCreateTemplate.Code != http.StatusOK {
		t.Fatalf("create template expected 200 got %d body=%s", rrCreateTemplate.Code, rrCreateTemplate.Body.String())
	}

	createTemplateVersionReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates/act_extra/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_tpl_v", map[string]any{
		"version": 1,
		"contract": map[string]any{
			"validation_tier": "standard",
		},
	})
	rrCreateTemplateVersion := httptest.NewRecorder()
	h.ServeHTTP(rrCreateTemplateVersion, createTemplateVersionReq)
	if rrCreateTemplateVersion.Code != http.StatusOK {
		t.Fatalf("create template version expected 200 got %d body=%s", rrCreateTemplateVersion.Code, rrCreateTemplateVersion.Body.String())
	}

	publishTemplateReq := newJSONRequest(t, http.MethodPost, "/v1/acceptance-contract-templates/act_extra/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_tpl_pub", map[string]any{"version": 1})
	rrPublishTemplate := httptest.NewRecorder()
	h.ServeHTTP(rrPublishTemplate, publishTemplateReq)
	if rrPublishTemplate.Code != http.StatusOK {
		t.Fatalf("publish template expected 200 got %d body=%s", rrPublishTemplate.Code, rrPublishTemplate.Body.String())
	}

	getTemplateReq := httptest.NewRequest(http.MethodGet, "/v1/acceptance-contract-templates/act_extra", nil)
	getTemplateReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetTemplate := httptest.NewRecorder()
	h.ServeHTTP(rrGetTemplate, getTemplateReq)
	if rrGetTemplate.Code != http.StatusOK {
		t.Fatalf("get template expected 200 got %d body=%s", rrGetTemplate.Code, rrGetTemplate.Body.String())
	}

	createCriticReq := newJSONRequest(t, http.MethodPost, "/v1/critics", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_critic", map[string]any{
		"critic_id":          "critic_extra",
		"name":               "Critic",
		"stage_name":         "plan",
		"task_family":        "code.patch",
		"critic_class":       "plan_soundness",
		"provider_family":    "provider_x",
		"failure_mode_class": "logic",
	})
	rrCreateCritic := httptest.NewRecorder()
	h.ServeHTTP(rrCreateCritic, createCriticReq)
	if rrCreateCritic.Code != http.StatusOK {
		t.Fatalf("create critic expected 200 got %d body=%s", rrCreateCritic.Code, rrCreateCritic.Body.String())
	}

	createCriticVersionReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_extra/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_critic_v", map[string]any{"version": 1})
	rrCreateCriticVersion := httptest.NewRecorder()
	h.ServeHTTP(rrCreateCriticVersion, createCriticVersionReq)
	if rrCreateCriticVersion.Code != http.StatusOK {
		t.Fatalf("create critic version expected 200 got %d body=%s", rrCreateCriticVersion.Code, rrCreateCriticVersion.Body.String())
	}

	publishCriticReq := newJSONRequest(t, http.MethodPost, "/v1/critics/critic_extra/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_critic_pub", map[string]any{"version": 1})
	rrPublishCritic := httptest.NewRecorder()
	h.ServeHTTP(rrPublishCritic, publishCriticReq)
	if rrPublishCritic.Code != http.StatusOK {
		t.Fatalf("publish critic expected 200 got %d body=%s", rrPublishCritic.Code, rrPublishCritic.Body.String())
	}

	listCriticsReq := httptest.NewRequest(http.MethodGet, "/v1/critics", nil)
	listCriticsReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListCritics := httptest.NewRecorder()
	h.ServeHTTP(rrListCritics, listCriticsReq)
	if rrListCritics.Code != http.StatusOK {
		t.Fatalf("list critics expected 200 got %d body=%s", rrListCritics.Code, rrListCritics.Body.String())
	}

	getCriticReq := httptest.NewRequest(http.MethodGet, "/v1/critics/critic_extra", nil)
	getCriticReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetCritic := httptest.NewRecorder()
	h.ServeHTTP(rrGetCritic, getCriticReq)
	if rrGetCritic.Code != http.StatusOK {
		t.Fatalf("get critic expected 200 got %d body=%s", rrGetCritic.Code, rrGetCritic.Body.String())
	}

	startValidationReq := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_extra/validation-runs", "human:ten_01:user_approver:approver", "idem_m8_extra_validation", map[string]any{
		"task_family":     "code.patch",
		"validation_tier": "standard",
		"inline_contract": map[string]any{"k": "v"},
	})
	rrStartValidation := httptest.NewRecorder()
	h.ServeHTTP(rrStartValidation, startValidationReq)
	if rrStartValidation.Code != http.StatusOK {
		t.Fatalf("start validation expected 200 got %d body=%s", rrStartValidation.Code, rrStartValidation.Body.String())
	}
	validationRunID := jsonValue(t, rrStartValidation.Body.Bytes(), "validation_run_id")

	listTaskValidationReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_extra/validation-runs", nil)
	listTaskValidationReq.Header.Set("Authorization", "Bearer human:ten_01:user_approver:approver")
	rrListTaskValidation := httptest.NewRecorder()
	h.ServeHTTP(rrListTaskValidation, listTaskValidationReq)
	if rrListTaskValidation.Code != http.StatusOK {
		t.Fatalf("list task validation expected 200 got %d body=%s", rrListTaskValidation.Code, rrListTaskValidation.Body.String())
	}

	getValidationMissingReq := httptest.NewRequest(http.MethodGet, "/v1/validation-runs/missing", nil)
	getValidationMissingReq.Header.Set("Authorization", "Bearer human:ten_01:user_approver:approver")
	rrGetValidationMissing := httptest.NewRecorder()
	h.ServeHTTP(rrGetValidationMissing, getValidationMissingReq)
	if rrGetValidationMissing.Code != http.StatusNotFound {
		t.Fatalf("get missing validation expected 404 got %d body=%s", rrGetValidationMissing.Code, rrGetValidationMissing.Body.String())
	}

	escalateForbiddenReq := newJSONRequest(t, http.MethodPost, "/v1/validation-runs/"+validationRunID+"/escalate", "human:ten_01:user_dev:developer", "idem_m8_extra_escalate_forbidden", map[string]any{"reason": "x"})
	rrEscalateForbidden := httptest.NewRecorder()
	h.ServeHTTP(rrEscalateForbidden, escalateForbiddenReq)
	if rrEscalateForbidden.Code != http.StatusForbidden {
		t.Fatalf("escalate forbidden expected 403 got %d body=%s", rrEscalateForbidden.Code, rrEscalateForbidden.Body.String())
	}

	listScenariosReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/scenarios", nil)
	listScenariosReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListScenarios := httptest.NewRecorder()
	h.ServeHTTP(rrListScenarios, listScenariosReq)
	if rrListScenarios.Code != http.StatusOK {
		t.Fatalf("list scenarios expected 200 got %d body=%s", rrListScenarios.Code, rrListScenarios.Body.String())
	}

	getMissingScenarioReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/scenarios/missing", nil)
	getMissingScenarioReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetMissingScenario := httptest.NewRecorder()
	h.ServeHTTP(rrGetMissingScenario, getMissingScenarioReq)
	if rrGetMissingScenario.Code != http.StatusNotFound {
		t.Fatalf("missing scenario expected 404 got %d body=%s", rrGetMissingScenario.Code, rrGetMissingScenario.Body.String())
	}

	createScenarioReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_scenario", map[string]any{
		"scenario_id": "approval_backlog_v1",
		"scope":       "tenant",
		"name":        "Approval Backlog",
	})
	rrCreateScenario := httptest.NewRecorder()
	h.ServeHTTP(rrCreateScenario, createScenarioReq)
	if rrCreateScenario.Code != http.StatusOK {
		t.Fatalf("create scenario expected 200 got %d body=%s", rrCreateScenario.Code, rrCreateScenario.Body.String())
	}

	createScenarioVersionReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios/approval_backlog_v1/versions", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_scenario_v", map[string]any{"version": 1, "spec": map[string]any{"k": "v"}})
	rrCreateScenarioVersion := httptest.NewRecorder()
	h.ServeHTTP(rrCreateScenarioVersion, createScenarioVersionReq)
	if rrCreateScenarioVersion.Code != http.StatusOK {
		t.Fatalf("create scenario version expected 200 got %d body=%s", rrCreateScenarioVersion.Code, rrCreateScenarioVersion.Body.String())
	}

	publishScenarioReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/scenarios/approval_backlog_v1/publish", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_scenario_pub", map[string]any{"version": 1})
	rrPublishScenario := httptest.NewRecorder()
	h.ServeHTTP(rrPublishScenario, publishScenarioReq)
	if rrPublishScenario.Code != http.StatusOK {
		t.Fatalf("publish scenario expected 200 got %d body=%s", rrPublishScenario.Code, rrPublishScenario.Body.String())
	}

	startRunReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/runs", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_run", map[string]any{"scenario_id": "approval_backlog_v1", "scenario_version": 1, "run_mode": "deterministic_stub"})
	rrStartRun := httptest.NewRecorder()
	h.ServeHTTP(rrStartRun, startRunReq)
	if rrStartRun.Code != http.StatusAccepted {
		t.Fatalf("start simulation run expected 202 got %d body=%s", rrStartRun.Code, rrStartRun.Body.String())
	}
	runID := jsonValue(t, rrStartRun.Body.Bytes(), "run_id")

	listRunsReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/runs", nil)
	listRunsReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListRuns := httptest.NewRecorder()
	h.ServeHTTP(rrListRuns, listRunsReq)
	if rrListRuns.Code != http.StatusOK {
		t.Fatalf("list runs expected 200 got %d body=%s", rrListRuns.Code, rrListRuns.Body.String())
	}

	cancelRunReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/runs/"+runID+"/cancel", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_cancel", map[string]any{})
	rrCancelRun := httptest.NewRecorder()
	h.ServeHTTP(rrCancelRun, cancelRunReq)
	if rrCancelRun.Code != http.StatusOK {
		t.Fatalf("cancel run expected 200 got %d body=%s", rrCancelRun.Code, rrCancelRun.Body.String())
	}

	promoteReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/runs/"+runID+"/promote-baseline", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_promote", map[string]any{"reason": "golden"})
	rrPromote := httptest.NewRecorder()
	h.ServeHTTP(rrPromote, promoteReq)
	if rrPromote.Code != http.StatusOK {
		t.Fatalf("promote baseline expected 200 got %d body=%s", rrPromote.Code, rrPromote.Body.String())
	}
	baselineID := jsonValue(t, rrPromote.Body.Bytes(), "baseline_id")

	listBaselinesReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/baselines", nil)
	listBaselinesReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrListBaselines := httptest.NewRecorder()
	h.ServeHTTP(rrListBaselines, listBaselinesReq)
	if rrListBaselines.Code != http.StatusOK {
		t.Fatalf("list baselines expected 200 got %d body=%s", rrListBaselines.Code, rrListBaselines.Body.String())
	}

	getBaselineReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/baselines/"+baselineID, nil)
	getBaselineReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetBaseline := httptest.NewRecorder()
	h.ServeHTTP(rrGetBaseline, getBaselineReq)
	if rrGetBaseline.Code != http.StatusOK {
		t.Fatalf("get baseline expected 200 got %d body=%s", rrGetBaseline.Code, rrGetBaseline.Body.String())
	}

	compareMissingReq := newJSONRequest(t, http.MethodPost, "/v1/simulation/compare", "human:ten_01:user_admin:tenant_admin", "idem_m8_extra_compare_missing", map[string]any{"candidate_run_id": runID, "baseline_id": "missing"})
	rrCompareMissing := httptest.NewRecorder()
	h.ServeHTTP(rrCompareMissing, compareMissingReq)
	if rrCompareMissing.Code != http.StatusNotFound {
		t.Fatalf("compare missing baseline expected 404 got %d body=%s", rrCompareMissing.Code, rrCompareMissing.Body.String())
	}

	getReplayMissingReq := httptest.NewRequest(http.MethodGet, "/v1/simulation/replays/missing", nil)
	getReplayMissingReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGetReplayMissing := httptest.NewRecorder()
	h.ServeHTTP(rrGetReplayMissing, getReplayMissingReq)
	if rrGetReplayMissing.Code != http.StatusNotFound {
		t.Fatalf("get missing replay expected 404 got %d body=%s", rrGetReplayMissing.Code, rrGetReplayMissing.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM8Helpers(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})

	rr := httptest.NewRecorder()
	srv.writeAssuranceError(rr, "corr_1", "code_1", "message", assurancesvc.ErrAlreadyExists)
	if rr.Code != http.StatusConflict {
		t.Fatalf("assurance already exists expected 409 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeAssuranceError(rr, "corr_1", "code_1", "message", assurancesvc.ErrCriticDiversity)
	if rr.Code != http.StatusConflict || !strings.Contains(rr.Body.String(), "critic_diversity_required") {
		t.Fatalf("assurance critic diversity expected 409 conflict payload got %d body=%s", rr.Code, rr.Body.String())
	}

	rr = httptest.NewRecorder()
	srv.writeAssuranceError(rr, "corr_1", "code_1", "message", errors.New("boom"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("assurance fallback expected 500 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeSimulationError(rr, "corr_1", "code_1", "message", simulationsvc.ErrNotFound)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("simulation not found expected 404 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeSimulationError(rr, "corr_1", "code_1", "message", simulationsvc.ErrInvalidRequest)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("simulation invalid request expected 400 got %d", rr.Code)
	}

	rr = httptest.NewRecorder()
	srv.writeSimulationError(rr, "corr_1", "code_1", "message", errors.New("boom"))
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("simulation fallback expected 500 got %d", rr.Code)
	}

	if got := atoiQueryDefault("", 11); got != 11 {
		t.Fatalf("atoiQueryDefault empty=%d want 11", got)
	}
	if got := atoiQueryDefault("bad", 11); got != 11 {
		t.Fatalf("atoiQueryDefault bad=%d want 11", got)
	}
	if got := atoiQueryDefault("14", 11); got != 14 {
		t.Fatalf("atoiQueryDefault valid=%d want 14", got)
	}
}

func TestM8GuardrailErrorBranches(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	tests := []struct {
		name       string
		method     string
		path       string
		token      string
		idem       string
		body       any
		wantStatus int
	}{
		{name: "create template forbidden", method: http.MethodPost, path: "/v1/acceptance-contract-templates", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_1", body: map[string]any{"task_family": "code.patch", "name": "x"}, wantStatus: http.StatusForbidden},
		{name: "publish template invalid payload", method: http.MethodPost, path: "/v1/acceptance-contract-templates/act_missing/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_guard_2", body: map[string]any{}, wantStatus: http.StatusBadRequest},
		{name: "create critic forbidden", method: http.MethodPost, path: "/v1/critics", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_3", body: map[string]any{"stage_name": "plan"}, wantStatus: http.StatusForbidden},
		{name: "publish critic invalid payload", method: http.MethodPost, path: "/v1/critics/critic_missing/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_guard_4", body: map[string]any{}, wantStatus: http.StatusBadRequest},
		{name: "start validation forbidden", method: http.MethodPost, path: "/v1/tasks/task_1/validation-runs", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_5", body: map[string]any{"task_family": "code.patch"}, wantStatus: http.StatusForbidden},
		{name: "escalate validation invalid reason", method: http.MethodPost, path: "/v1/validation-runs/val_missing/escalate", token: "human:ten_01:user_admin:tenant_admin,approver", idem: "idem_m8_guard_6", body: map[string]any{}, wantStatus: http.StatusBadRequest},
		{name: "create simulation scenario forbidden", method: http.MethodPost, path: "/v1/simulation/scenarios", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_7", body: map[string]any{"name": "x"}, wantStatus: http.StatusForbidden},
		{name: "publish scenario forbidden", method: http.MethodPost, path: "/v1/simulation/scenarios/scn_1/publish", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_8", body: map[string]any{"version": 1}, wantStatus: http.StatusForbidden},
		{name: "start run forbidden", method: http.MethodPost, path: "/v1/simulation/runs", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_9", body: map[string]any{"scenario_id": "scn_1", "scenario_version": 1}, wantStatus: http.StatusForbidden},
		{name: "cancel run forbidden", method: http.MethodPost, path: "/v1/simulation/runs/run_1/cancel", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_10", body: map[string]any{}, wantStatus: http.StatusForbidden},
		{name: "promote baseline forbidden", method: http.MethodPost, path: "/v1/simulation/runs/run_1/promote-baseline", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_11", body: map[string]any{}, wantStatus: http.StatusForbidden},
		{name: "request replay forbidden", method: http.MethodPost, path: "/v1/simulation/replays", token: "human:ten_01:user_dev:developer", idem: "idem_m8_guard_12", body: map[string]any{"source_type": "incident_trace", "source_ref": "inc_1"}, wantStatus: http.StatusForbidden},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := newJSONRequest(t, tt.method, tt.path, tt.token, tt.idem, tt.body)
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != tt.wantStatus {
				t.Fatalf("%s expected %d got %d body=%s", tt.name, tt.wantStatus, rr.Code, rr.Body.String())
			}
		})
	}

	for _, path := range []string{
		"/v1/simulation/runs/run_missing/events",
		"/v1/simulation/runs/run_missing/metrics",
		"/v1/simulation/runs/run_missing/findings",
		"/v1/simulation/runs/run_missing/artifacts",
	} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("missing run collection for %s expected 404 got %d body=%s", path, rr.Code, rr.Body.String())
		}
	}
}

func TestM8InvalidJSONPayloads(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	tests := []struct {
		name  string
		path  string
		token string
		idem  string
	}{
		{name: "template create bad json", path: "/v1/acceptance-contract-templates", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_1"},
		{name: "template version bad json", path: "/v1/acceptance-contract-templates/tpl_1/versions", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_2"},
		{name: "template publish bad json", path: "/v1/acceptance-contract-templates/tpl_1/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_3"},
		{name: "critic create bad json", path: "/v1/critics", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_4"},
		{name: "critic version bad json", path: "/v1/critics/c_1/versions", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_5"},
		{name: "critic publish bad json", path: "/v1/critics/c_1/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_6"},
		{name: "validation start bad json", path: "/v1/tasks/task_1/validation-runs", token: "human:ten_01:user_admin:tenant_admin,approver", idem: "idem_m8_bad_7"},
		{name: "validation escalate bad json", path: "/v1/validation-runs/val_1/escalate", token: "human:ten_01:user_admin:tenant_admin,approver", idem: "idem_m8_bad_8"},
		{name: "scenario create bad json", path: "/v1/simulation/scenarios", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_9"},
		{name: "scenario version bad json", path: "/v1/simulation/scenarios/scn_1/versions", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_10"},
		{name: "scenario publish bad json", path: "/v1/simulation/scenarios/scn_1/publish", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_11"},
		{name: "run start bad json", path: "/v1/simulation/runs", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_12"},
		{name: "promote baseline bad json", path: "/v1/simulation/runs/run_1/promote-baseline", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_13"},
		{name: "compare bad json", path: "/v1/simulation/compare", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_14"},
		{name: "replay bad json", path: "/v1/simulation/replays", token: "human:ten_01:user_admin:tenant_admin", idem: "idem_m8_bad_15"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader("{"))
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("Idempotency-Key", tt.idem)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()
			h.ServeHTTP(rr, req)
			if rr.Code != http.StatusBadRequest {
				t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
			}
		})
	}
}
