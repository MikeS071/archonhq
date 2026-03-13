package httpserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/integrations/paperclip"
	"github.com/MikeS071/archonhq/pkg/apierrors"
	paperclipconnector "github.com/MikeS071/archonhq/services/paperclip-connector"
)

const (
	paperclipProjectionSourceOfTruth = "postgres"
	defaultPaperclipSyncLimit        = 50
	maxPaperclipSyncLimit            = 500
)

type paperclipSyncRequest struct {
	SyncID string `json:"sync_id,omitempty"`
	DryRun bool   `json:"dry_run,omitempty"`
	Force  bool   `json:"force,omitempty"`
	Limit  int    `json:"limit,omitempty"`
}

func (s *Server) handlePaperclipSyncV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("platform_admin", "tenant_admin", "operator", "approver", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for Paperclip sync.", corrID, nil)
		return
	}

	var req paperclipSyncRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil && !errors.Is(err, io.EOF) {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	limit := defaultPaperclipSyncLimit
	if req.Limit != 0 {
		if req.Limit < 0 || req.Limit > maxPaperclipSyncLimit {
			apierrors.Write(w, http.StatusBadRequest, "invalid_request", "limit must be an integer between 1 and 500.", corrID, nil)
			return
		}
		limit = req.Limit
	}

	syncID := strings.TrimSpace(req.SyncID)
	if syncID == "" {
		syncID = "pcsync_" + randomID(6)
	}

	payload, err := s.collectPaperclipProjectionPayload(r, actor.TenantID, limit)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "integration_sync_failed", "Failed to collect Paperclip projection payload.", corrID, nil)
		return
	}
	payload.Metadata = map[string]any{
		"sync_id": syncID,
		"force":   req.Force,
		"limit":   limit,
	}

	s.appendEvent(r, actor.TenantID, "integration", "paperclip", "integration.paperclip_sync_requested", map[string]any{
		"sync_id":         syncID,
		"dry_run":         req.DryRun,
		"force":           req.Force,
		"limit":           limit,
		"source_of_truth": paperclipProjectionSourceOfTruth,
	})

	service := paperclipconnector.New(paperclip.NoopConnector{})
	result, err := service.Sync(r.Context(), paperclipconnector.SyncRequest{
		SyncID:      syncID,
		InitiatedBy: actor.ID,
		DryRun:      req.DryRun,
		Force:       req.Force,
		Payload:     payload,
	})
	if err != nil {
		s.appendEvent(r, actor.TenantID, "integration", "paperclip", "integration.paperclip_sync_failed", map[string]any{
			"sync_id": syncID,
			"error":   err.Error(),
		})
		apierrors.Write(w, http.StatusBadGateway, "integration_sync_failed", "Failed to push projection snapshot to Paperclip.", corrID, map[string]any{
			"sync_id": syncID,
		})
		return
	}

	s.appendEvent(r, actor.TenantID, "integration", "paperclip", "integration.paperclip_sync_completed", map[string]any{
		"sync_id":         result.SyncID,
		"status":          result.Status,
		"external_ref":    result.ExternalRef,
		"source_of_truth": result.SourceOfTruth,
		"surface_counts":  result.SurfaceCounts,
	})

	writeJSON(w, http.StatusAccepted, map[string]any{
		"sync_id":                 result.SyncID,
		"status":                  result.Status,
		"external_ref":            result.ExternalRef,
		"source_of_truth":         result.SourceOfTruth,
		"paperclip_authoritative": false,
		"surface_counts":          result.SurfaceCounts,
		"generated_at":            result.GeneratedAt,
		"idempotency_key_echo":    r.Header.Get("Idempotency-Key"),
		"requested_by":            actor.ID,
		"correlation_id":          corrID,
	})
}

