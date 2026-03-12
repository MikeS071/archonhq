package httpserver

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/MikeS071/archonhq/pkg/auth"
	"github.com/MikeS071/archonhq/pkg/telemetry"
)

func newServerWithMock(t *testing.T) (*Server, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	return srv, mock, dbMock
}

func reqWithActor(method, path, body string, actor auth.Actor) *http.Request {
	req := httptest.NewRequest(method, path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "idem_unit")
	ctx := auth.WithActor(req.Context(), actor)
	ctx = telemetry.WithCorrelationID(ctx, "corr_unit")
	return req.WithContext(ctx)
}

func TestM2HandlersErrorBranchesDirect(t *testing.T) {
	actorAdmin := auth.Actor{Type: "human", ID: "user_1", TenantID: "ten_01", Roles: map[string]struct{}{"tenant_admin": {}}}
	actorDeveloper := auth.Actor{Type: "human", ID: "user_1", TenantID: "ten_01", Roles: map[string]struct{}{"developer": {}}}

	t.Run("create tenant forbidden role", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/tenants", `{"name":"x","signup_mode":"open"}`, actorDeveloper)
		s.handleCreateTenantV2(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("create tenant insert failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectExec("INSERT INTO tenants").WillReturnError(errors.New("insert failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/tenants", `{"tenant_id":"ten_01","name":"x","signup_mode":"open"}`, actorAdmin)
		s.handleCreateTenantV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("create tenant membership failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectExec("INSERT INTO tenants").WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec("INSERT INTO memberships").WillReturnError(errors.New("membership failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/tenants", `{"tenant_id":"ten_01","name":"x","signup_mode":"open"}`, actorAdmin)
		s.handleCreateTenantV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("get tenant not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT tenant_id, name, signup_mode, status, created_at FROM tenants").WithArgs("ten_01").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/tenants/ten_01", "", actorAdmin)
		req.SetPathValue("tenant_id", "ten_01")
		s.handleGetTenantV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("patch tenant update failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		now := time.Now().UTC()
		mock.ExpectQuery("SELECT tenant_id, name, signup_mode, status, created_at FROM tenants").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "name", "signup_mode", "status", "created_at"}).AddRow("ten_01", "x", "open", "active", now))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE tenants SET name = $1, signup_mode = $2, status = $3 WHERE tenant_id = $4")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPatch, "/v1/tenants/ten_01", `{}`, actorAdmin)
		req.SetPathValue("tenant_id", "ten_01")
		s.handlePatchTenantV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("members query error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT membership_id, user_id, role, created_at FROM memberships").WillReturnError(errors.New("query failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/tenants/ten_01/members", "", actorAdmin)
		req.SetPathValue("tenant_id", "ten_01")
		s.handleGetTenantMembersV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("workspace insert failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectExec("INSERT INTO workspaces").WillReturnError(errors.New("insert failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/workspaces", `{"tenant_id":"ten_01","name":"x"}`, actorAdmin)
		s.handleCreateWorkspaceV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("workspace task query error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT task_id, task_family, title, status, created_at FROM tasks").WillReturnError(errors.New("query failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/workspaces/ws_1/tasks", "", actorAdmin)
		req.SetPathValue("workspace_id", "ws_1")
		s.handleWorkspaceTasksV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("register challenge missing", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT tenant_id, operator_id, challenge_nonce, expires_at, status FROM node_registration_challenges").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/nodes/register", `{"challenge_id":"ch_1","tenant_id":"ten_01","operator_id":"op_1","public_key":"pk","runtime_type":"hermes","runtime_version":"1","signature":"sig"}`, actorAdmin)
		s.handleNodeRegisterV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("node heartbeat update fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorNode := auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01", CredentialID: "cred_1", TokenRaw: "node:ten_01:node_1:cred_1"}
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE nodes SET last_heartbeat_at = $1 WHERE node_id = $2 AND tenant_id = $3")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/nodes/node_1/heartbeat", `{}`, actorNode)
		req.SetPathValue("node_id", "node_1")
		s.handleNodeHeartbeatV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("get task not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT task_id, tenant_id, workspace_id, task_family, title, description, status, created_at FROM tasks").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/tasks/task_x", "", actorAdmin)
		req.SetPathValue("task_id", "task_x")
		s.handleGetTaskV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("get task internal error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT task_id, tenant_id, workspace_id, task_family, title, description, status, created_at FROM tasks").WillReturnError(errors.New("boom"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/tasks/task_x", "", actorAdmin)
		req.SetPathValue("task_id", "task_x")
		s.handleGetTaskV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("approve update task fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorApprover := auth.Actor{Type: "human", ID: "user_1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_1", "pending"))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5")).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/approvals/apr_1/approve", `{}`, actorApprover)
		req.SetPathValue("approval_id", "apr_1")
		s.handleApproveV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("deny update task fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorApprover := auth.Actor{Type: "human", ID: "user_1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_1", "pending"))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5")).WillReturnResult(sqlmock.NewResult(1, 1))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/approvals/apr_1/deny", `{}`, actorApprover)
		req.SetPathValue("approval_id", "apr_1")
		s.handleDenyV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("create lease insert fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorApprover := auth.Actor{Type: "human", ID: "user_1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery("SELECT status FROM tasks").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("approved"))
		mock.ExpectQuery("SELECT status FROM nodes").WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))
		mock.ExpectExec("INSERT INTO leases").WillReturnError(errors.New("insert failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases", `{"task_id":"task_1","node_id":"node_1"}`, actorApprover)
		s.handleCreateLeaseV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("claim lease exec fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorNode := auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01", CredentialID: "cred_1", TokenRaw: "node:ten_01:node_1:cred_1"}
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases/lease_1/claim", `{}`, actorNode)
		req.SetPathValue("lease_id", "lease_1")
		s.handleClaimLeaseV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("release lease exec fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorNode := auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01", CredentialID: "cred_1", TokenRaw: "node:ten_01:node_1:cred_1"}
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases/lease_1/release", `{}`, actorNode)
		req.SetPathValue("lease_id", "lease_1")
		s.handleReleaseLeaseV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("extend lease exec fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actorNode := auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01", CredentialID: "cred_1", TokenRaw: "node:ten_01:node_1:cred_1"}
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET expires_at = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WillReturnError(errors.New("update failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases/lease_1/extend", `{"extend_seconds":123}`, actorNode)
		req.SetPathValue("lease_id", "lease_1")
		s.handleExtendLeaseV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("require actor missing", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		_, _, ok := s.requireActor(rr, req)
		if ok {
			t.Fatalf("expected missing actor")
		}
	})

	t.Run("ensure tenant access platform admin", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		actorPlatform := auth.Actor{Type: "human", TenantID: "another", Roles: map[string]struct{}{"platform_admin": {}}}
		if !s.ensureTenantAccess(actorPlatform, "ten_01") {
			t.Fatalf("platform admin should bypass tenant check")
		}
	})

	t.Run("append event with nil store", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		s.events = nil
		req := reqWithActor(http.MethodGet, "/", "", actorAdmin)
		s.appendEvent(req, "ten_01", "entity", "id", "event", map[string]any{"k": "v"})
	})

	t.Run("validate node credential missing and invalid", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		rrMissing := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/", "", auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01"})
		if s.validateNodeCredential(rrMissing, req, auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01"}, "corr") {
			t.Fatalf("expected false for missing credential")
		}

		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnError(sql.ErrNoRows)
		rrInvalid := httptest.NewRecorder()
		actorNode := auth.Actor{Type: "node", ID: "node_1", TenantID: "ten_01", CredentialID: "cred_1"}
		if s.validateNodeCredential(rrInvalid, req, actorNode, "corr") {
			t.Fatalf("expected false for invalid credential")
		}
	})
}

