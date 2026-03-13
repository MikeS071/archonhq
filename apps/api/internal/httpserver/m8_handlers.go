package httpserver

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	assurancesvc "github.com/MikeS071/archonhq/services/assurance"
	simulationsvc "github.com/MikeS071/archonhq/services/simulation"
)

type createAcceptanceTemplateRequest struct {
	TemplateID string `json:"template_id,omitempty"`
	TaskFamily string `json:"task_family"`
	Name       string `json:"name"`
}

type createAcceptanceTemplateVersionRequest struct {
	Version  int            `json:"version"`
	Contract map[string]any `json:"contract"`
}

type publishVersionRequest struct {
	Version int `json:"version"`
}

type createCriticRequest struct {
	CriticID         string `json:"critic_id,omitempty"`
	Name             string `json:"name"`
	StageName        string `json:"stage_name"`
	TaskFamily       string `json:"task_family"`
	CriticClass      string `json:"critic_class"`
	ProviderFamily   string `json:"provider_family"`
	FailureModeClass string `json:"failure_mode_class"`
	Enabled          *bool  `json:"enabled,omitempty"`
}

type createCriticVersionRequest struct {
	Version int            `json:"version"`
	Config  map[string]any `json:"config,omitempty"`
}

type startValidationRunRequest struct {
	TaskFamily      string         `json:"task_family"`
	ValidationTier  string         `json:"validation_tier,omitempty"`
	TemplateID      string         `json:"template_id,omitempty"`
	TemplateVersion int            `json:"template_version,omitempty"`
	InlineContract  map[string]any `json:"acceptance_contract,omitempty"`
	ResultReference string         `json:"result_reference,omitempty"`
}

type escalateValidationRunRequest struct {
	Reason string `json:"reason"`
}

type createSimulationScenarioRequest struct {
	ScenarioID string `json:"scenario_id,omitempty"`
	Scope      string `json:"scope,omitempty"`
	Name       string `json:"name"`
	Goal       string `json:"goal,omitempty"`
}

type createSimulationScenarioVersionRequest struct {
	Version int            `json:"version"`
	Spec    map[string]any `json:"spec,omitempty"`
}

type startSimulationRunRequest struct {
	ScenarioID      string         `json:"scenario_id"`
	ScenarioVersion int            `json:"scenario_version"`
	RunMode         string         `json:"run_mode,omitempty"`
	Seed            string         `json:"seed,omitempty"`
	ScaleProfile    string         `json:"scale_profile,omitempty"`
	BudgetLimitJW   float64        `json:"budget_limit_jw,omitempty"`
	TimeboxSeconds  int            `json:"timebox_seconds,omitempty"`
	PolicyOverrides map[string]any `json:"policy_overrides,omitempty"`
}

type promoteBaselineRequest struct {
	Reason string `json:"reason,omitempty"`
}

type compareSimulationRequest struct {
	CandidateRunID   string   `json:"candidate_run_id"`
	BaselineID       string   `json:"baseline_id"`
	FailOnSeverities []string `json:"fail_on_severities,omitempty"`
}

type requestReplayRequest struct {
	SourceType      string `json:"source_type"`
	SourceRef       string `json:"source_ref"`
	Sensitivity     string `json:"sensitivity,omitempty"`
	ApprovalGranted bool   `json:"approval_granted,omitempty"`
}

func (s *Server) handleCreateAcceptanceContractTemplateV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for acceptance template create.", corrID, nil)
		return
	}

	var req createAcceptanceTemplateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	templateID := strings.TrimSpace(req.TemplateID)
	if templateID == "" {
		templateID = "act_" + randomID(6)
	}
	template, err := s.assurance.CreateTemplate(r.Context(), assurancesvc.CreateTemplateRequest{
		TenantID:   actor.TenantID,
		TemplateID: templateID,
		TaskFamily: strings.TrimSpace(req.TaskFamily),
		Name:       strings.TrimSpace(req.Name),
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "acceptance_template_create_failed", "Failed to create acceptance template.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "acceptance_contract", template.TemplateID, "acceptance_contract.template_created", map[string]any{
		"task_family": template.TaskFamily,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"template":       template,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListAcceptanceContractTemplatesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "developer", "approver", "auditor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for acceptance template read.", corrID, nil)
		return
	}

	items := s.assurance.ListTemplates(r.Context(), actor.TenantID)
	writeJSON(w, http.StatusOK, map[string]any{
		"templates":      items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetAcceptanceContractTemplateV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	tpl, err := s.assurance.GetTemplate(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("template_id")))
	if err != nil {
		s.writeAssuranceError(w, corrID, "acceptance_template_not_found", "Acceptance template not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"template":       tpl,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateAcceptanceContractTemplateVersionV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for acceptance template version create.", corrID, nil)
		return
	}

	var req createAcceptanceTemplateVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	version, err := s.assurance.CreateTemplateVersion(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("template_id")), assurancesvc.CreateTemplateVersionRequest{
		Version:  req.Version,
		Contract: req.Contract,
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "acceptance_template_version_failed", "Failed to create acceptance template version.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"template_version": version,
		"correlation_id":   corrID,
	})
}

