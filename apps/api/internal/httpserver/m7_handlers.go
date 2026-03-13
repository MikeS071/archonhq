package httpserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	reductionsvc "github.com/MikeS071/archonhq/services/reduction"
	schedulersvc "github.com/MikeS071/archonhq/services/scheduler"
	verificationsvc "github.com/MikeS071/archonhq/services/verification"
)

const (
	defaultDecomposeMaxChildren = 4
	maxDecomposeMaxChildren     = 20
	defaultAutoMaxIterations    = 3
	defaultAutoBudgetJW         = 10.0
)

type taskDecomposeRequest struct {
	MergeStrategy string                  `json:"merge_strategy,omitempty"`
	MaxChildren   int                     `json:"max_children,omitempty"`
	Autosearch    *boundedAutosearchInput `json:"autosearch,omitempty"`
}

type boundedAutosearchInput struct {
	Enabled                  bool      `json:"enabled,omitempty"`
	ExperimentID             string    `json:"experiment_id,omitempty"`
	MaxIterations            int       `json:"max_iterations,omitempty"`
	BudgetLimitJW            float64   `json:"budget_limit_jw,omitempty"`
	RequireApproval          bool      `json:"require_approval,omitempty"`
	ApprovalGranted          bool      `json:"approval_granted,omitempty"`
	MinAcceptScore           float64   `json:"min_accept_score,omitempty"`
	CandidateBenchmarkDeltas []float64 `json:"candidate_benchmark_deltas,omitempty"`
}

type approvalAutoModeRequest struct {
	Enabled         *bool   `json:"enabled,omitempty"`
	MaxIterations   int     `json:"max_iterations,omitempty"`
	BudgetLimitJW   float64 `json:"budget_limit_jw,omitempty"`
	RequireApproval bool    `json:"require_approval,omitempty"`
	ApprovalGranted bool    `json:"approval_granted,omitempty"`
}

type createVerificationRequest struct {
	VerificationID string                            `json:"verification_id,omitempty"`
	ResultID       string                            `json:"result_id"`
	VerifierType   string                            `json:"verifier_type"`
	VerifierID     string                            `json:"verifier_id,omitempty"`
	Score          *float64                          `json:"score,omitempty"`
	Report         map[string]any                    `json:"report,omitempty"`
	Iterative      *verificationsvc.IterativeContext `json:"iterative,omitempty"`
}

