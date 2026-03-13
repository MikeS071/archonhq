package httpserver

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	"github.com/MikeS071/archonhq/pkg/auth"
	"github.com/MikeS071/archonhq/pkg/db"
	"github.com/MikeS071/archonhq/pkg/domain"
	"github.com/MikeS071/archonhq/pkg/events"
	natsclient "github.com/MikeS071/archonhq/pkg/nats"
	"github.com/MikeS071/archonhq/pkg/objectstore"
	redisclient "github.com/MikeS071/archonhq/pkg/redis"
	"github.com/MikeS071/archonhq/pkg/telemetry"
)

type Server struct {
	logger      *slog.Logger
	postgres    *db.Postgres
	nats        *natsclient.Client
	redis       *redisclient.Client
	objectStore *objectstore.Client
	events      events.Store
}

func New(
	logger *slog.Logger,
	postgres *db.Postgres,
	nats *natsclient.Client,
	redis *redisclient.Client,
	objectStore *objectstore.Client,
	eventStore events.Store,
) *Server {
	return &Server{
		logger:      logger,
		postgres:    postgres,
		nats:        nats,
		redis:       redis,
		objectStore: objectStore,
		events:      eventStore,
	}
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /health", s.handleHealth)
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ArchonHQ API"))
	})

	mux.Handle("POST /v1/tenants", auth.RequireHuman(http.HandlerFunc(s.handleCreateTenantV2)))
	mux.Handle("GET /v1/tenants/{tenant_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetTenantV2)))
	mux.Handle("PATCH /v1/tenants/{tenant_id}", auth.RequireHuman(http.HandlerFunc(s.handlePatchTenantV2)))
	mux.Handle("GET /v1/tenants/{tenant_id}/members", auth.RequireHuman(http.HandlerFunc(s.handleGetTenantMembersV2)))

	mux.Handle("POST /v1/workspaces", auth.RequireHuman(http.HandlerFunc(s.handleCreateWorkspaceV2)))
	mux.Handle("GET /v1/workspaces/{workspace_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetWorkspaceV2)))
	mux.Handle("GET /v1/workspaces/{workspace_id}/summary", auth.RequireHuman(http.HandlerFunc(s.handleWorkspaceSummaryV2)))
	mux.Handle("GET /v1/workspaces/{workspace_id}/tasks", auth.RequireHuman(http.HandlerFunc(s.handleWorkspaceTasksV2)))
	mux.Handle("GET /v1/workspaces/{workspace_id}/ledger", auth.RequireHuman(http.HandlerFunc(s.handleWorkspaceLedgerV2)))

	mux.Handle("POST /v1/nodes/register-intent", auth.RequireHuman(http.HandlerFunc(s.handleNodeRegisterIntentV2)))
	mux.Handle("POST /v1/nodes/register", auth.RequireHuman(http.HandlerFunc(s.handleNodeRegisterV2)))
	mux.Handle("GET /v1/nodes", auth.RequireHuman(http.HandlerFunc(s.handleListNodesV2)))
	mux.Handle("POST /v1/nodes/{node_id}/heartbeat", auth.RequireNode(http.HandlerFunc(s.handleNodeHeartbeatV2)))
	mux.Handle("GET /v1/nodes/{node_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetNodeV2)))
	mux.Handle("GET /v1/nodes/{node_id}/leases", auth.RequireHuman(http.HandlerFunc(s.handleGetNodeLeasesV2)))

	mux.Handle("POST /v1/tasks", auth.RequireHuman(http.HandlerFunc(s.handleCreateTaskV2)))
	mux.Handle("GET /v1/tasks/{task_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetTaskV2)))
	mux.Handle("GET /v1/tasks/feed", auth.RequireHuman(http.HandlerFunc(s.handleTaskFeedV2)))
	mux.Handle("POST /v1/tasks/{task_id}/cancel", auth.RequireHuman(http.HandlerFunc(s.handleCancelTaskV2)))

	mux.Handle("GET /v1/approvals/queue", auth.RequireHuman(http.HandlerFunc(s.handleApprovalQueueV2)))
	mux.Handle("GET /v1/approvals/{approval_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetApprovalV2)))
	mux.Handle("POST /v1/approvals/{approval_id}/approve", auth.RequireHuman(http.HandlerFunc(s.handleApproveV2)))
	mux.Handle("POST /v1/approvals/{approval_id}/deny", auth.RequireHuman(http.HandlerFunc(s.handleDenyV2)))

	mux.Handle("POST /v1/leases", auth.RequireHuman(http.HandlerFunc(s.handleCreateLeaseV2)))
	mux.Handle("POST /v1/leases/{lease_id}/claim", auth.RequireNode(http.HandlerFunc(s.handleClaimLeaseV2)))
	mux.Handle("POST /v1/leases/{lease_id}/release", auth.RequireNode(http.HandlerFunc(s.handleReleaseLeaseV2)))
	mux.Handle("POST /v1/leases/{lease_id}/extend", auth.RequireNode(http.HandlerFunc(s.handleExtendLeaseV2)))

	mux.Handle("POST /v1/artifacts/upload-url", auth.RequireNode(http.HandlerFunc(s.handleArtifactUploadURLV2)))
	mux.Handle("POST /v1/artifacts/register", auth.RequireNode(http.HandlerFunc(s.handleArtifactRegisterV2)))
	mux.Handle("GET /v1/artifacts/{artifact_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetArtifactV2)))
	mux.Handle("GET /v1/artifacts/{artifact_id}/download-url", auth.RequireHuman(http.HandlerFunc(s.handleArtifactDownloadURLV2)))

	mux.Handle("POST /v1/results", auth.RequireNode(http.HandlerFunc(s.handleSubmitResult)))
	mux.Handle("GET /v1/results/{result_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetResultV2)))
	mux.Handle("GET /v1/tasks/{task_id}/results", auth.RequireHuman(http.HandlerFunc(s.handleTaskResultsV2)))

	mux.Handle("GET /v1/reliability/subjects/{subject_type}/{subject_id}", auth.RequireHuman(http.HandlerFunc(s.handleReliabilitySubjectV2)))
	mux.Handle("GET /v1/operators/{operator_id}/reliability", auth.RequireHuman(http.HandlerFunc(s.handleOperatorReliabilityV2)))

	mux.Handle("POST /v1/pricing/quote", auth.RequireHuman(http.HandlerFunc(s.handlePricingQuoteV2)))
	mux.Handle("GET /v1/pricing/rate-cards", auth.RequireHuman(http.HandlerFunc(s.handlePricingRateCardsV2)))
	mux.Handle("POST /v1/pricing/bids", auth.RequireHuman(http.HandlerFunc(s.handlePricingBidsV2)))

	mux.Handle("GET /v1/ledger/accounts/{account_id}", auth.RequireHuman(http.HandlerFunc(s.handleGetLedgerAccountV2)))
	mux.Handle("GET /v1/ledger/accounts/{account_id}/entries", auth.RequireHuman(http.HandlerFunc(s.handleGetLedgerAccountEntriesV2)))
	mux.Handle("GET /v1/operators/{operator_id}/earnings-summary", auth.RequireHuman(http.HandlerFunc(s.handleOperatorEarningsSummaryV2)))
	mux.Handle("GET /v1/operators/{operator_id}/reserve-holds", auth.RequireHuman(http.HandlerFunc(s.handleOperatorReserveHoldsV2)))
	mux.Handle("POST /v1/ledger/settlements", auth.RequireHuman(http.HandlerFunc(s.handlePostSettlementV2)))
	mux.Handle("POST /v1/reserve-holds/{reserve_hold_id}/release", auth.RequireHuman(http.HandlerFunc(s.handleReleaseReserveHoldV2)))

	mux.Handle("GET /v1/policies", auth.RequireHuman(http.HandlerFunc(s.handleGetPoliciesV2)))
	mux.Handle("POST /v1/policies", auth.RequireHuman(http.HandlerFunc(s.handleCreatePolicyV2)))
	mux.Handle("PATCH /v1/policies/{policy_id}", auth.RequireHuman(http.HandlerFunc(s.handlePatchPolicyV2)))
	mux.Handle("POST /v1/integrations/paperclip/sync", auth.RequireHuman(http.HandlerFunc(s.handlePaperclipSyncV2)))
	mux.Handle("GET /v1/integrations/paperclip/status", auth.RequireHuman(http.HandlerFunc(s.handlePaperclipStatusV2)))

	// API contract placeholders.
	for _, route := range []string{
		"POST /v1/tasks/{task_id}/decompose",
		"POST /v1/approvals/{approval_id}/auto-mode",
		"POST /v1/verifications", "GET /v1/verifications/{verification_id}", "GET /v1/results/{result_id}/verifications",
		"POST /v1/reductions", "GET /v1/reductions/{reduction_id}",
		"GET /v1/tasks/{task_id}/market",
	} {
		mux.Handle(route, http.HandlerFunc(s.handleNotImplemented))
	}

	return s.withCorrelationID(s.withIdempotencyValidation(mux))
}

type createTaskRequest struct {
	WorkspaceID string `json:"workspace_id"`
	TaskFamily  string `json:"task_family"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	corrID := telemetry.CorrelationIDFromContext(r.Context())
	idemKey := r.Header.Get("Idempotency-Key")

	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	if strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.Title) == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "workspace_id and title are required.", corrID, nil)
		return
	}
	if _, ok := domain.SupportedTaskFamilies[req.TaskFamily]; !ok {
		apierrors.Write(w, http.StatusBadRequest, "invalid_task_family", "Unsupported task family.", corrID, map[string]any{"task_family": req.TaskFamily})
		return
	}

	taskID := "task_" + randomID(6)
	approvalID := "apr_" + randomID(6)

	if s.events != nil {
		_ = s.events.Append(r.Context(), events.Envelope{
			EventID:        "evt_" + randomID(8),
			TenantID:       "tenant_stub",
			WorkspaceID:    req.WorkspaceID,
			EntityType:     "task",
			EntityID:       taskID,
			EventType:      "task.created",
			EventVersion:   1,
			ActorType:      "human",
			ActorID:        "clerk_stub_user",
			CorrelationID:  corrID,
			IdempotencyKey: idemKey,
			Payload: map[string]any{
				"task_family": req.TaskFamily,
				"title":       req.Title,
			},
			OccurredAt: time.Now().UTC(),
		})
		_ = s.events.Append(r.Context(), events.Envelope{
			EventID:        "evt_" + randomID(8),
			TenantID:       "tenant_stub",
			WorkspaceID:    req.WorkspaceID,
			EntityType:     "approval",
			EntityID:       approvalID,
			EventType:      "approval.requested",
			EventVersion:   1,
			ActorType:      "human",
			ActorID:        "clerk_stub_user",
			CorrelationID:  corrID,
			IdempotencyKey: idemKey,
			Payload: map[string]any{
				"task_id": taskID,
			},
			OccurredAt: time.Now().UTC(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"task_id":              taskID,
		"status":               "awaiting_approval",
		"approval_request_id":  approvalID,
		"correlation_id":       corrID,
		"idempotency_key_echo": idemKey,
	})
}

type submitResultRequest struct {
	ResultID      string        `json:"result_id"`
	TaskID        string        `json:"task_id"`
	LeaseID       string        `json:"lease_id"`
	Outputs       []string      `json:"output_refs"`
	Signature     string        `json:"signature"`
	TelemetryRefs telemetryRefs `json:"telemetry_refs,omitempty"`
}

func (s *Server) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}
	idemKey := r.Header.Get("Idempotency-Key")

	var req submitResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.ResultID = strings.TrimSpace(req.ResultID)
	req.TaskID = strings.TrimSpace(req.TaskID)
	req.LeaseID = strings.TrimSpace(req.LeaseID)
	req.Signature = strings.TrimSpace(req.Signature)
	if req.ResultID == "" || req.TaskID == "" || req.LeaseID == "" || req.Signature == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "result_id, task_id, lease_id, and signature are required.", corrID, nil)
		return
	}

	const leaseQ = "SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1"
	var leaseTenantID, leaseNodeID, leaseTaskID, leaseStatus string
	if err := s.postgres.DB.QueryRowContext(r.Context(), leaseQ, req.LeaseID).Scan(&leaseTenantID, &leaseNodeID, &leaseTaskID, &leaseStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "lease_not_found", "Lease not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "lease_lookup_failed", "Failed to validate lease.", corrID, nil)
		return
	}
	if leaseTenantID != actor.TenantID || leaseNodeID != actor.ID || leaseTaskID != req.TaskID {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Lease does not belong to node/tenant/task.", corrID, nil)
		return
	}
	if leaseStatus != "claimed" && leaseStatus != "granted" {
		apierrors.Write(w, http.StatusConflict, "invalid_lease_status", "Lease is not in a submittable status.", corrID, nil)
		return
	}
	if !verifyResultSignature(req.Signature, actor.ID, req.LeaseID, req.ResultID) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_signature", "Invalid result signature.", corrID, nil)
		return
	}

	const workspaceQ = "SELECT workspace_id FROM tasks WHERE task_id = $1 AND tenant_id = $2"
	var workspaceID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), workspaceQ, req.TaskID, actor.TenantID).Scan(&workspaceID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to validate task.", corrID, nil)
		return
	}
	const artifactQ = "SELECT artifact_id FROM artifacts WHERE artifact_id = $1 AND tenant_id = $2 AND workspace_id = $3"
	for _, artifactID := range req.Outputs {
		artifactID = strings.TrimSpace(artifactID)
		if artifactID == "" {
			continue
		}
		var validatedID string
		if err := s.postgres.DB.QueryRowContext(r.Context(), artifactQ, artifactID, actor.TenantID, workspaceID).Scan(&validatedID); err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				apierrors.Write(w, http.StatusBadRequest, "invalid_output_ref", "Output artifact is missing or outside tenant/workspace scope.", corrID, map[string]any{"artifact_id": artifactID})
				return
			}
			apierrors.Write(w, http.StatusInternalServerError, "artifact_lookup_failed", "Failed to validate output artifacts.", corrID, nil)
			return
		}
	}

	meteringJSON, _ := json.Marshal(map[string]any{"output_count": len(req.Outputs)})
	qualityInputsJSON, _ := json.Marshal(map[string]any{"signature_verified": true})
	const insertResultQ = "INSERT INTO results (result_id, tenant_id, task_id, lease_id, node_id, status, signature, metering_json, quality_inputs_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertResultQ, req.ResultID, actor.TenantID, req.TaskID, req.LeaseID, actor.ID, "submitted", req.Signature, meteringJSON, qualityInputsJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "result_create_failed", "Failed to persist result.", corrID, nil)
		return
	}
	for _, artifactID := range req.Outputs {
		artifactID = strings.TrimSpace(artifactID)
		if artifactID == "" {
			continue
		}
		const insertOutputRefQ = "INSERT INTO result_output_refs (result_id, artifact_id) VALUES ($1,$2)"
		if _, err := s.postgres.DB.ExecContext(r.Context(), insertOutputRefQ, req.ResultID, artifactID); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "result_output_ref_failed", "Failed to persist result outputs.", corrID, nil)
			return
		}
	}
	if req.TelemetryRefs.LogsArtifactID != "" || req.TelemetryRefs.ToolCallsArtifactID != "" || req.TelemetryRefs.MetricsArtifactID != "" {
		const insertTelemetryQ = "INSERT INTO run_telemetry_refs (run_telemetry_id, tenant_id, lease_id, result_id, logs_artifact_id, tool_calls_artifact_id, metrics_artifact_id) VALUES ($1,$2,$3,$4,$5,$6,$7)"
		if _, err := s.postgres.DB.ExecContext(
			r.Context(),
			insertTelemetryQ,
			"rtel_"+randomID(6),
			actor.TenantID,
			req.LeaseID,
			req.ResultID,
			strings.TrimSpace(req.TelemetryRefs.LogsArtifactID),
			strings.TrimSpace(req.TelemetryRefs.ToolCallsArtifactID),
			strings.TrimSpace(req.TelemetryRefs.MetricsArtifactID),
		); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "telemetry_capture_failed", "Failed to persist telemetry references.", corrID, nil)
			return
		}
	}

	if s.events != nil {
		_ = s.events.Append(r.Context(), events.Envelope{
			EventID:        "evt_" + randomID(8),
			TenantID:       actor.TenantID,
			EntityType:     "result",
			EntityID:       req.ResultID,
			EventType:      "result.submitted",
			EventVersion:   1,
			ActorType:      "node",
			ActorID:        actor.ID,
			CorrelationID:  corrID,
			IdempotencyKey: idemKey,
			Payload: map[string]any{
				"task_id":      req.TaskID,
				"lease_id":     req.LeaseID,
				"output_count": len(req.Outputs),
				"signature_ok": true,
			},
			OccurredAt: time.Now().UTC(),
		})
		s.appendEvent(r, actor.TenantID, "result", req.ResultID, "result.signature_verified", map[string]any{"lease_id": req.LeaseID})
		if req.TelemetryRefs.LogsArtifactID != "" || req.TelemetryRefs.ToolCallsArtifactID != "" || req.TelemetryRefs.MetricsArtifactID != "" {
			s.appendEvent(r, actor.TenantID, "result", req.ResultID, "result.telemetry_captured", map[string]any{
				"logs_artifact_id":       req.TelemetryRefs.LogsArtifactID,
				"tool_calls_artifact_id": req.TelemetryRefs.ToolCallsArtifactID,
				"metrics_artifact_id":    req.TelemetryRefs.MetricsArtifactID,
			})
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"result_id":          req.ResultID,
		"status":             "accepted_for_processing",
		"signature_verified": true,
		"correlation_id":     corrID,
	})
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
	defer cancel()

	status := map[string]string{}
	httpStatus := http.StatusOK

	if err := s.postgres.Health(ctx); err != nil {
		httpStatus = http.StatusServiceUnavailable
		status["postgres"] = err.Error()
	} else {
		status["postgres"] = "ok"
	}
	if err := s.nats.Health(ctx); err != nil {
		httpStatus = http.StatusServiceUnavailable
		status["nats"] = err.Error()
	} else {
		status["nats"] = "ok"
	}
	if err := s.redis.Health(); err != nil {
		httpStatus = http.StatusServiceUnavailable
		status["redis"] = err.Error()
	} else {
		status["redis"] = "ok"
	}
	if err := s.objectStore.Health(); err != nil {
		httpStatus = http.StatusServiceUnavailable
		status["object_store"] = err.Error()
	} else {
		status["object_store"] = "ok"
	}

	writeJSON(w, httpStatus, map[string]any{
		"status": "ok",
		"checks": status,
	})
}

func (s *Server) handleNotImplemented(w http.ResponseWriter, r *http.Request) {
	corrID := telemetry.CorrelationIDFromContext(r.Context())
	apierrors.Write(w, http.StatusNotImplemented, "not_implemented", fmt.Sprintf("%s %s is not implemented yet.", r.Method, r.URL.Path), corrID, nil)
}

func (s *Server) withCorrelationID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		corrID := r.Header.Get("X-Correlation-ID")
		if corrID == "" {
			corrID = "corr_" + randomID(8)
		}
		w.Header().Set("X-Correlation-ID", corrID)
		next.ServeHTTP(w, r.WithContext(telemetry.WithCorrelationID(r.Context(), corrID)))
	})
}

func (s *Server) withIdempotencyValidation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !isMutatingMethod(r.Method) || !strings.HasPrefix(r.URL.Path, "/v1/") {
			next.ServeHTTP(w, r)
			return
		}
		if strings.TrimSpace(r.Header.Get("Idempotency-Key")) == "" {
			corrID := telemetry.CorrelationIDFromContext(r.Context())
			apierrors.Write(w, http.StatusBadRequest, "idempotency_key_required", "Mutating routes require Idempotency-Key header.", corrID, nil)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func isMutatingMethod(method string) bool {
	switch method {
	case http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func writeJSON(w http.ResponseWriter, statusCode int, body map[string]any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(body)
}

func randomID(nBytes int) string {
	buf := make([]byte, nBytes)
	_, _ = rand.Read(buf)
	return hex.EncodeToString(buf)
}
