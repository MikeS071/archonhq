package httpserver

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
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

	mux.Handle("POST /v1/results", auth.RequireNode(http.HandlerFunc(s.handleSubmitResult)))

	// API contract placeholders.
	for _, route := range []string{
		"POST /v1/tasks/{task_id}/decompose",
		"POST /v1/approvals/{approval_id}/auto-mode",
		"POST /v1/artifacts/upload-url", "POST /v1/artifacts/register", "GET /v1/artifacts/{artifact_id}", "GET /v1/artifacts/{artifact_id}/download-url",
		"GET /v1/results/{result_id}", "GET /v1/tasks/{task_id}/results",
		"POST /v1/verifications", "GET /v1/verifications/{verification_id}", "GET /v1/results/{result_id}/verifications",
		"POST /v1/reductions", "GET /v1/reductions/{reduction_id}",
		"GET /v1/reliability/subjects/{subject_type}/{subject_id}", "GET /v1/operators/{operator_id}/reliability",
		"POST /v1/pricing/quote", "GET /v1/pricing/rate-cards", "POST /v1/pricing/bids", "GET /v1/tasks/{task_id}/market",
		"GET /v1/ledger/accounts/{account_id}", "GET /v1/ledger/accounts/{account_id}/entries", "GET /v1/operators/{operator_id}/earnings-summary", "GET /v1/operators/{operator_id}/reserve-holds",
		"GET /v1/policies", "POST /v1/policies", "PATCH /v1/policies/{policy_id}",
		"POST /v1/integrations/paperclip/sync", "GET /v1/integrations/paperclip/status",
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
	ResultID string   `json:"result_id"`
	TaskID   string   `json:"task_id"`
	LeaseID  string   `json:"lease_id"`
	Outputs  []string `json:"output_refs"`
}

func (s *Server) handleSubmitResult(w http.ResponseWriter, r *http.Request) {
	corrID := telemetry.CorrelationIDFromContext(r.Context())
	idemKey := r.Header.Get("Idempotency-Key")

	var req submitResultRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if req.ResultID == "" || req.TaskID == "" || req.LeaseID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "result_id, task_id, and lease_id are required.", corrID, nil)
		return
	}

	if s.events != nil {
		_ = s.events.Append(r.Context(), events.Envelope{
			EventID:        "evt_" + randomID(8),
			TenantID:       "tenant_stub",
			EntityType:     "result",
			EntityID:       req.ResultID,
			EventType:      "result.submitted",
			EventVersion:   1,
			ActorType:      "node",
			ActorID:        "node_stub",
			CorrelationID:  corrID,
			IdempotencyKey: idemKey,
			Payload: map[string]any{
				"task_id":      req.TaskID,
				"lease_id":     req.LeaseID,
				"output_count": len(req.Outputs),
			},
			OccurredAt: time.Now().UTC(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"result_id":      req.ResultID,
		"status":         "accepted_for_processing",
		"correlation_id": corrID,
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