func (s *Server) handlePublishAcceptanceContractTemplateV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for acceptance template publish.", corrID, nil)
		return
	}

	var req publishVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if req.Version <= 0 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "version is required.", corrID, nil)
		return
	}

	if err := s.assurance.PublishTemplate(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("template_id")), req.Version); err != nil {
		s.writeAssuranceError(w, corrID, "acceptance_template_publish_failed", "Failed to publish acceptance template.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"template_id":       strings.TrimSpace(r.PathValue("template_id")),
		"published_version": req.Version,
		"status":            assurancesvc.StatusPublished,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleCreateCriticV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for critic create.", corrID, nil)
		return
	}

	var req createCriticRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	criticID := strings.TrimSpace(req.CriticID)
	if criticID == "" {
		criticID = "critic_" + randomID(6)
	}

	critic, err := s.assurance.CreateCritic(r.Context(), assurancesvc.CreateCriticRequest{
		TenantID:         actor.TenantID,
		CriticID:         criticID,
		Name:             strings.TrimSpace(req.Name),
		StageName:        strings.TrimSpace(req.StageName),
		TaskFamily:       strings.TrimSpace(req.TaskFamily),
		CriticClass:      strings.TrimSpace(req.CriticClass),
		ProviderFamily:   strings.TrimSpace(req.ProviderFamily),
		FailureModeClass: strings.TrimSpace(req.FailureModeClass),
		Enabled:          req.Enabled,
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "critic_create_failed", "Failed to create critic.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"critic":         critic,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListCriticsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	items := s.assurance.ListCritics(r.Context(), actor.TenantID)
	writeJSON(w, http.StatusOK, map[string]any{
		"critics":        items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetCriticV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	critic, err := s.assurance.GetCritic(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("critic_id")))
	if err != nil {
		s.writeAssuranceError(w, corrID, "critic_not_found", "Critic not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"critic":         critic,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateCriticVersionV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for critic version create.", corrID, nil)
		return
	}

	var req createCriticVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	version, err := s.assurance.CreateCriticVersion(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("critic_id")), assurancesvc.CreateCriticVersionRequest{
		Version: req.Version,
		Config:  req.Config,
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "critic_version_failed", "Failed to create critic version.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"critic_version": version,
		"correlation_id": corrID,
	})
}

func (s *Server) handlePublishCriticV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for critic publish.", corrID, nil)
		return
	}

	var req publishVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if req.Version <= 0 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "version is required.", corrID, nil)
		return
	}

	if err := s.assurance.PublishCritic(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("critic_id")), req.Version); err != nil {
		s.writeAssuranceError(w, corrID, "critic_publish_failed", "Failed to publish critic.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"critic_id":         strings.TrimSpace(r.PathValue("critic_id")),
		"published_version": req.Version,
		"status":            assurancesvc.StatusPublished,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleStartValidationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "tenant_admin", "platform_admin", "operator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for validation run create.", corrID, nil)
		return
	}

	var req startValidationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	run, err := s.assurance.StartValidationRun(r.Context(), assurancesvc.StartValidationRunRequest{
		TenantID:        actor.TenantID,
		TaskID:          strings.TrimSpace(r.PathValue("task_id")),
		TaskFamily:      strings.TrimSpace(req.TaskFamily),
		ValidationTier:  strings.TrimSpace(req.ValidationTier),
		TemplateID:      strings.TrimSpace(req.TemplateID),
		TemplateVersion: req.TemplateVersion,
		InlineContract:  req.InlineContract,
		ResultReference: strings.TrimSpace(req.ResultReference),
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "validation_run_failed", "Failed to start validation run.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "validation", run.ValidationRunID, "validation.run_started", map[string]any{
		"task_id":         run.TaskID,
		"validation_tier": run.ValidationTier,
	})
	if run.Status == assurancesvc.RunStatusCompleted {
		s.appendEvent(r, actor.TenantID, "validation", run.ValidationRunID, "validation.run_completed", map[string]any{"decision": run.Decision})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"validation_run_id": run.ValidationRunID,
		"validation_tier":   run.ValidationTier,
		"status":            run.Status,
		"decision":          run.Decision,
		"stage_results":     run.StageResults,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleListTaskValidationRunsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	runs := s.assurance.ListTaskValidationRuns(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("task_id")))
	writeJSON(w, http.StatusOK, map[string]any{
		"validation_runs": runs,
		"correlation_id":  corrID,
	})
}

