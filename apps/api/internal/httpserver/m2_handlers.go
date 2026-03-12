package httpserver

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/MikeS071/archonhq/pkg/apierrors"
	"github.com/MikeS071/archonhq/pkg/auth"
	"github.com/MikeS071/archonhq/pkg/domain"
	"github.com/MikeS071/archonhq/pkg/events"
	"github.com/MikeS071/archonhq/pkg/telemetry"
)

var allowedSignupModes = map[string]struct{}{
	"open":     {},
	"invite":   {},
	"approval": {},
	"mixed":    {},
}

var minimumRoles = map[string]struct{}{
	"platform_admin": {},
	"tenant_admin":   {},
	"operator":       {},
	"approver":       {},
	"auditor":        {},
	"finance_viewer": {},
	"developer":      {},
}

type createTenantRequest struct {
	TenantID   string `json:"tenant_id"`
	Name       string `json:"name"`
	SignupMode string `json:"signup_mode"`
}

func (s *Server) handleCreateTenantV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for tenant creation.", corrID, nil)
		return
	}

	var req createTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	req.Name = strings.TrimSpace(req.Name)
	req.SignupMode = strings.TrimSpace(req.SignupMode)
	if req.Name == "" || req.SignupMode == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "name and signup_mode are required.", corrID, nil)
		return
	}
	if _, ok := allowedSignupModes[req.SignupMode]; !ok {
		apierrors.Write(w, http.StatusBadRequest, "invalid_signup_mode", "Unsupported signup_mode.", corrID, map[string]any{"signup_mode": req.SignupMode})
		return
	}

	tenantID := strings.TrimSpace(req.TenantID)
	if tenantID == "" {
		tenantID = "ten_" + randomID(6)
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	const insertTenant = "INSERT INTO tenants (tenant_id, name, signup_mode, status) VALUES ($1,$2,$3,$4)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertTenant, tenantID, req.Name, req.SignupMode, "active"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "tenant_create_failed", "Failed to create tenant.", corrID, nil)
		return
	}

	const insertMembership = "INSERT INTO memberships (membership_id, tenant_id, user_id, role) VALUES ($1,$2,$3,$4)"
	membershipID := "mem_" + randomID(6)
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertMembership, membershipID, tenantID, actor.ID, "tenant_admin"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "membership_create_failed", "Failed to create tenant membership.", corrID, nil)
		return
	}

	s.appendEvent(r, tenantID, "tenant", tenantID, "tenant.created", map[string]any{"name": req.Name, "signup_mode": req.SignupMode})
	s.appendEvent(r, tenantID, "membership", membershipID, "membership.created", map[string]any{"user_id": actor.ID, "role": "tenant_admin"})

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id":      tenantID,
		"name":           req.Name,
		"signup_mode":    req.SignupMode,
		"status":         "active",
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetTenantV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	tenantID := r.PathValue("tenant_id")
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	const q = "SELECT tenant_id, name, signup_mode, status, created_at FROM tenants WHERE tenant_id = $1"
	var id, name, signupMode, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, tenantID).Scan(&id, &name, &signupMode, &status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "tenant_not_found", "Tenant not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "tenant_lookup_failed", "Failed to fetch tenant.", corrID, nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id":      id,
		"name":           name,
		"signup_mode":    signupMode,
		"status":         status,
		"created_at":     createdAt,
		"correlation_id": corrID,
	})
}

type patchTenantRequest struct {
	Name       *string `json:"name,omitempty"`
	SignupMode *string `json:"signup_mode,omitempty"`
	Status     *string `json:"status,omitempty"`
}

