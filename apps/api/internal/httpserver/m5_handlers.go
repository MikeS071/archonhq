package httpserver

import (
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"

	"github.com/MikeS071/archonhq/pkg/apierrors"
)

type listNodesResponseItem struct {
	NodeID          string        `json:"node_id"`
	OperatorID      string        `json:"operator_id"`
	RuntimeType     string        `json:"runtime_type"`
	RuntimeVersion  string        `json:"runtime_version"`
	Status          string        `json:"status"`
	LastHeartbeatAt sql.NullTime  `json:"last_heartbeat_at"`
	ActiveLeases    sql.NullInt64 `json:"active_leases"`
}

func (s *Server) handleListNodesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}

	limit := 200
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		parsed, err := strconv.Atoi(rawLimit)
		if err != nil || parsed <= 0 || parsed > 500 {
			apierrors.Write(w, http.StatusBadRequest, "invalid_request", "limit must be an integer between 1 and 500.", corrID, nil)
			return
		}
		limit = parsed
	}

	const q = `
SELECT
  n.node_id,
  n.operator_id,
  n.runtime_type,
  n.runtime_version,
  n.status,
  n.last_heartbeat_at,
  COALESCE(l.active_leases, 0) AS active_leases
FROM nodes n
LEFT JOIN (
  SELECT node_id, COUNT(*) AS active_leases
  FROM leases
  WHERE tenant_id = $1 AND status IN ('granted', 'claimed')
  GROUP BY node_id
) l ON l.node_id = n.node_id
WHERE n.tenant_id = $1
ORDER BY n.last_heartbeat_at DESC NULLS LAST, n.node_id ASC
LIMIT $2
`
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, limit)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "nodes_lookup_failed", "Failed to fetch node list.", corrID, nil)
		return
	}
	defer rows.Close()

	nodes := make([]map[string]any, 0)
	for rows.Next() {
		var item listNodesResponseItem
		if err := rows.Scan(&item.NodeID, &item.OperatorID, &item.RuntimeType, &item.RuntimeVersion, &item.Status, &item.LastHeartbeatAt, &item.ActiveLeases); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "nodes_lookup_failed", "Failed to scan node list.", corrID, nil)
			return
		}
		nodes = append(nodes, map[string]any{
			"node_id":           item.NodeID,
			"operator_id":       item.OperatorID,
			"runtime_type":      item.RuntimeType,
			"runtime_version":   item.RuntimeVersion,
			"status":            item.Status,
			"last_heartbeat_at": item.LastHeartbeatAt.Time,
			"active_leases":     item.ActiveLeases.Int64,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"nodes":          nodes,
		"correlation_id": corrID,
	})
}

type policyRecord struct {
	PolicyID    string
	TenantID    string
	WorkspaceID sql.NullString
	Family      sql.NullString
	Version     int
	PolicyJSON  []byte
}

type createPolicyRequest struct {
	PolicyID         string         `json:"policy_id,omitempty"`
	WorkspaceID      string         `json:"workspace_id,omitempty"`
	Family           string         `json:"family,omitempty"`
	Version          int            `json:"version,omitempty"`
	PolicyJSON       map[string]any `json:"policy_json,omitempty"`
	Provider         string         `json:"provider,omitempty"`
	Model            string         `json:"model,omitempty"`
	MaxUSDPerTask    float64        `json:"max_usd_per_task,omitempty"`
	Retries          int            `json:"retries,omitempty"`
	RequiresApproval bool           `json:"requires_approval,omitempty"`
}

type patchPolicyRequest struct {
	WorkspaceID      *string        `json:"workspace_id,omitempty"`
	Family           *string        `json:"family,omitempty"`
	Version          *int           `json:"version,omitempty"`
	PolicyJSON       map[string]any `json:"policy_json,omitempty"`
	Provider         *string        `json:"provider,omitempty"`
	Model            *string        `json:"model,omitempty"`
	MaxUSDPerTask    *float64       `json:"max_usd_per_task,omitempty"`
	Retries          *int           `json:"retries,omitempty"`
	RequiresApproval *bool          `json:"requires_approval,omitempty"`
}