func (s *Server) handleGetValidationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	run, err := s.assurance.GetValidationRun(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("validation_run_id")))
	if err != nil {
		s.writeAssuranceError(w, corrID, "validation_run_not_found", "Validation run not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"validation_run": run,
		"correlation_id": corrID,
	})
}

func (s *Server) handleValidationRunStagesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	stages, err := s.assurance.ListValidationStages(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("validation_run_id")))
	if err != nil {
		s.writeAssuranceError(w, corrID, "validation_run_not_found", "Validation run not found.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"validation_run_id": strings.TrimSpace(r.PathValue("validation_run_id")),
		"stages":            stages,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleEscalateValidationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for validation escalation.", corrID, nil)
		return
	}

	var req escalateValidationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	run, err := s.assurance.EscalateValidationRun(r.Context(), assurancesvc.EscalateValidationRunRequest{
		TenantID:         actor.TenantID,
		ValidationRunID:  strings.TrimSpace(r.PathValue("validation_run_id")),
		Reason:           strings.TrimSpace(req.Reason),
		EscalatedByActor: actor.ID,
	})
	if err != nil {
		s.writeAssuranceError(w, corrID, "validation_escalation_failed", "Failed to escalate validation run.", err)
		return
	}

	s.appendEvent(r, actor.TenantID, "validation", run.ValidationRunID, "validation.escalated", map[string]any{
		"reason": req.Reason,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"validation_run": run,
		"correlation_id": corrID,
	})
}

func (s *Server) handleValidationDashboardV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for validation dashboard read.", corrID, nil)
		return
	}

	limit := atoiQueryDefault(r.URL.Query().Get("limit"), 20)
	dashboard := s.assurance.ValidationDashboard(r.Context(), actor.TenantID, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"validation_dashboard": dashboard,
		"correlation_id":       corrID,
	})
}

func (s *Server) handleCreateSimulationScenarioV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation scenario create.", corrID, nil)
		return
	}

	var req createSimulationScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	scenarioID := strings.TrimSpace(req.ScenarioID)
	if scenarioID == "" {
		scenarioID = "scn_" + randomID(6)
	}

	scenario, err := s.simulation.CreateScenario(r.Context(), simulationsvc.CreateScenarioRequest{
		TenantID:   actor.TenantID,
		ScenarioID: scenarioID,
		Scope:      strings.TrimSpace(req.Scope),
		Name:       strings.TrimSpace(req.Name),
		Goal:       strings.TrimSpace(req.Goal),
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_scenario_create_failed", "Failed to create simulation scenario.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"scenario":       scenario,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListSimulationScenariosV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}
	items := s.simulation.ListScenarios(r.Context(), actor.TenantID)
	writeJSON(w, http.StatusOK, map[string]any{
		"scenarios":      items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetSimulationScenarioV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}
	scenario, err := s.simulation.GetScenario(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("scenario_id")))
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_scenario_not_found", "Simulation scenario not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"scenario":       scenario,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateSimulationScenarioVersionV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation scenario version create.", corrID, nil)
		return
	}

	var req createSimulationScenarioVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	version, err := s.simulation.CreateScenarioVersion(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("scenario_id")), simulationsvc.CreateScenarioVersionRequest{
		Version: req.Version,
		Spec:    req.Spec,
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_scenario_version_failed", "Failed to create simulation scenario version.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"scenario_version": version,
		"correlation_id":   corrID,
	})
}

func (s *Server) handlePublishSimulationScenarioV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation scenario publish.", corrID, nil)
		return
	}

	var req publishVersionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	if err := s.simulation.PublishScenario(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("scenario_id")), req.Version); err != nil {
		s.writeSimulationError(w, corrID, "simulation_scenario_publish_failed", "Failed to publish simulation scenario.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"scenario_id":       strings.TrimSpace(r.PathValue("scenario_id")),
		"published_version": req.Version,
		"status":            simulationsvc.ScenarioStatusPublished,
		"correlation_id":    corrID,
	})
}