func (s *Server) handlePatchTenantV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for tenant patch.", corrID, nil)
		return
	}

	tenantID := r.PathValue("tenant_id")
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	const getQ = "SELECT tenant_id, name, signup_mode, status, created_at FROM tenants WHERE tenant_id = $1"
	var id, name, signupMode, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), getQ, tenantID).Scan(&id, &name, &signupMode, &status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "tenant_not_found", "Tenant not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "tenant_lookup_failed", "Failed to fetch tenant.", corrID, nil)
		return
	}

	var req patchTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if req.Name != nil {
		name = strings.TrimSpace(*req.Name)
	}
	if req.SignupMode != nil {
		signupMode = strings.TrimSpace(*req.SignupMode)
		if _, ok := allowedSignupModes[signupMode]; !ok {
			apierrors.Write(w, http.StatusBadRequest, "invalid_signup_mode", "Unsupported signup_mode.", corrID, nil)
			return
		}
	}
	if req.Status != nil {
		status = strings.TrimSpace(*req.Status)
	}

	const updateQ = "UPDATE tenants SET name = $1, signup_mode = $2, status = $3 WHERE tenant_id = $4"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateQ, name, signupMode, status, tenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "tenant_update_failed", "Failed to patch tenant.", corrID, nil)
		return
	}

	s.appendEvent(r, tenantID, "tenant", tenantID, "tenant.updated", map[string]any{"name": name, "signup_mode": signupMode, "status": status})

	writeJSON(w, http.StatusOK, map[string]any{
		"tenant_id":      id,
		"name":           name,
		"signup_mode":    signupMode,
		"status":         status,
		"created_at":     createdAt,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetTenantMembersV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	tenantID := r.PathValue("tenant_id")
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	const q = "SELECT membership_id, user_id, role, created_at FROM memberships WHERE tenant_id = $1 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, tenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "members_lookup_failed", "Failed to list memberships.", corrID, nil)
		return
	}
	defer rows.Close()

	members := []map[string]any{}
	for rows.Next() {
		var membershipID, userID, role string
		var createdAt time.Time
		if err := rows.Scan(&membershipID, &userID, &role, &createdAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "members_lookup_failed", "Failed to scan memberships.", corrID, nil)
			return
		}
		members = append(members, map[string]any{
			"membership_id": membershipID,
			"user_id":       userID,
			"role":          role,
			"created_at":    createdAt,
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{"members": members, "correlation_id": corrID})
}

type createWorkspaceRequest struct {
	WorkspaceID string `json:"workspace_id"`
	TenantID    string `json:"tenant_id"`
	Name        string `json:"name"`
}

func (s *Server) handleCreateWorkspaceV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for workspace creation.", corrID, nil)
		return
	}

	var req createWorkspaceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.WorkspaceID = strings.TrimSpace(req.WorkspaceID)
	req.Name = strings.TrimSpace(req.Name)
	if req.TenantID == "" || req.Name == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "tenant_id and name are required.", corrID, nil)
		return
	}
	if req.WorkspaceID == "" {
		req.WorkspaceID = "ws_" + randomID(6)
	}
	if !s.ensureTenantAccess(actor, req.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	const insertQ = "INSERT INTO workspaces (workspace_id, tenant_id, name, status) VALUES ($1,$2,$3,$4)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertQ, req.WorkspaceID, req.TenantID, req.Name, "active"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "workspace_create_failed", "Failed to create workspace.", corrID, nil)
		return
	}

	const getQ = "SELECT workspace_id, tenant_id, name, status, created_at FROM workspaces WHERE workspace_id = $1 AND tenant_id = $2"
	var wsID, tenantID, name, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), getQ, req.WorkspaceID, req.TenantID).Scan(&wsID, &tenantID, &name, &status, &createdAt); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "workspace_lookup_failed", "Failed to fetch workspace.", corrID, nil)
		return
	}

	s.appendEvent(r, req.TenantID, "workspace", req.WorkspaceID, "workspace.created", map[string]any{"name": req.Name})

	writeJSON(w, http.StatusOK, map[string]any{
		"workspace_id":   wsID,
		"tenant_id":      tenantID,
		"name":           name,
		"status":         status,
		"created_at":     createdAt,
		"correlation_id": corrID,
	})
}

func (s *Server) handleGetWorkspaceV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	workspaceID := r.PathValue("workspace_id")
	const q = "SELECT workspace_id, tenant_id, name, status, created_at FROM workspaces WHERE workspace_id = $1"
	var wsID, tenantID, name, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, workspaceID).Scan(&wsID, &tenantID, &name, &status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "workspace_not_found", "Workspace not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "workspace_lookup_failed", "Failed to fetch workspace.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"workspace_id": wsID, "tenant_id": tenantID, "name": name, "status": status, "created_at": createdAt, "correlation_id": corrID})
}

func (s *Server) handleWorkspaceSummaryV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	workspaceID := r.PathValue("workspace_id")
	tenantID := actor.TenantID
	var tasksTotal, pendingApprovals int64
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM tasks WHERE tenant_id = $1 AND workspace_id = $2", tenantID, workspaceID).Scan(&tasksTotal)
	_ = s.postgres.DB.QueryRowContext(r.Context(), "SELECT COUNT(*) FROM approval_requests WHERE tenant_id = $1 AND status = 'pending'", tenantID).Scan(&pendingApprovals)
	writeJSON(w, http.StatusOK, map[string]any{"workspace_id": workspaceID, "tasks_total": tasksTotal, "pending_approvals": pendingApprovals, "correlation_id": corrID})
}

func (s *Server) handleWorkspaceTasksV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	workspaceID := r.PathValue("workspace_id")
	const q = "SELECT task_id, task_family, title, status, created_at FROM tasks WHERE tenant_id = $1 AND workspace_id = $2 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, workspaceID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "workspace_tasks_failed", "Failed to list workspace tasks.", corrID, nil)
		return
	}
	defer rows.Close()

	tasks := []map[string]any{}
	for rows.Next() {
		var taskID, family, title, status string
		var createdAt time.Time
		if err := rows.Scan(&taskID, &family, &title, &status, &createdAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "workspace_tasks_failed", "Failed to scan tasks.", corrID, nil)
			return
		}
		tasks = append(tasks, map[string]any{"task_id": taskID, "task_family": family, "title": title, "status": status, "created_at": createdAt})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks, "correlation_id": corrID})
}