func TestEnsureTenantAccessAndRoleValidationDirect(t *testing.T) {
	if err := validateActorRoles(auth.Actor{Type: "human", Roles: map[string]struct{}{"bad": {}}}); err == nil {
		t.Fatalf("expected role validation error")
	}
	if err := validateActorRoles(auth.Actor{Type: "node"}); err != nil {
		t.Fatalf("node actor should bypass role validation: %v", err)
	}

	actor := auth.Actor{Type: "human", TenantID: "ten_01", Roles: map[string]struct{}{}}
	s := &Server{}
	if !s.ensureTenantAccess(actor, "ten_01") {
		t.Fatalf("expected tenant access")
	}
	if s.ensureTenantAccess(actor, "ten_02") {
		t.Fatalf("did not expect cross-tenant access")
	}
}

func TestRequestContextWithActorHelper(t *testing.T) {
	actor := auth.Actor{Type: "human", TenantID: "ten_01", Roles: map[string]struct{}{"operator": {}}}
	req := reqWithActor(http.MethodGet, "/", "", actor)
	if got := telemetry.CorrelationIDFromContext(req.Context()); got != "corr_unit" {
		t.Fatalf("expected correlation id in context")
	}
	if _, ok := auth.ActorFromContext(req.Context()); !ok {
		t.Fatalf("expected actor in context")
	}
	if req.Context() == context.Background() {
		t.Fatalf("expected wrapped context")
	}
}