func (s *Server) handleStartSimulationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation run create.", corrID, nil)
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}

	var req startSimulationRunRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	run, err := s.simulation.StartRun(r.Context(), simulationsvc.StartRunRequest{
		TenantID:        actor.TenantID,
		ScenarioID:      strings.TrimSpace(req.ScenarioID),
		ScenarioVersion: req.ScenarioVersion,
		RunMode:         strings.TrimSpace(req.RunMode),
		Seed:            strings.TrimSpace(req.Seed),
		ScaleProfile:    strings.TrimSpace(req.ScaleProfile),
		BudgetLimitJW:   req.BudgetLimitJW,
		TimeboxSeconds:  req.TimeboxSeconds,
		PolicyOverrides: req.PolicyOverrides,
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_run_failed", "Failed to start simulation run.", err)
		return
	}

	writeJSON(w, http.StatusAccepted, map[string]any{
		"run_id":         run.RunID,
		"status":         run.Status,
		"scenario_id":    run.ScenarioID,
		"run_mode":       run.RunMode,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListSimulationRunsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}
	items := s.simulation.ListRuns(r.Context(), actor.TenantID, strings.TrimSpace(r.URL.Query().Get("scenario_id")))
	writeJSON(w, http.StatusOK, map[string]any{
		"runs":           items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetSimulationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	run, err := s.simulation.GetRun(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("run_id")))
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_run_not_found", "Simulation run not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"run":            run,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCancelSimulationRunV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation run cancel.", corrID, nil)
		return
	}

	run, err := s.simulation.CancelRun(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("run_id")))
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_run_cancel_failed", "Failed to cancel simulation run.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"run":            run,
		"correlation_id": corrID,
	})
}

func (s *Server) handleSimulationRunEventsV2(w http.ResponseWriter, r *http.Request) {
	s.handleSimulationRunCollectionReadV2(w, r, "events")
}

func (s *Server) handleSimulationRunMetricsV2(w http.ResponseWriter, r *http.Request) {
	s.handleSimulationRunCollectionReadV2(w, r, "metrics")
}

func (s *Server) handleSimulationRunFindingsV2(w http.ResponseWriter, r *http.Request) {
	s.handleSimulationRunCollectionReadV2(w, r, "findings")
}

func (s *Server) handleSimulationRunArtifactsV2(w http.ResponseWriter, r *http.Request) {
	s.handleSimulationRunCollectionReadV2(w, r, "artifacts")
}

func (s *Server) handleSimulationRunCollectionReadV2(w http.ResponseWriter, r *http.Request, mode string) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	runID := strings.TrimSpace(r.PathValue("run_id"))

	switch mode {
	case "events":
		items, err := s.simulation.ListRunEvents(r.Context(), actor.TenantID, runID)
		if err != nil {
			s.writeSimulationError(w, corrID, "simulation_run_not_found", "Simulation run not found.", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "events": items, "correlation_id": corrID})
	case "metrics":
		items, err := s.simulation.ListRunMetrics(r.Context(), actor.TenantID, runID)
		if err != nil {
			s.writeSimulationError(w, corrID, "simulation_run_not_found", "Simulation run not found.", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "metrics": items, "correlation_id": corrID})
	case "findings":
		items, err := s.simulation.ListRunFindings(r.Context(), actor.TenantID, runID)
		if err != nil {
			s.writeSimulationError(w, corrID, "simulation_run_not_found", "Simulation run not found.", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "findings": items, "correlation_id": corrID})
	case "artifacts":
		items, err := s.simulation.ListRunArtifacts(r.Context(), actor.TenantID, runID)
		if err != nil {
			s.writeSimulationError(w, corrID, "simulation_run_not_found", "Simulation run not found.", err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"run_id": runID, "artifacts": items, "correlation_id": corrID})
	}
}

func (s *Server) handlePromoteSimulationBaselineV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for baseline promotion.", corrID, nil)
		return
	}

	var req promoteBaselineRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if strings.TrimSpace(req.Reason) == "" {
		req.Reason = "manual_promotion"
	}

	baseline, err := s.simulation.PromoteBaseline(r.Context(), simulationsvc.PromoteBaselineRequest{
		TenantID:   actor.TenantID,
		RunID:      strings.TrimSpace(r.PathValue("run_id")),
		Reason:     strings.TrimSpace(req.Reason),
		PromotedBy: actor.ID,
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_baseline_promote_failed", "Failed to promote simulation baseline.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"baseline_id":    baseline.BaselineID,
		"baseline":       baseline,
		"correlation_id": corrID,
	})
}

func (s *Server) handleListSimulationBaselinesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}
	items := s.simulation.ListBaselines(r.Context(), actor.TenantID, strings.TrimSpace(r.URL.Query().Get("scenario_id")))
	writeJSON(w, http.StatusOK, map[string]any{
		"baselines":      items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetSimulationBaselineV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	baseline, err := s.simulation.GetBaseline(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("baseline_id")))
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_baseline_not_found", "Simulation baseline not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"baseline":       baseline,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCompareSimulationRunsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	var req compareSimulationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if len(req.FailOnSeverities) == 0 {
		req.FailOnSeverities = []string{simulationsvc.SeverityHigh, simulationsvc.SeverityCritical}
	}

	result, err := s.simulation.Compare(r.Context(), simulationsvc.CompareRequest{
		TenantID:         actor.TenantID,
		CandidateRunID:   strings.TrimSpace(req.CandidateRunID),
		BaselineID:       strings.TrimSpace(req.BaselineID),
		FailOnSeverities: req.FailOnSeverities,
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_compare_failed", "Failed to compare simulation run with baseline.", err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"compare":        result,
		"correlation_id": corrID,
	})
}

func (s *Server) handleRequestSimulationReplayV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation replay request.", corrID, nil)
		return
	}

	var req requestReplayRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	replay, err := s.simulation.RequestReplay(r.Context(), simulationsvc.RequestReplayRequest{
		TenantID:        actor.TenantID,
		SourceType:      strings.TrimSpace(req.SourceType),
		SourceRef:       strings.TrimSpace(req.SourceRef),
		Sensitivity:     strings.TrimSpace(req.Sensitivity),
		RequestedBy:     actor.ID,
		ApprovalGranted: req.ApprovalGranted,
	})
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_replay_failed", "Failed to request simulation replay.", err)
		return
	}

	statusCode := http.StatusAccepted
	writeJSON(w, statusCode, map[string]any{
		"replay_id":      replay.ReplayID,
		"status":         replay.Status,
		"replay":         replay,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetSimulationReplayV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	replay, err := s.simulation.GetReplay(r.Context(), actor.TenantID, strings.TrimSpace(r.PathValue("replay_id")))
	if err != nil {
		s.writeSimulationError(w, corrID, "simulation_replay_not_found", "Simulation replay not found.", err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"replay":         replay,
		"correlation_id": corrID,
	})
}

func (s *Server) handleSimulationDashboardV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator", "approver", "auditor") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for simulation dashboard read.", corrID, nil)
		return
	}
	if err := s.simulation.EnsureV1ScenarioLibrary(r.Context(), actor.TenantID); err != nil {
		s.writeSimulationError(w, corrID, "simulation_seed_failed", "Failed to seed v1 simulation scenario library.", err)
		return
	}

	limit := atoiQueryDefault(r.URL.Query().Get("limit"), 20)
	dashboard := s.simulation.Dashboard(r.Context(), actor.TenantID, limit)
	writeJSON(w, http.StatusOK, map[string]any{
		"simulation_dashboard": dashboard,
		"correlation_id":       corrID,
	})
}

func (s *Server) writeAssuranceError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, assurancesvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, assurancesvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, assurancesvc.ErrCriticDiversity):
		apierrors.Write(w, http.StatusConflict, "critic_diversity_required", "High assurance validation requires critic diversity across provider and failure-mode classes.", corrID, nil)
	case errors.Is(err, assurancesvc.ErrTemplateNotReady):
		apierrors.Write(w, http.StatusConflict, "acceptance_template_not_published", "Acceptance template must be published before use.", corrID, nil)
	case errors.Is(err, assurancesvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}

func (s *Server) writeSimulationError(w http.ResponseWriter, corrID, code, message string, err error) {
	switch {
	case errors.Is(err, simulationsvc.ErrNotFound):
		apierrors.Write(w, http.StatusNotFound, code, message, corrID, nil)
	case errors.Is(err, simulationsvc.ErrAlreadyExists):
		apierrors.Write(w, http.StatusConflict, code, message, corrID, nil)
	case errors.Is(err, simulationsvc.ErrInvalidRequest):
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", err.Error(), corrID, nil)
	default:
		apierrors.Write(w, http.StatusInternalServerError, code, message, corrID, map[string]any{"reason": err.Error()})
	}
}

func atoiQueryDefault(raw string, fallback int) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return fallback
	}
	v, err := strconv.Atoi(raw)
	if err != nil || v <= 0 {
		return fallback
	}
	return v
}