func (s *Server) handleWorkspaceLedgerV2(w http.ResponseWriter, r *http.Request) {
	_, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"entries": []map[string]any{}, "correlation_id": corrID})
}

type nodeRegisterIntentRequest struct {
	ChallengeID string `json:"challenge_id"`
	TenantID    string `json:"tenant_id"`
	OperatorID  string `json:"operator_id"`
	RuntimeType string `json:"runtime_type"`
}

func (s *Server) handleNodeRegisterIntentV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for node registration intent.", corrID, nil)
		return
	}

	var req nodeRegisterIntentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.OperatorID = strings.TrimSpace(req.OperatorID)
	req.RuntimeType = strings.TrimSpace(req.RuntimeType)
	if req.TenantID == "" || req.OperatorID == "" || req.RuntimeType == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "tenant_id, operator_id, and runtime_type are required.", corrID, nil)
		return
	}
	if req.RuntimeType != "hermes" {
		apierrors.Write(w, http.StatusBadRequest, "runtime_not_allowed", "Only Hermes is allowed in v1.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, req.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	challengeID := strings.TrimSpace(req.ChallengeID)
	if challengeID == "" {
		challengeID = "ch_" + randomID(6)
	}
	challengeNonce := "challenge_" + randomID(6)
	expiresAt := time.Now().UTC().Add(10 * time.Minute)

	const q = "INSERT INTO node_registration_challenges (challenge_id, tenant_id, operator_id, requested_runtime, challenge_nonce, expires_at, status) VALUES ($1,$2,$3,$4,$5,$6,$7)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, challengeID, req.TenantID, req.OperatorID, req.RuntimeType, challengeNonce, expiresAt, "pending"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "node_register_intent_failed", "Failed to create node registration intent.", corrID, nil)
		return
	}

	s.appendEvent(r, req.TenantID, "node", challengeID, "node.register_intent_created", map[string]any{"operator_id": req.OperatorID, "runtime_type": req.RuntimeType})

	writeJSON(w, http.StatusOK, map[string]any{
		"challenge_id":   challengeID,
		"challenge":      challengeNonce,
		"expires_at":     expiresAt,
		"correlation_id": corrID,
	})
}

type nodeRegisterRequest struct {
	ChallengeID    string `json:"challenge_id"`
	NodeID         string `json:"node_id"`
	TenantID       string `json:"tenant_id"`
	OperatorID     string `json:"operator_id"`
	PublicKey      string `json:"public_key"`
	RuntimeType    string `json:"runtime_type"`
	RuntimeVersion string `json:"runtime_version"`
	Signature      string `json:"signature"`
}