func TestM2TenantIsolationAndCredentialBranches(t *testing.T) {
	t.Run("create task workspace missing", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"operator": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
			WithArgs("ws_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/tasks", `{"workspace_id":"ws_missing","task_family":"research.extract","title":"t"}`, actor)
		s.handleCreateTaskV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("create task workspace lookup error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"operator": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
			WithArgs("ws_err").
			WillReturnError(errors.New("boom"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/tasks", `{"workspace_id":"ws_err","task_family":"research.extract","title":"t"}`, actor)
		s.handleCreateTaskV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("approve request not pending", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_1", "approved"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/approvals/apr_1/approve", `{}`, actor)
		req.SetPathValue("approval_id", "apr_1")
		s.handleApproveV2(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d", rr.Code)
		}
	})

	t.Run("deny request not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_1", "ten_01").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/approvals/apr_1/deny", `{}`, actor)
		req.SetPathValue("approval_id", "apr_1")
		s.handleDenyV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("deny request not pending", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_1", "denied"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/approvals/apr_1/deny", `{}`, actor)
		req.SetPathValue("approval_id", "apr_1")
		s.handleDenyV2(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d", rr.Code)
		}
	})

	t.Run("create lease task not approved", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("awaiting_approval"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases", `{"task_id":"task_1","node_id":"node_1"}`, actor)
		s.handleCreateLeaseV2(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d", rr.Code)
		}
	})

	t.Run("create lease task missing", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_1", "ten_01").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases", `{"task_id":"task_1","node_id":"node_1"}`, actor)
		s.handleCreateLeaseV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("create lease node not active", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("approved"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
			WithArgs("node_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("inactive"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases", `{"task_id":"task_1","node_id":"node_1"}`, actor)
		s.handleCreateLeaseV2(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d", rr.Code)
		}
	})

	t.Run("create lease node missing", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{Type: "human", ID: "u1", TenantID: "ten_01", Roles: map[string]struct{}{"approver": {}}}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("approved"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
			WithArgs("node_1", "ten_01").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases", `{"task_id":"task_1","node_id":"node_1"}`, actor)
		s.handleCreateLeaseV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("node credential hash mismatch", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		actor := auth.Actor{
			Type:         "node",
			ID:           "node_1",
			TenantID:     "ten_01",
			CredentialID: "cred_1",
			TokenRaw:     "node:ten_01:node_1:cred_1",
		}
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_1", "ten_01", "node_1").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", "not-a-match"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/leases/lease_1/claim", `{}`, actor)
		if s.validateNodeCredential(rr, req, actor, "corr") {
			t.Fatalf("expected false for mismatched token hash")
		}
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d", rr.Code)
		}
	})
}