type reductionCandidateRequest struct {
	ResultID string         `json:"result_id"`
	Score    float64        `json:"score"`
	PatchOps []string       `json:"patch_ops,omitempty"`
	StateRef string         `json:"state_ref,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type createReductionRequest struct {
	ReductionID string                      `json:"reduction_id,omitempty"`
	TaskID      string                      `json:"task_id"`
	Strategy    string                      `json:"strategy,omitempty"`
	Candidates  []reductionCandidateRequest `json:"candidates"`
	Autosearch  *boundedAutosearchInput     `json:"autosearch,omitempty"`
}

type taskCoreRecord struct {
	TaskID        string
	TenantID      string
	WorkspaceID   string
	TaskFamily    string
	Title         string
	Status        string
	MergeStrategy sql.NullString
}

func (s *Server) handleTaskDecomposeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "developer", "approver", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for task decomposition.", corrID, nil)
		return
	}

	var req taskDecomposeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	task, err := s.lookupTaskCore(r, strings.TrimSpace(r.PathValue("task_id")))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to fetch task for decomposition.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, task.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	maxChildren := req.MaxChildren
	if maxChildren == 0 {
		maxChildren = defaultDecomposeMaxChildren
	}
	if maxChildren < 1 || maxChildren > maxDecomposeMaxChildren {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "max_children must be between 1 and 20.", corrID, nil)
		return
	}

	requestedStrategy := strings.TrimSpace(req.MergeStrategy)
	if requestedStrategy == "" {
		requestedStrategy = strings.TrimSpace(task.MergeStrategy.String)
	}
	mergeStrategy := reductionsvc.ResolveStrategy(task.TaskFamily, requestedStrategy)
	plan := buildDecompositionPlan(task, maxChildren)

	var autosearchPreview any = nil
	if task.TaskFamily == "autosearch.self_improve" {
		loopReq := buildAutosearchLoopRequest(req.Autosearch)
		loopReq.ExperimentID = nonEmpty(loopReq.ExperimentID, "exp_"+task.TaskID)

		loopResult, err := schedulersvc.New(nil, nil).Run(r.Context(), loopReq)
		if err != nil {
			switch {
			case errors.Is(err, schedulersvc.ErrApprovalGateRequired):
				apierrors.Write(w, http.StatusConflict, "approval_required", "Autosearch loop requires approval gate.", corrID, nil)
			case errors.Is(err, schedulersvc.ErrInvalidLoopPolicy):
				apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid autosearch guardrail policy.", corrID, nil)
			default:
				apierrors.Write(w, http.StatusInternalServerError, "task_decompose_failed", "Failed to execute bounded autosearch preview.", corrID, nil)
			}
			return
		}
		autosearchPreview = loopResult
	}

	simulationEntrypoints := buildSimulationEntrypoints(task.TaskFamily, mergeStrategy)
	s.appendEvent(r, actor.TenantID, "task", task.TaskID, "task.decomposed", map[string]any{
		"task_family":    task.TaskFamily,
		"merge_strategy": mergeStrategy,
		"max_children":   maxChildren,
		"has_autosearch": autosearchPreview != nil,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":                task.TaskID,
		"task_family":            task.TaskFamily,
		"merge_strategy":         mergeStrategy,
		"children":               plan,
		"autosearch_preview":     autosearchPreview,
		"simulation_entrypoints": simulationEntrypoints,
		"lineage": map[string]any{
			"task_id":      task.TaskID,
			"workspace_id": task.WorkspaceID,
			"task_status":  task.Status,
			"generated_at": time.Now().UTC(),
		},
		"idempotency_key_echo": r.Header.Get("Idempotency-Key"),
		"correlation_id":       corrID,
	})
}

func (s *Server) handleApprovalAutoModeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for approval auto-mode.", corrID, nil)
		return
	}

	var req approvalAutoModeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	approvalID := strings.TrimSpace(r.PathValue("approval_id"))
	const lookupQ = "SELECT task_id, status, payload_json FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2"
	var taskID, status string
	var payloadJSON []byte
	if err := s.postgres.DB.QueryRowContext(r.Context(), lookupQ, approvalID, actor.TenantID).Scan(&taskID, &status, &payloadJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "approval_not_found", "Approval not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "approval_lookup_failed", "Failed to fetch approval.", corrID, nil)
		return
	}
	if status != "pending" && status != "approved" {
		apierrors.Write(w, http.StatusConflict, "approval_not_pending", "Approval is not pending/approved for auto-mode update.", corrID, nil)
		return
	}

	enabled := true
	if req.Enabled != nil {
		enabled = *req.Enabled
	}
	maxIterations := req.MaxIterations
	if maxIterations == 0 {
		maxIterations = defaultAutoMaxIterations
	}
	if maxIterations < 1 || maxIterations > 25 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "max_iterations must be between 1 and 25.", corrID, nil)
		return
	}
	budgetLimitJW := req.BudgetLimitJW
	if budgetLimitJW <= 0 {
		budgetLimitJW = defaultAutoBudgetJW
	}

	payload := map[string]any{}
	if len(payloadJSON) > 0 {
		_ = json.Unmarshal(payloadJSON, &payload)
	}
	if payload == nil {
		payload = map[string]any{}
	}

	autoMode := map[string]any{
		"enabled":          enabled,
		"max_iterations":   maxIterations,
		"budget_limit_jw":  budgetLimitJW,
		"require_approval": req.RequireApproval,
		"approval_granted": req.ApprovalGranted,
		"updated_by":       actor.ID,
		"updated_at":       time.Now().UTC(),
	}
	payload["auto_mode"] = autoMode
	updatedPayload, _ := json.Marshal(payload)

	const updateQ = "UPDATE approval_requests SET payload_json = $1 WHERE approval_id = $2 AND tenant_id = $3"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateQ, updatedPayload, approvalID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "approval_update_failed", "Failed to update approval auto-mode.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "approval", approvalID, "approval.auto_mode_updated", map[string]any{
		"task_id":         taskID,
		"enabled":         enabled,
		"max_iterations":  maxIterations,
		"budget_limit_jw": budgetLimitJW,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"approval_id":          approvalID,
		"task_id":              taskID,
		"status":               status,
		"auto_mode":            autoMode,
		"correlation_id":       corrID,
		"idempotency_key_echo": r.Header.Get("Idempotency-Key"),
	})
}

func (s *Server) handleCreateVerificationV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "developer", "approver", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for verification creation.", corrID, nil)
		return
	}

	var req createVerificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.ResultID = strings.TrimSpace(req.ResultID)
	req.VerifierType = strings.TrimSpace(req.VerifierType)
	if req.ResultID == "" || req.VerifierType == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "result_id and verifier_type are required.", corrID, nil)
		return
	}

	const resultQ = "SELECT result_id, tenant_id, task_id FROM results WHERE result_id = $1"
	var resultID, tenantID, taskID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), resultQ, req.ResultID).Scan(&resultID, &tenantID, &taskID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "result_not_found", "Result not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "verification_failed", "Failed to load result for verification.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	var taskFamily string
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2", taskID, tenantID).Scan(&taskFamily)

	verificationResult, err := verificationsvc.New(nil, nil).Evaluate(r.Context(), verificationsvc.Request{
		VerifierType: req.VerifierType,
		Report:       req.Report,
		InputScore:   req.Score,
		Iterative:    req.Iterative,
	})
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "verification_failed", "Verification evaluation failed.", corrID, map[string]any{"reason": err.Error()})
		return
	}

	verificationID := strings.TrimSpace(req.VerificationID)
	if verificationID == "" {
		verificationID = "ver_" + randomID(6)
	}
	if req.Report == nil {
		req.Report = map[string]any{}
	}
	req.Report["hook_outputs"] = verificationResult.HookOutputs
	req.Report["lineage"] = map[string]any{
		"task_id":       taskID,
		"result_id":     resultID,
		"verifier_type": req.VerifierType,
	}
	req.Report["simulation_entrypoints"] = buildSimulationEntrypoints(taskFamily, reductionsvc.ResolveStrategy(taskFamily, ""))
	reportJSON, _ := json.Marshal(req.Report)

	const insertQ = "INSERT INTO verifications (verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, verificationID, tenantID, resultID, req.VerifierType, strings.TrimSpace(req.VerifierID), verificationResult.Score, verificationResult.Decision, reportJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "verification_failed", "Failed to persist verification.", corrID, nil)
		return
	}

	s.appendEvent(r, tenantID, "verification", verificationID, "verification.completed", map[string]any{
		"result_id":     resultID,
		"decision":      verificationResult.Decision,
		"score":         verificationResult.Score,
		"verifier_type": req.VerifierType,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"verification_id": verificationID,
		"result_id":       resultID,
		"verifier_type":   req.VerifierType,
		"score":           verificationResult.Score,
		"decision":        verificationResult.Decision,
		"report":          req.Report,
		"correlation_id":  corrID,
	})
}

func (s *Server) handleGetVerificationV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE verification_id = $1"
	var verificationID, tenantID, resultID, verifierType, verifierID, decision string
	var score sql.NullFloat64
	var reportJSON []byte
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, strings.TrimSpace(r.PathValue("verification_id"))).Scan(&verificationID, &tenantID, &resultID, &verifierType, &verifierID, &score, &decision, &reportJSON, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "verification_not_found", "Verification not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "verification_failed", "Failed to fetch verification.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	report := map[string]any{}
	if len(reportJSON) > 0 {
		_ = json.Unmarshal(reportJSON, &report)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"verification_id": verificationID,
		"tenant_id":       tenantID,
		"result_id":       resultID,
		"verifier_type":   verifierType,
		"verifier_id":     verifierID,
		"score":           score.Float64,
		"decision":        decision,
		"report":          report,
		"created_at":      createdAt,
		"correlation_id":  corrID,
	})
}

func (s *Server) handleResultVerificationsV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	const q = "SELECT verification_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE result_id = $1 AND tenant_id = $2 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, strings.TrimSpace(r.PathValue("result_id")), actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "verification_failed", "Failed to list verifications.", corrID, nil)
		return
	}
	defer rows.Close()

	items := make([]map[string]any, 0)
	for rows.Next() {
		var verificationID, verifierType, verifierID, decision string
		var score sql.NullFloat64
		var reportJSON []byte
		var createdAt time.Time
		if err := rows.Scan(&verificationID, &verifierType, &verifierID, &score, &decision, &reportJSON, &createdAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "verification_failed", "Failed to scan verifications.", corrID, nil)
			return
		}
		report := map[string]any{}
		if len(reportJSON) > 0 {
			_ = json.Unmarshal(reportJSON, &report)
		}
		items = append(items, map[string]any{
			"verification_id": verificationID,
			"verifier_type":   verifierType,
			"verifier_id":     verifierID,
			"score":           score.Float64,
			"decision":        decision,
			"report":          report,
			"created_at":      createdAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"result_id":      strings.TrimSpace(r.PathValue("result_id")),
		"verifications":  items,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreateReductionV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "developer", "approver", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for reduction creation.", corrID, nil)
		return
	}

	var req createReductionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.TaskID = strings.TrimSpace(req.TaskID)
	if req.TaskID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "task_id is required.", corrID, nil)
		return
	}

	task, err := s.lookupTaskCore(r, req.TaskID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to fetch task for reduction.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, task.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	candidates := make([]reductionsvc.Candidate, 0, len(req.Candidates))
	for _, c := range req.Candidates {
		candidates = append(candidates, reductionsvc.Candidate{
			ResultID: strings.TrimSpace(c.ResultID),
			Score:    c.Score,
			PatchOps: c.PatchOps,
			StateRef: strings.TrimSpace(c.StateRef),
			Metadata: c.Metadata,
		})
	}
	if len(candidates) == 0 {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "candidates must contain at least one item.", corrID, nil)
		return
	}

	reductionService := reductionsvc.New()
	decision, err := reductionService.Merge(reductionsvc.MergeRequest{
		TaskFamily: task.TaskFamily,
		Strategy:   req.Strategy,
		Candidates: candidates,
	})
	if err != nil {
		switch {
		case errors.Is(err, reductionsvc.ErrUnsupportedMergeStrategy):
			apierrors.Write(w, http.StatusBadRequest, "merge_strategy_unsupported", "Unsupported merge strategy.", corrID, nil)
		case errors.Is(err, reductionsvc.ErrNoCandidates):
			apierrors.Write(w, http.StatusBadRequest, "invalid_request", "At least one reduction candidate is required.", corrID, nil)
		default:
			apierrors.Write(w, http.StatusInternalServerError, "reduction_failed", "Reduction merge failed.", corrID, map[string]any{"reason": err.Error()})
		}
		return
	}

	var autosearchPreview any = nil
	if task.TaskFamily == "autosearch.self_improve" || (req.Autosearch != nil && req.Autosearch.Enabled) {
		loopReq := buildAutosearchLoopRequest(req.Autosearch)
		if loopReq.ExperimentID == "" {
			loopReq.ExperimentID = "exp_" + task.TaskID
		}
		if len(loopReq.Candidates) == 0 {
			loopReq.Candidates = make([]schedulersvc.IterationCandidate, 0, len(candidates))
			for _, c := range candidates {
				loopReq.Candidates = append(loopReq.Candidates, schedulersvc.IterationCandidate{
					CandidateID:     c.ResultID,
					BenchmarkDelta:  c.Score - 0.5,
					EstimatedCostJW: 1,
				})
			}
		}
		loopResult, err := schedulersvc.New(nil, nil).Run(r.Context(), loopReq)
		if err != nil {
			switch {
			case errors.Is(err, schedulersvc.ErrApprovalGateRequired):
				apierrors.Write(w, http.StatusConflict, "approval_required", "Autosearch loop requires approval gate.", corrID, nil)
			case errors.Is(err, schedulersvc.ErrInvalidLoopPolicy):
				apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid autosearch guardrail policy.", corrID, nil)
			default:
				apierrors.Write(w, http.StatusInternalServerError, "reduction_failed", "Failed to execute bounded autosearch loop.", corrID, nil)
			}
			return
		}
		autosearchPreview = loopResult
	}

	reductionID := strings.TrimSpace(req.ReductionID)
	if reductionID == "" {
		reductionID = "red_" + randomID(6)
	}

	simulationEntrypoints := buildSimulationEntrypoints(task.TaskFamily, decision.Strategy)
	decisionJSON, _ := json.Marshal(map[string]any{
		"status":                 decision.Status,
		"winner_result_id":       decision.WinnerResultID,
		"ranked_result_ids":      decision.RankedResultIDs,
		"merged_patch_ops":       decision.MergedPatchOps,
		"explanation":            decision.Explanation,
		"autosearch_preview":     autosearchPreview,
		"simulation_entrypoints": simulationEntrypoints,
		"lineage": map[string]any{
			"task_id":     task.TaskID,
			"task_family": task.TaskFamily,
			"created_by":  actor.ID,
		},
	})

	const insertQ = "INSERT INTO reductions (reduction_id, tenant_id, task_id, strategy, output_state_ref, decision_json) VALUES ($1,$2,$3,$4,$5,$6)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, reductionID, task.TenantID, task.TaskID, decision.Strategy, decision.OutputStateRef, decisionJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "reduction_failed", "Failed to persist reduction.", corrID, nil)
		return
	}

	eventType := "reduction.created"
	if decision.Status == "accepted" {
		eventType = "reduction.accepted"
	}
	s.appendEvent(r, task.TenantID, "reduction", reductionID, eventType, map[string]any{
		"task_id":          task.TaskID,
		"strategy":         decision.Strategy,
		"winner_result_id": decision.WinnerResultID,
		"status":           decision.Status,
	})

	writeJSON(w, http.StatusOK, map[string]any{
		"reduction_id":           reductionID,
		"task_id":                task.TaskID,
		"strategy":               decision.Strategy,
		"status":                 decision.Status,
		"winner_result_id":       decision.WinnerResultID,
		"ranked_result_ids":      decision.RankedResultIDs,
		"merged_patch_ops":       decision.MergedPatchOps,
		"output_state_ref":       decision.OutputStateRef,
		"autosearch_preview":     autosearchPreview,
		"simulation_entrypoints": simulationEntrypoints,
		"correlation_id":         corrID,
	})
}

func (s *Server) handleGetReductionV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	const q = "SELECT reduction_id, tenant_id, task_id, strategy, output_state_ref, decision_json, created_at FROM reductions WHERE reduction_id = $1"
	var reductionID, tenantID, taskID, strategy, outputStateRef string
	var decisionJSON []byte
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, strings.TrimSpace(r.PathValue("reduction_id"))).Scan(&reductionID, &tenantID, &taskID, &strategy, &outputStateRef, &decisionJSON, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "reduction_not_found", "Reduction not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "reduction_failed", "Failed to fetch reduction.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	decision := map[string]any{}
	if len(decisionJSON) > 0 {
		_ = json.Unmarshal(decisionJSON, &decision)
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"reduction_id":     reductionID,
		"tenant_id":        tenantID,
		"task_id":          taskID,
		"strategy":         strategy,
		"output_state_ref": outputStateRef,
		"decision":         decision,
		"created_at":       createdAt,
		"correlation_id":   corrID,
	})
}

func (s *Server) handleTaskMarketV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	task, err := s.lookupTaskCore(r, strings.TrimSpace(r.PathValue("task_id")))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_market_failed", "Failed to fetch task market context.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, task.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	var quote map[string]any
	{
		const quoteQ = "SELECT quote_id, strategy_name, quote_json, created_at FROM price_quotes WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1"
		var quoteID, strategyName string
		var quoteJSON []byte
		var createdAt time.Time
		if err := s.postgres.DB.QueryRowContext(r.Context(), quoteQ, task.TaskID, task.TenantID).Scan(&quoteID, &strategyName, &quoteJSON, &createdAt); err == nil {
			payload := map[string]any{}
			_ = json.Unmarshal(quoteJSON, &payload)
			quote = map[string]any{
				"quote_id":      quoteID,
				"strategy_name": strategyName,
				"quote":         payload,
				"created_at":    createdAt,
			}
		}
	}

	var rateSnapshot map[string]any
	{
		const rateQ = "SELECT rate_snapshot_id, strategy_name, rate_value, metadata_json, created_at FROM rate_snapshots WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1"
		var rateSnapshotID, strategyName string
		var rateValue float64
		var metadataJSON []byte
		var createdAt time.Time
		if err := s.postgres.DB.QueryRowContext(r.Context(), rateQ, task.TaskID, task.TenantID).Scan(&rateSnapshotID, &strategyName, &rateValue, &metadataJSON, &createdAt); err == nil {
			metadata := map[string]any{}
			_ = json.Unmarshal(metadataJSON, &metadata)
			rateSnapshot = map[string]any{
				"rate_snapshot_id": rateSnapshotID,
				"strategy_name":    strategyName,
				"rate_value":       rateValue,
				"metadata":         metadata,
				"created_at":       createdAt,
			}
		}
	}

	var totalResults, acceptedVerifications int64
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM results WHERE task_id = $1 AND tenant_id = $2", task.TaskID, task.TenantID).Scan(&totalResults)
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM verifications v WHERE v.tenant_id = $1 AND v.result_id IN (SELECT result_id FROM results WHERE task_id = $2 AND tenant_id = $1) AND v.decision = 'accepted'", task.TenantID, task.TaskID).Scan(&acceptedVerifications)

	acceptanceRate := 0.0
	if totalResults > 0 {
		acceptanceRate = float64(acceptedVerifications) / float64(totalResults)
	}

	var latestReductionID string
	var latestReductionAt time.Time
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT reduction_id, created_at FROM reductions WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1", task.TaskID, task.TenantID).Scan(&latestReductionID, &latestReductionAt)

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":                    task.TaskID,
		"task_family":                task.TaskFamily,
		"status":                     task.Status,
		"recommended_merge_strategy": reductionsvc.ResolveStrategy(task.TaskFamily, task.MergeStrategy.String),
		"quote":                      quote,
		"rate_snapshot":              rateSnapshot,
		"signals": map[string]any{
			"total_results":                totalResults,
			"accepted_verifications":       acceptedVerifications,
			"verification_acceptance_rate": acceptanceRate,
		},
		"lineage_view": map[string]any{
			"latest_reduction_id": latestReductionID,
			"latest_reduction_at": latestReductionAt,
		},
		"simulation_entrypoints": buildSimulationEntrypoints(task.TaskFamily, reductionsvc.ResolveStrategy(task.TaskFamily, task.MergeStrategy.String)),
		"correlation_id":         corrID,
	})
}

func (s *Server) lookupTaskCore(r *http.Request, taskID string) (taskCoreRecord, error) {
	const q = "SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1"
	var rec taskCoreRecord
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, taskID).Scan(&rec.TaskID, &rec.TenantID, &rec.WorkspaceID, &rec.TaskFamily, &rec.Title, &rec.Status, &rec.MergeStrategy); err != nil {
		return taskCoreRecord{}, err
	}
	return rec, nil
}

func buildDecompositionPlan(task taskCoreRecord, maxChildren int) []map[string]any {
	out := make([]map[string]any, 0, maxChildren)
	for i := 0; i < maxChildren; i++ {
		out = append(out, map[string]any{
			"child_id":    "sub_" + task.TaskID + "_" + strings.TrimSpace(strings.ToLower(time.Now().UTC().Format("150405"))) + "_" + strings.TrimSpace(strings.ToLower(string(rune('a'+i)))),
			"title":       task.Title + " :: shard " + strings.TrimSpace(string(rune('A'+i))),
			"task_family": task.TaskFamily,
			"status":      "planned",
			"priority":    i + 1,
		})
	}
	return out
}

func buildAutosearchLoopRequest(in *boundedAutosearchInput) schedulersvc.LoopRequest {
	req := schedulersvc.LoopRequest{
		Candidates: []schedulersvc.IterationCandidate{
			{CandidateID: "cand_1", BenchmarkDelta: 0.12, EstimatedCostJW: 1},
			{CandidateID: "cand_2", BenchmarkDelta: 0.07, EstimatedCostJW: 1},
			{CandidateID: "cand_3", BenchmarkDelta: 0.03, EstimatedCostJW: 1},
		},
		Policy: schedulersvc.LoopPolicy{
			MaxIterations:   defaultAutoMaxIterations,
			BudgetLimitJW:   defaultAutoBudgetJW,
			RequireApproval: false,
			ApprovalGranted: true,
			MinAcceptScore:  0.62,
		},
	}
	if in == nil {
		return req
	}
	req.ExperimentID = strings.TrimSpace(in.ExperimentID)
	if in.MaxIterations > 0 {
		req.Policy.MaxIterations = in.MaxIterations
	}
	if in.BudgetLimitJW > 0 {
		req.Policy.BudgetLimitJW = in.BudgetLimitJW
	}
	req.Policy.RequireApproval = in.RequireApproval
	req.Policy.ApprovalGranted = in.ApprovalGranted
	if in.MinAcceptScore > 0 {
		req.Policy.MinAcceptScore = in.MinAcceptScore
	}
	if len(in.CandidateBenchmarkDeltas) > 0 {
		req.Candidates = make([]schedulersvc.IterationCandidate, 0, len(in.CandidateBenchmarkDeltas))
		for i, delta := range in.CandidateBenchmarkDeltas {
			req.Candidates = append(req.Candidates, schedulersvc.IterationCandidate{
				CandidateID:     "cand_" + strings.TrimSpace(strings.ToLower(time.Now().UTC().Format("150405"))) + "_" + strings.TrimSpace(string(rune('a'+i))),
				BenchmarkDelta:  delta,
				EstimatedCostJW: 1,
			})
		}
	}
	return req
}

func buildSimulationEntrypoints(taskFamily, mergeStrategy string) []map[string]any {
	out := []map[string]any{
		{
			"scenario_id": "approval_backlog_v1",
			"run_mode":    "deterministic_stub",
		},
	}
	switch strings.TrimSpace(strings.ToLower(taskFamily)) {
	case "code.patch":
		out = append(out, map[string]any{
			"scenario_id":    "code_patch_merge_storm_v1",
			"run_mode":       "deterministic_stub",
			"merge_strategy": mergeStrategy,
		})
	case "autosearch.self_improve":
		out = append(out, map[string]any{
			"scenario_id": "autosearch_reward_hacking_v1",
			"run_mode":    "deterministic_stub",
		})
	case "reduce.merge":
		out = append(out, map[string]any{
			"scenario_id": "reducer_instability_v1",
			"run_mode":    "deterministic_stub",
		})
	}
	return out
}

func nonEmpty(v, fallback string) string {
	if strings.TrimSpace(v) != "" {
		return strings.TrimSpace(v)
	}
	return fallback
}