func (s *Server) handleNodeRegisterV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "platform_admin", "operator") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for node register.", corrID, nil)
		return
	}

	var req nodeRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}

	req.ChallengeID = strings.TrimSpace(req.ChallengeID)
	req.TenantID = strings.TrimSpace(req.TenantID)
	req.OperatorID = strings.TrimSpace(req.OperatorID)
	req.PublicKey = strings.TrimSpace(req.PublicKey)
	req.RuntimeType = strings.TrimSpace(req.RuntimeType)
	req.RuntimeVersion = strings.TrimSpace(req.RuntimeVersion)
	req.Signature = strings.TrimSpace(req.Signature)
	if req.ChallengeID == "" || req.TenantID == "" || req.OperatorID == "" || req.PublicKey == "" || req.RuntimeType == "" || req.RuntimeVersion == "" || req.Signature == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "challenge_id, tenant_id, operator_id, public_key, runtime_type, runtime_version, and signature are required.", corrID, nil)
		return
	}
	if req.RuntimeType != "hermes" {
		apierrors.Write(w, http.StatusBadRequest, "runtime_not_allowed", "Only Hermes is allowed in v1.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, req.TenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	const getChallengeQ = "SELECT tenant_id, operator_id, challenge_nonce, expires_at, status FROM node_registration_challenges WHERE challenge_id = $1"
	var challengeTenantID, challengeOperatorID, challengeNonce, challengeStatus string
	var challengeExpiresAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), getChallengeQ, req.ChallengeID).Scan(&challengeTenantID, &challengeOperatorID, &challengeNonce, &challengeExpiresAt, &challengeStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "node_registration_failed", "Challenge not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "node_registration_failed", "Failed to read challenge.", corrID, nil)
		return
	}
	if challengeStatus != "pending" || time.Now().UTC().After(challengeExpiresAt) {
		apierrors.Write(w, http.StatusBadRequest, "node_registration_failed", "Challenge is expired or already used.", corrID, nil)
		return
	}
	if challengeTenantID != req.TenantID || challengeOperatorID != req.OperatorID {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Challenge does not match tenant/operator.", corrID, nil)
		return
	}
	if req.Signature != "signed:"+challengeNonce {
		apierrors.Write(w, http.StatusBadRequest, "invalid_signature", "Invalid registration signature.", corrID, nil)
		return
	}

	const updateChallengeQ = "UPDATE node_registration_challenges SET status = $1 WHERE challenge_id = $2"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateChallengeQ, "used", req.ChallengeID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "node_registration_failed", "Failed to update challenge status.", corrID, nil)
		return
	}

	nodeID := strings.TrimSpace(req.NodeID)
	if nodeID == "" {
		nodeID = "node_" + randomID(6)
	}

	const insertNodeQ = "INSERT INTO nodes (node_id, tenant_id, operator_id, public_key, runtime_type, runtime_version, status) VALUES ($1,$2,$3,$4,$5,$6,$7)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertNodeQ, nodeID, req.TenantID, req.OperatorID, req.PublicKey, req.RuntimeType, req.RuntimeVersion, "active"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "node_registration_failed", "Failed to register node.", corrID, nil)
		return
	}

	credentialID := "cred_" + randomID(6)
	nodeToken := fmt.Sprintf("node:%s:%s:%s", req.TenantID, nodeID, credentialID)
	tokenHash := hashString(nodeToken)

	const insertCredentialQ = "INSERT INTO node_credentials (credential_id, tenant_id, node_id, token_hash, status) VALUES ($1,$2,$3,$4,$5)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertCredentialQ, credentialID, req.TenantID, nodeID, tokenHash, "active"); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "node_registration_failed", "Failed to issue node credential.", corrID, nil)
		return
	}

	s.appendEvent(r, req.TenantID, "node", nodeID, "node.registered", map[string]any{"operator_id": req.OperatorID, "runtime_type": req.RuntimeType})

	writeJSON(w, http.StatusOK, map[string]any{
		"node_id":        nodeID,
		"credential_id":  credentialID,
		"node_token":     nodeToken,
		"correlation_id": corrID,
	})
}

func (s *Server) handleNodeHeartbeatV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	nodeID := r.PathValue("node_id")
	if actor.ID != nodeID {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Node token does not match requested node.", corrID, nil)
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}

	const q = "UPDATE nodes SET last_heartbeat_at = $1 WHERE node_id = $2 AND tenant_id = $3"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, time.Now().UTC(), nodeID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "heartbeat_failed", "Failed to update heartbeat.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "node", nodeID, "node.heartbeat", map[string]any{})
	writeJSON(w, http.StatusOK, map[string]any{"node_id": nodeID, "status": "ok", "correlation_id": corrID})
}

func (s *Server) handleGetNodeV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	nodeID := r.PathValue("node_id")
	const q = "SELECT node_id, tenant_id, operator_id, public_key, runtime_type, runtime_version, status, last_heartbeat_at FROM nodes WHERE node_id = $1"
	var id, tenantID, operatorID, publicKey, runtimeType, runtimeVersion, status string
	var lastHeartbeatAt sql.NullTime
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, nodeID).Scan(&id, &tenantID, &operatorID, &publicKey, &runtimeType, &runtimeVersion, &status, &lastHeartbeatAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "node_not_found", "Node not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "node_lookup_failed", "Failed to fetch node.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, tenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant read is not allowed.", corrID, nil)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"node_id": id, "tenant_id": tenantID, "operator_id": operatorID, "public_key": publicKey, "runtime_type": runtimeType, "runtime_version": runtimeVersion, "status": status, "last_heartbeat_at": lastHeartbeatAt.Time, "correlation_id": corrID})
}

func (s *Server) handleGetNodeLeasesV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	nodeID := r.PathValue("node_id")
	const q = "SELECT lease_id, task_id, node_id, status, approval_state, granted_at, expires_at FROM leases WHERE node_id = $1 AND tenant_id = $2 ORDER BY granted_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, nodeID, actor.TenantID)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "leases_lookup_failed", "Failed to fetch node leases.", corrID, nil)
		return
	}
	defer rows.Close()

	leases := []map[string]any{}
	for rows.Next() {
		var leaseID, taskID, leaseNodeID, status, approvalState string
		var grantedAt, expiresAt sql.NullTime
		if err := rows.Scan(&leaseID, &taskID, &leaseNodeID, &status, &approvalState, &grantedAt, &expiresAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "leases_lookup_failed", "Failed to scan leases.", corrID, nil)
			return
		}
		leases = append(leases, map[string]any{"lease_id": leaseID, "task_id": taskID, "node_id": leaseNodeID, "status": status, "approval_state": approvalState, "granted_at": grantedAt.Time, "expires_at": expiresAt.Time})
	}
	writeJSON(w, http.StatusOK, map[string]any{"leases": leases, "correlation_id": corrID})
}