func (s *Server) handlePaperclipStatusV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("platform_admin", "tenant_admin", "operator", "approver", "finance_viewer", "auditor", "developer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for Paperclip status.", corrID, nil)
		return
	}

	const q = `
SELECT event_id, event_type, payload_json, occurred_at
FROM event_records
WHERE tenant_id = $1
  AND entity_type = 'integration'
  AND entity_id = 'paperclip'
  AND event_type IN (
    'integration.paperclip_sync_requested',
    'integration.paperclip_sync_completed',
    'integration.paperclip_sync_failed'
  )
ORDER BY occurred_at DESC
LIMIT 1
`
	var eventID, eventType string
	var payloadJSON []byte
	var occurredAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, actor.TenantID).Scan(&eventID, &eventType, &payloadJSON, &occurredAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{
				"integration":             "paperclip",
				"status":                  "never_synced",
				"source_of_truth":         paperclipProjectionSourceOfTruth,
				"paperclip_authoritative": false,
				"correlation_id":          corrID,
			})
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "integration_status_failed", "Failed to fetch Paperclip status.", corrID, nil)
		return
	}

	payload := map[string]any{}
	if len(payloadJSON) > 0 {
		_ = json.Unmarshal(payloadJSON, &payload)
	}
	status := paperclipStatusFromEvent(eventType, payload)

	writeJSON(w, http.StatusOK, map[string]any{
		"integration":             "paperclip",
		"status":                  status,
		"last_event_id":           eventID,
		"last_event_type":         eventType,
		"last_sync_id":            stringFromAny(payload["sync_id"]),
		"last_external_ref":       stringFromAny(payload["external_ref"]),
		"last_surface_counts":     payload["surface_counts"],
		"last_error":              stringFromAny(payload["error"]),
		"last_sync_at":            occurredAt,
		"source_of_truth":         paperclipProjectionSourceOfTruth,
		"paperclip_authoritative": false,
		"correlation_id":          corrID,
	})
}

func paperclipStatusFromEvent(eventType string, payload map[string]any) string {
	switch eventType {
	case "integration.paperclip_sync_completed":
		if status := stringFromAny(payload["status"]); status != "" {
			return status
		}
		return "completed"
	case "integration.paperclip_sync_failed":
		return "failed"
	case "integration.paperclip_sync_requested":
		return "requested"
	default:
		return "unknown"
	}
}

func stringFromAny(v any) string {
	switch t := v.(type) {
	case string:
		return strings.TrimSpace(t)
	default:
		return ""
	}
}

func (s *Server) collectPaperclipProjectionPayload(r *http.Request, tenantID string, limit int) (paperclip.ProjectionPayload, error) {
	workspaces, err := s.queryPaperclipWorkspaceSummaries(r, tenantID, limit)
	if err != nil {
		return paperclip.ProjectionPayload{}, err
	}
	approvals, err := s.queryPaperclipApprovals(r, tenantID, limit)
	if err != nil {
		return paperclip.ProjectionPayload{}, err
	}
	fleet, err := s.queryPaperclipFleet(r, tenantID, limit)
	if err != nil {
		return paperclip.ProjectionPayload{}, err
	}
	reliability, err := s.queryPaperclipReliability(r, tenantID, limit)
	if err != nil {
		return paperclip.ProjectionPayload{}, err
	}
	settlements, err := s.queryPaperclipSettlements(r, tenantID, limit)
	if err != nil {
		return paperclip.ProjectionPayload{}, err
	}

	return paperclip.ProjectionPayload{
		TenantID:           tenantID,
		GeneratedAt:        time.Now().UTC(),
		SourceOfTruth:      paperclipProjectionSourceOfTruth,
		WorkspaceSummaries: workspaces,
		Approvals:          approvals,
		Fleet:              fleet,
		Reliability:        reliability,
		Settlements:        settlements,
	}, nil
}