func (s *Server) handleGetPoliciesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("platform_admin", "tenant_admin", "operator", "approver", "finance_viewer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for policy read.", corrID, nil)
		return
	}

	const q = "SELECT policy_id, tenant_id, workspace_id, family, version, policy_json FROM policy_bundles WHERE tenant_id = $1 ORDER BY family, policy_id"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "policy_lookup_failed", "Failed to fetch policies.", corrID, nil)
		return
	}
	defer rows.Close()

	policies := make([]map[string]any, 0)
	for rows.Next() {
		var rec policyRecord
		if err := rows.Scan(&rec.PolicyID, &rec.TenantID, &rec.WorkspaceID, &rec.Family, &rec.Version, &rec.PolicyJSON); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "policy_lookup_failed", "Failed to scan policies.", corrID, nil)
			return
		}
		policyMap := map[string]any{}
		if err := json.Unmarshal(rec.PolicyJSON, &policyMap); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "policy_lookup_failed", "Failed to decode policy JSON.", corrID, nil)
			return
		}
		policies = append(policies, serializePolicyRecord(rec, policyMap))
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"policies":       policies,
		"correlation_id": corrID,
	})
}

func (s *Server) handleCreatePolicyV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("platform_admin", "tenant_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for policy create.", corrID, nil)
		return
	}

	var req createPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if strings.TrimSpace(req.Family) == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "family is required.", corrID, nil)
		return
	}

	if req.Version <= 0 {
		req.Version = 1
	}
	policyID := strings.TrimSpace(req.PolicyID)
	if policyID == "" {
		policyID = "pol_" + randomID(6)
	}

	policyMap := mergedPolicyMap(req.PolicyJSON, req.Provider, req.Model, req.MaxUSDPerTask, req.Retries, req.RequiresApproval)
	policyJSON, err := json.Marshal(policyMap)
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "policy_json is invalid.", corrID, nil)
		return
	}

	const q = "INSERT INTO policy_bundles (policy_id, tenant_id, workspace_id, family, version, policy_json) VALUES ($1,$2,$3,$4,$5,$6)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, policyID, actor.TenantID, nullableString(req.WorkspaceID), req.Family, req.Version, policyJSON); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "policy_create_failed", "Failed to create policy.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "policy", policyID, "policy.created", map[string]any{
		"family":  req.Family,
		"version": req.Version,
	})

	rec := policyRecord{
		PolicyID:    policyID,
		TenantID:    actor.TenantID,
		WorkspaceID: nullableString(req.WorkspaceID),
		Family:      nullableString(req.Family),
		Version:     req.Version,
		PolicyJSON:  policyJSON,
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"policy":         serializePolicyRecord(rec, policyMap),
		"correlation_id": corrID,
	})
}