type createTaskRequestV2 struct {
	TaskID       string `json:"task_id"`
	WorkspaceID  string `json:"workspace_id"`
	TaskFamily   string `json:"task_family"`
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	SchemaRef    string `json:"schema_ref,omitempty"`
	ApprovalMode string `json:"approval_mode,omitempty"`
}

func (s *Server) handleCreateTaskV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "developer") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for task creation.", corrID, nil)
		return
	}

	var req createTaskRequestV2
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if strings.TrimSpace(req.WorkspaceID) == "" || strings.TrimSpace(req.Title) == "" || strings.TrimSpace(req.TaskFamily) == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "workspace_id, task_family, and title are required.", corrID, nil)
		return
	}
	if _, ok := domain.SupportedTaskFamilies[req.TaskFamily]; !ok {
		apierrors.Write(w, http.StatusBadRequest, "invalid_task_family", "Unsupported task family.", corrID, map[string]any{"task_family": req.TaskFamily})
		return
	}
	const workspaceTenantQ = "SELECT tenant_id FROM workspaces WHERE workspace_id = $1"
	var workspaceTenantID string
	if err := s.postgres.DB.QueryRowContext(r.Context(), workspaceTenantQ, req.WorkspaceID).Scan(&workspaceTenantID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "workspace_not_found", "Workspace not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "workspace_lookup_failed", "Failed to validate workspace.", corrID, nil)
		return
	}
	if !s.ensureTenantAccess(actor, workspaceTenantID) {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Cross-tenant write is not allowed.", corrID, nil)
		return
	}

	taskID := strings.TrimSpace(req.TaskID)
	if taskID == "" {
		taskID = "task_" + randomID(6)
	}
	approvalMode := strings.TrimSpace(req.ApprovalMode)
	if approvalMode == "" {
		approvalMode = "always_required"
	}
	approvalID := approvalIDFromTaskID(taskID)
	status := "awaiting_approval"

	const insertTaskQ = "INSERT INTO tasks (task_id, tenant_id, workspace_id, task_family, title, description, status, schema_ref, approval_mode, created_by) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertTaskQ, taskID, actor.TenantID, req.WorkspaceID, req.TaskFamily, req.Title, req.Description, status, req.SchemaRef, approvalMode, actor.ID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "task_create_failed", "Failed to create task.", corrID, nil)
		return
	}

	const insertApprovalQ = "INSERT INTO approval_requests (approval_id, tenant_id, task_id, status, requested_by) VALUES ($1,$2,$3,$4,$5)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), insertApprovalQ, approvalID, actor.TenantID, taskID, "pending", actor.ID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "approval_create_failed", "Failed to create approval request.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "task", taskID, "task.created", map[string]any{"workspace_id": req.WorkspaceID, "task_family": req.TaskFamily, "title": req.Title})
	s.appendEvent(r, actor.TenantID, "approval", approvalID, "approval.requested", map[string]any{"task_id": taskID})

	writeJSON(w, http.StatusOK, map[string]any{"task_id": taskID, "status": status, "approval_request_id": approvalID, "correlation_id": corrID})
}

func (s *Server) handleGetTaskV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	taskID := r.PathValue("task_id")
	const q = "SELECT task_id, tenant_id, workspace_id, task_family, title, description, status, created_at FROM tasks WHERE task_id = $1 AND tenant_id = $2"
	var id, tenantID, workspaceID, taskFamily, title, description, status string
	var createdAt time.Time
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, taskID, actor.TenantID).Scan(&id, &tenantID, &workspaceID, &taskFamily, &title, &description, &status, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to fetch task.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"task_id": id, "tenant_id": tenantID, "workspace_id": workspaceID, "task_family": taskFamily, "title": title, "description": description, "status": status, "created_at": createdAt, "correlation_id": corrID})
}

func (s *Server) handleTaskFeedV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	limit := 50
	if rawLimit := strings.TrimSpace(r.URL.Query().Get("limit")); rawLimit != "" {
		if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}
	const q = "SELECT task_id, workspace_id, task_family, title, status, created_at FROM tasks WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, limit)
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "task_feed_failed", "Failed to fetch task feed.", corrID, nil)
		return
	}
	defer rows.Close()

	tasks := []map[string]any{}
	for rows.Next() {
		var taskID, workspaceID, family, title, status string
		var createdAt time.Time
		if err := rows.Scan(&taskID, &workspaceID, &family, &title, &status, &createdAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "task_feed_failed", "Failed to scan task feed.", corrID, nil)
			return
		}
		tasks = append(tasks, map[string]any{"task_id": taskID, "workspace_id": workspaceID, "task_family": family, "title": title, "status": status, "created_at": createdAt})
	}
	writeJSON(w, http.StatusOK, map[string]any{"tasks": tasks, "correlation_id": corrID})
}