func (s *Server) queryPaperclipWorkspaceSummaries(r *http.Request, tenantID string, limit int) ([]map[string]any, error) {
	const q = `
SELECT
  w.workspace_id,
  w.name,
  COALESCE(t.tasks_total, 0) AS tasks_total,
  COALESCE(a.pending_approvals, 0) AS pending_approvals
FROM workspaces w
LEFT JOIN (
  SELECT workspace_id, COUNT(*) AS tasks_total
  FROM tasks
  WHERE tenant_id = $1
  GROUP BY workspace_id
) t ON t.workspace_id = w.workspace_id
LEFT JOIN (
  SELECT t.workspace_id, COUNT(*) AS pending_approvals
  FROM approval_requests a
  JOIN tasks t ON t.task_id = a.task_id
  WHERE a.tenant_id = $1
    AND a.status IN ('pending', 'awaiting_decision')
  GROUP BY t.workspace_id
) a ON a.workspace_id = w.workspace_id
WHERE w.tenant_id = $1
ORDER BY w.created_at DESC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var workspaceID, name string
		var tasksTotal, pendingApprovals int64
		if err := rows.Scan(&workspaceID, &name, &tasksTotal, &pendingApprovals); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"workspace_id":      workspaceID,
			"name":              name,
			"tasks_total":       tasksTotal,
			"pending_approvals": pendingApprovals,
		})
	}
	return out, nil
}

func (s *Server) queryPaperclipApprovals(r *http.Request, tenantID string, limit int) ([]map[string]any, error) {
	const q = `
SELECT
  aq.approval_id,
  aq.task_id,
  aq.status,
  aq.created_at,
  t.workspace_id,
  t.title,
  t.status AS task_status
FROM rm_approval_queue aq
LEFT JOIN tasks t ON t.task_id = aq.task_id AND t.tenant_id = aq.tenant_id
WHERE aq.tenant_id = $1
ORDER BY aq.created_at DESC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var approvalID, taskID, status string
		var createdAt time.Time
		var workspaceID, title, taskStatus sql.NullString
		if err := rows.Scan(&approvalID, &taskID, &status, &createdAt, &workspaceID, &title, &taskStatus); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"approval_id":  approvalID,
			"task_id":      taskID,
			"status":       status,
			"created_at":   createdAt,
			"workspace_id": workspaceID.String,
			"title":        title.String,
			"task_status":  taskStatus.String,
		})
	}
	return out, nil
}

func (s *Server) queryPaperclipFleet(r *http.Request, tenantID string, limit int) ([]map[string]any, error) {
	const q = `
SELECT node_id, operator_id, runtime_type, runtime_version, status, last_heartbeat_at
FROM rm_fleet_overview
WHERE tenant_id = $1
ORDER BY last_heartbeat_at DESC NULLS LAST, node_id ASC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var nodeID, operatorID, runtimeType, runtimeVersion, status string
		var lastHeartbeat sql.NullTime
		if err := rows.Scan(&nodeID, &operatorID, &runtimeType, &runtimeVersion, &status, &lastHeartbeat); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"node_id":           nodeID,
			"operator_id":       operatorID,
			"runtime_type":      runtimeType,
			"runtime_version":   runtimeVersion,
			"status":            status,
			"last_heartbeat_at": lastHeartbeat.Time,
		})
	}
	return out, nil
}

func (s *Server) queryPaperclipReliability(r *http.Request, tenantID string, limit int) ([]map[string]any, error) {
	const q = `
SELECT subject_type, subject_id, family, window_name, rf_value, created_at
FROM rm_reliability_summary
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var subjectType, subjectID, windowName string
		var family sql.NullString
		var rfValue float64
		var createdAt time.Time
		if err := rows.Scan(&subjectType, &subjectID, &family, &windowName, &rfValue, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"subject_type": subjectType,
			"subject_id":   subjectID,
			"family":       family.String,
			"window_name":  windowName,
			"rf_value":     rfValue,
			"created_at":   createdAt,
		})
	}
	return out, nil
}

func (s *Server) queryPaperclipSettlements(r *http.Request, tenantID string, limit int) ([]map[string]any, error) {
	const q = `
SELECT entry_id, account_id, result_id, credited_jw::text, net_amount::text, status, created_at
FROM rm_recent_settlements
WHERE tenant_id = $1
ORDER BY created_at DESC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]map[string]any, 0)
	for rows.Next() {
		var entryID, accountID, creditedJW, netAmount, status string
		var resultID sql.NullString
		var createdAt time.Time
		if err := rows.Scan(&entryID, &accountID, &resultID, &creditedJW, &netAmount, &status, &createdAt); err != nil {
			return nil, err
		}
		out = append(out, map[string]any{
			"entry_id":    entryID,
			"account_id":  accountID,
			"result_id":   resultID.String,
			"credited_jw": creditedJW,
			"net_amount":  netAmount,
			"status":      status,
			"created_at":  createdAt,
		})
	}
	return out, nil
}