func (s *Server) handlePatchPolicyV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("platform_admin", "tenant_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for policy patch.", corrID, nil)
		return
	}

	policyID := strings.TrimSpace(r.PathValue("policy_id"))
	if policyID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "policy_id is required.", corrID, nil)
		return
	}

	var req patchPolicyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	const selectQ = "SELECT policy_id, tenant_id, workspace_id, family, version, policy_json FROM policy_bundles WHERE policy_id = $1 AND tenant_id = $2"
	var rec policyRecord
	if err := s.postgres.DB.QueryRowContext(r.Context(), selectQ, policyID, actor.TenantID).Scan(&rec.PolicyID, &rec.TenantID, &rec.WorkspaceID, &rec.Family, &rec.Version, &rec.PolicyJSON); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "policy_not_found", "Policy not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "policy_patch_failed", "Failed to load policy.", corrID, nil)
		return
	}

	policyMap := map[string]any{}
	if err := json.Unmarshal(rec.PolicyJSON, &policyMap); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "policy_patch_failed", "Failed to decode current policy JSON.", corrID, nil)
		return
	}
	for k, v := range req.PolicyJSON {
		policyMap[k] = v
	}
	if req.Provider != nil {
		policyMap["provider"] = strings.TrimSpace(*req.Provider)
	}
	if req.Model != nil {
		policyMap["model"] = strings.TrimSpace(*req.Model)
	}
	if req.MaxUSDPerTask != nil {
		policyMap["max_usd_per_task"] = *req.MaxUSDPerTask
	}
	if req.Retries != nil {
		policyMap["retries"] = *req.Retries
	}
	if req.RequiresApproval != nil {
		policyMap["requires_approval"] = *req.RequiresApproval
	}

	workspaceID := rec.WorkspaceID
	if req.WorkspaceID != nil {
		workspaceID = nullableString(*req.WorkspaceID)
	}
	family := rec.Family
	if req.Family != nil {
		family = nullableString(*req.Family)
	}
	version := rec.Version
	if req.Version != nil && *req.Version > 0 {
		version = *req.Version
	}

	updatedJSON, err := json.Marshal(policyMap)
	if err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "policy_json is invalid.", corrID, nil)
		return
	}

	const updateQ = "UPDATE policy_bundles SET workspace_id = $1, family = $2, version = $3, policy_json = $4 WHERE policy_id = $5 AND tenant_id = $6"
	result, err := s.postgres.DB.ExecContext(r.Context(), updateQ, workspaceID, family, version, updatedJSON, policyID, actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "policy_patch_failed", "Failed to update policy.", corrID, nil)
		return
	}
	if rows, _ := result.RowsAffected(); rows == 0 {
		apierrors.Write(w, http.StatusNotFound, "policy_not_found", "Policy not found.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "policy", policyID, "policy.updated", map[string]any{
		"version": version,
	})

	rec.WorkspaceID = workspaceID
	rec.Family = family
	rec.Version = version
	rec.PolicyJSON = updatedJSON
	writeJSON(w, http.StatusOK, map[string]any{
		"policy":         serializePolicyRecord(rec, policyMap),
		"correlation_id": corrID,
	})
}

func mergedPolicyMap(base map[string]any, provider, model string, maxUSDPerTask float64, retries int, requiresApproval bool) map[string]any {
	out := map[string]any{}
	for k, v := range base {
		out[k] = v
	}
	if strings.TrimSpace(provider) != "" {
		out["provider"] = strings.TrimSpace(provider)
	}
	if strings.TrimSpace(model) != "" {
		out["model"] = strings.TrimSpace(model)
	}
	if maxUSDPerTask > 0 {
		out["max_usd_per_task"] = maxUSDPerTask
	}
	if retries > 0 {
		out["retries"] = retries
	}
	if requiresApproval {
		out["requires_approval"] = true
	}
	return out
}

func serializePolicyRecord(rec policyRecord, policyMap map[string]any) map[string]any {
	return map[string]any{
		"policy_id":         rec.PolicyID,
		"tenant_id":         rec.TenantID,
		"workspace_id":      rec.WorkspaceID.String,
		"family":            rec.Family.String,
		"version":           rec.Version,
		"policy_json":       policyMap,
		"provider":          stringFromMap(policyMap, "provider"),
		"model":             stringFromMap(policyMap, "model"),
		"max_usd_per_task":  numberFromMap(policyMap, "max_usd_per_task"),
		"retries":           int(numberFromMap(policyMap, "retries")),
		"requires_approval": boolFromMap(policyMap, "requires_approval"),
	}
}

func nullableString(v string) sql.NullString {
	clean := strings.TrimSpace(v)
	if clean == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: clean, Valid: true}
}

func stringFromMap(values map[string]any, key string) string {
	v, ok := values[key]
	if !ok {
		return ""
	}
	switch typed := v.(type) {
	case string:
		return typed
	default:
		return ""
	}
}

func numberFromMap(values map[string]any, key string) float64 {
	v, ok := values[key]
	if !ok {
		return 0
	}
	switch typed := v.(type) {
	case float64:
		return typed
	case float32:
		return float64(typed)
	case int:
		return float64(typed)
	case int64:
		return float64(typed)
	case json.Number:
		f, _ := typed.Float64()
		return f
	default:
		return 0
	}
}

func boolFromMap(values map[string]any, key string) bool {
	v, ok := values[key]
	if !ok {
		return false
	}
	switch typed := v.(type) {
	case bool:
		return typed
	default:
		return false
	}
}