func (s *Server) handleCancelTaskV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("tenant_admin", "operator", "approver") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for task cancel.", corrID, nil)
		return
	}
	taskID := r.PathValue("task_id")
	const q = "UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, "cancelled", taskID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "task_cancel_failed", "Failed to cancel task.", corrID, nil)
		return
	}
	s.appendEvent(r, actor.TenantID, "task", taskID, "task.cancelled", map[string]any{})
	writeJSON(w, http.StatusOK, map[string]any{"task_id": taskID, "status": "cancelled", "correlation_id": corrID})
}

func (s *Server) handleApprovalQueueV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT approval_id, task_id, status, created_at, decided_at FROM approval_requests WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC"
	rows, err := s.postgres.DB.QueryContext(r.Context(), q, actor.TenantID, "pending")
	if err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "approvals_lookup_failed", "Failed to fetch approval queue.", corrID, nil)
		return
	}
	defer rows.Close()

	approvals := []map[string]any{}
	for rows.Next() {
		var approvalID, taskID, status string
		var createdAt time.Time
		var decidedAt sql.NullTime
		if err := rows.Scan(&approvalID, &taskID, &status, &createdAt, &decidedAt); err != nil {
			apierrors.Write(w, http.StatusInternalServerError, "approvals_lookup_failed", "Failed to scan approval queue.", corrID, nil)
			return
		}
		approvals = append(approvals, map[string]any{"approval_id": approvalID, "task_id": taskID, "status": status, "created_at": createdAt, "decided_at": decidedAt.Time})
	}
	writeJSON(w, http.StatusOK, map[string]any{"approvals": approvals, "correlation_id": corrID})
}

func (s *Server) handleGetApprovalV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	const q = "SELECT approval_id, task_id, status, requested_by, decided_by, decision_reason, created_at, decided_at FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2"
	var approvalID, taskID, status string
	var requestedBy, decidedBy, decisionReason sql.NullString
	var createdAt time.Time
	var decidedAt sql.NullTime
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, r.PathValue("approval_id"), actor.TenantID).Scan(&approvalID, &taskID, &status, &requestedBy, &decidedBy, &decisionReason, &createdAt, &decidedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "approval_not_found", "Approval not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "approval_lookup_failed", "Failed to fetch approval.", corrID, nil)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"approval_id": approvalID, "task_id": taskID, "status": status, "requested_by": requestedBy.String, "decided_by": decidedBy.String, "decision_reason": decisionReason.String, "created_at": createdAt, "decided_at": decidedAt.Time, "correlation_id": corrID})
}

func (s *Server) handleApproveV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for approval decision.", corrID, nil)
		return
	}

	approvalID := r.PathValue("approval_id")
	taskID := taskIDFromApprovalID(approvalID)
	if taskID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "approval_id format is invalid.", corrID, nil)
		return
	}
	const approvalQ = "SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2"
	var approvalStatus string
	if err := s.postgres.DB.QueryRowContext(r.Context(), approvalQ, approvalID, actor.TenantID).Scan(&taskID, &approvalStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "approval_not_found", "Approval not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "approval_lookup_failed", "Failed to fetch approval.", corrID, nil)
		return
	}
	if approvalStatus != "pending" {
		apierrors.Write(w, http.StatusConflict, "approval_not_pending", "Approval is not pending.", corrID, nil)
		return
	}

	const updateApprovalQ = "UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateApprovalQ, "approved", actor.ID, time.Now().UTC(), approvalID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "approval_update_failed", "Failed to approve request.", corrID, nil)
		return
	}
	const updateTaskQ = "UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateTaskQ, "approved", taskID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "task_update_failed", "Failed to update task status.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "approval", approvalID, "approval.approved", map[string]any{"task_id": taskID})
	writeJSON(w, http.StatusOK, map[string]any{"approval_id": approvalID, "task_id": taskID, "status": "approved", "correlation_id": corrID})
}

func (s *Server) handleDenyV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "tenant_admin", "platform_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for approval decision.", corrID, nil)
		return
	}
	approvalID := r.PathValue("approval_id")
	taskID := taskIDFromApprovalID(approvalID)
	if taskID == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "approval_id format is invalid.", corrID, nil)
		return
	}
	const approvalQ = "SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2"
	var approvalStatus string
	if err := s.postgres.DB.QueryRowContext(r.Context(), approvalQ, approvalID, actor.TenantID).Scan(&taskID, &approvalStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "approval_not_found", "Approval not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "approval_lookup_failed", "Failed to fetch approval.", corrID, nil)
		return
	}
	if approvalStatus != "pending" {
		apierrors.Write(w, http.StatusConflict, "approval_not_pending", "Approval is not pending.", corrID, nil)
		return
	}
	const updateApprovalQ = "UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateApprovalQ, "denied", actor.ID, time.Now().UTC(), approvalID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "approval_update_failed", "Failed to deny request.", corrID, nil)
		return
	}
	const updateTaskQ = "UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3"
	if _, err := s.postgres.DB.ExecContext(r.Context(), updateTaskQ, "denied", taskID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "task_update_failed", "Failed to update task status.", corrID, nil)
		return
	}
	s.appendEvent(r, actor.TenantID, "approval", approvalID, "approval.denied", map[string]any{"task_id": taskID})
	writeJSON(w, http.StatusOK, map[string]any{"approval_id": approvalID, "task_id": taskID, "status": "denied", "correlation_id": corrID})
}

type createLeaseRequest struct {
	LeaseID          string         `json:"lease_id"`
	TaskID           string         `json:"task_id"`
	NodeID           string         `json:"node_id"`
	ExpiresInSeconds int            `json:"expires_in_seconds"`
	ExecutionPolicy  map[string]any `json:"execution_policy"`
}

func (s *Server) handleCreateLeaseV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !actor.HasAnyRole("approver", "operator", "tenant_admin") {
		apierrors.Write(w, http.StatusForbidden, "forbidden", "Insufficient role for lease creation.", corrID, nil)
		return
	}

	var req createLeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if strings.TrimSpace(req.TaskID) == "" || strings.TrimSpace(req.NodeID) == "" {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "task_id and node_id are required.", corrID, nil)
		return
	}
	const taskLookupQ = "SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2"
	var taskStatus string
	if err := s.postgres.DB.QueryRowContext(r.Context(), taskLookupQ, req.TaskID, actor.TenantID).Scan(&taskStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "task_not_found", "Task not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "task_lookup_failed", "Failed to validate task.", corrID, nil)
		return
	}
	if taskStatus != "approved" {
		apierrors.Write(w, http.StatusConflict, "task_not_approved", "Task must be approved before lease creation.", corrID, nil)
		return
	}
	const nodeLookupQ = "SELECT status FROM nodes WHERE node_id = $1 AND tenant_id = $2"
	var nodeStatus string
	if err := s.postgres.DB.QueryRowContext(r.Context(), nodeLookupQ, req.NodeID, actor.TenantID).Scan(&nodeStatus); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			apierrors.Write(w, http.StatusNotFound, "node_not_found", "Node not found.", corrID, nil)
			return
		}
		apierrors.Write(w, http.StatusInternalServerError, "node_lookup_failed", "Failed to validate node.", corrID, nil)
		return
	}
	if nodeStatus != "active" {
		apierrors.Write(w, http.StatusConflict, "node_not_active", "Node must be active before lease creation.", corrID, nil)
		return
	}
	leaseID := strings.TrimSpace(req.LeaseID)
	if leaseID == "" {
		leaseID = "lease_" + randomID(6)
	}
	expiresIn := req.ExpiresInSeconds
	if expiresIn <= 0 {
		expiresIn = 600
	}
	grantedAt := time.Now().UTC()
	expiresAt := grantedAt.Add(time.Duration(expiresIn) * time.Second)

	const q = "INSERT INTO leases (lease_id, tenant_id, task_id, node_id, attempt, status, approval_state, granted_at, expires_at) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, leaseID, actor.TenantID, req.TaskID, req.NodeID, 1, "granted", "approved", grantedAt, expiresAt); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "lease_create_failed", "Failed to create lease.", corrID, nil)
		return
	}

	s.appendEvent(r, actor.TenantID, "lease", leaseID, "lease.granted", map[string]any{"task_id": req.TaskID, "node_id": req.NodeID})
	writeJSON(w, http.StatusOK, map[string]any{"lease_id": leaseID, "status": "granted", "task_id": req.TaskID, "node_id": req.NodeID, "expires_at": expiresAt, "correlation_id": corrID})
}

func (s *Server) handleClaimLeaseV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}
	leaseID := r.PathValue("lease_id")
	const q = "UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, "claimed", leaseID, actor.ID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "lease_claim_failed", "Failed to claim lease.", corrID, nil)
		return
	}
	s.appendEvent(r, actor.TenantID, "lease", leaseID, "lease.claimed", map[string]any{"node_id": actor.ID})
	writeJSON(w, http.StatusOK, map[string]any{"lease_id": leaseID, "status": "claimed", "correlation_id": corrID})
}

func (s *Server) handleReleaseLeaseV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}
	leaseID := r.PathValue("lease_id")
	const q = "UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, "released", leaseID, actor.ID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "lease_release_failed", "Failed to release lease.", corrID, nil)
		return
	}
	s.appendEvent(r, actor.TenantID, "lease", leaseID, "lease.released", map[string]any{"node_id": actor.ID})
	writeJSON(w, http.StatusOK, map[string]any{"lease_id": leaseID, "status": "released", "correlation_id": corrID})
}

type extendLeaseRequest struct {
	ExtendSeconds int `json:"extend_seconds"`
}

func (s *Server) handleExtendLeaseV2(w http.ResponseWriter, r *http.Request) {
	actor, corrID, ok := s.requireActor(w, r)
	if !ok {
		return
	}
	if !s.validateNodeCredential(w, r, actor, corrID) {
		return
	}
	leaseID := r.PathValue("lease_id")

	var req extendLeaseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		apierrors.Write(w, http.StatusBadRequest, "invalid_request", "Invalid JSON payload.", corrID, nil)
		return
	}
	if req.ExtendSeconds <= 0 {
		req.ExtendSeconds = 300
	}
	expiresAt := time.Now().UTC().Add(time.Duration(req.ExtendSeconds) * time.Second)

	const q = "UPDATE leases SET expires_at = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4"
	if _, err := s.postgres.DB.ExecContext(r.Context(), q, expiresAt, leaseID, actor.ID, actor.TenantID); err != nil {
		apierrors.Write(w, http.StatusInternalServerError, "lease_extend_failed", "Failed to extend lease.", corrID, nil)
		return
	}
	s.appendEvent(r, actor.TenantID, "lease", leaseID, "lease.extended", map[string]any{"node_id": actor.ID, "expires_at": expiresAt})
	writeJSON(w, http.StatusOK, map[string]any{"lease_id": leaseID, "status": "extended", "expires_at": expiresAt, "correlation_id": corrID})
}

func (s *Server) requireActor(w http.ResponseWriter, r *http.Request) (auth.Actor, string, bool) {
	actor, ok := auth.ActorFromContext(r.Context())
	corrID := telemetry.CorrelationIDFromContext(r.Context())
	if !ok {
		apierrors.Write(w, http.StatusUnauthorized, "unauthorized", "Missing authenticated actor context.", corrID, nil)
		return auth.Actor{}, corrID, false
	}
	if err := validateActorRoles(actor); err != nil {
		apierrors.Write(w, http.StatusForbidden, "forbidden", err.Error(), corrID, nil)
		return auth.Actor{}, corrID, false
	}
	return actor, corrID, true
}

func validateActorRoles(actor auth.Actor) error {
	if actor.Type != "human" {
		return nil
	}
	for role := range actor.Roles {
		if _, ok := minimumRoles[role]; !ok {
			return fmt.Errorf("unsupported role: %s", role)
		}
	}
	return nil
}

func (s *Server) ensureTenantAccess(actor auth.Actor, tenantID string) bool {
	if actor.HasRole("platform_admin") {
		return true
	}
	return actor.TenantID == tenantID
}

func (s *Server) validateNodeCredential(w http.ResponseWriter, r *http.Request, actor auth.Actor, corrID string) bool {
	if actor.CredentialID == "" {
		apierrors.Write(w, http.StatusUnauthorized, "unauthorized", "Node credential id is required.", corrID, nil)
		return false
	}
	const q = "SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3"
	var status string
	var tokenHash string
	if err := s.postgres.DB.QueryRowContext(r.Context(), q, actor.CredentialID, actor.TenantID, actor.ID).Scan(&status, &tokenHash); err != nil {
		apierrors.Write(w, http.StatusUnauthorized, "unauthorized", "Node credential is invalid or revoked.", corrID, nil)
		return false
	}
	if status != "active" {
		apierrors.Write(w, http.StatusUnauthorized, "unauthorized", "Node credential is not active.", corrID, nil)
		return false
	}
	if hashString(actor.TokenRaw) != tokenHash {
		apierrors.Write(w, http.StatusUnauthorized, "unauthorized", "Node credential token does not match.", corrID, nil)
		return false
	}
	return true
}

func (s *Server) appendEvent(r *http.Request, tenantID, entityType, entityID, eventType string, payload map[string]any) {
	if s.events == nil {
		return
	}
	actor, _ := auth.ActorFromContext(r.Context())
	_ = s.events.Append(r.Context(), events.Envelope{
		EventID:        "evt_" + randomID(8),
		TenantID:       tenantID,
		EntityType:     entityType,
		EntityID:       entityID,
		EventType:      eventType,
		EventVersion:   1,
		ActorType:      actor.Type,
		ActorID:        actor.ID,
		CorrelationID:  telemetry.CorrelationIDFromContext(r.Context()),
		IdempotencyKey: r.Header.Get("Idempotency-Key"),
		Payload:        payload,
		OccurredAt:     time.Now().UTC(),
	})
}

func hashString(value string) string {
	sum := sha256.Sum256([]byte(value))
	return hex.EncodeToString(sum[:])
}

func approvalIDFromTaskID(taskID string) string {
	if strings.HasPrefix(taskID, "task_") {
		return "apr_" + strings.TrimPrefix(taskID, "task_")
	}
	return "apr_" + randomID(6)
}

func taskIDFromApprovalID(approvalID string) string {
	if strings.HasPrefix(approvalID, "apr_") {
		return "task_" + strings.TrimPrefix(approvalID, "apr_")
	}
	return ""
}
