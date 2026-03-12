package httpserver

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestAdditionalM2EndpointsCoverage(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	eventStore := &inMemoryEventStore{}
	srv := newTestServer(t, dbMock, eventStore)
	h := srv.Handler()

	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, name, signup_mode, status, created_at FROM tenants WHERE tenant_id = $1")).
		WithArgs("ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "name", "signup_mode", "status", "created_at"}).
			AddRow("ten_01", "Acme", "open", "active", now))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE tenants SET name = $1, signup_mode = $2, status = $3 WHERE tenant_id = $4")).
		WithArgs("Acme Updated", "mixed", "active", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))

	patchTenant := newJSONRequest(t, http.MethodPatch, "/v1/tenants/ten_01", "human:ten_01:user_admin:tenant_admin,approver", "idem_patch_tenant", map[string]any{
		"name":        "Acme Updated",
		"signup_mode": "mixed",
		"status":      "active",
	})
	rrPatchTenant := httptest.NewRecorder()
	h.ServeHTTP(rrPatchTenant, patchTenant)
	if rrPatchTenant.Code != http.StatusOK {
		t.Fatalf("patch tenant expected 200 got %d body=%s", rrPatchTenant.Code, rrPatchTenant.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT membership_id, user_id, role, created_at FROM memberships WHERE tenant_id = $1 ORDER BY created_at DESC")).
		WithArgs("ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"membership_id", "user_id", "role", "created_at"}).
			AddRow("mem_01", "user_admin", "tenant_admin", now))
	getMembers := httptest.NewRequest(http.MethodGet, "/v1/tenants/ten_01/members", nil)
	getMembers.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrMembers := httptest.NewRecorder()
	h.ServeHTTP(rrMembers, getMembers)
	if rrMembers.Code != http.StatusOK {
		t.Fatalf("get members expected 200 got %d body=%s", rrMembers.Code, rrMembers.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id, tenant_id, name, status, created_at FROM workspaces WHERE workspace_id = $1")).
		WithArgs("ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "tenant_id", "name", "status", "created_at"}).
			AddRow("ws_01", "ten_01", "Main", "active", now))
	getWS := httptest.NewRequest(http.MethodGet, "/v1/workspaces/ws_01", nil)
	getWS.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrWS := httptest.NewRecorder()
	h.ServeHTTP(rrWS, getWS)
	if rrWS.Code != http.StatusOK {
		t.Fatalf("get workspace expected 200 got %d body=%s", rrWS.Code, rrWS.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM tasks WHERE tenant_id = $1 AND workspace_id = $2")).
		WithArgs("ten_01", "ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(5))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM approval_requests WHERE tenant_id = $1 AND status = 'pending'")).
		WithArgs("ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	wsSummary := httptest.NewRequest(http.MethodGet, "/v1/workspaces/ws_01/summary", nil)
	wsSummary.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrSummary := httptest.NewRecorder()
	h.ServeHTTP(rrSummary, wsSummary)
	if rrSummary.Code != http.StatusOK {
		t.Fatalf("workspace summary expected 200 got %d body=%s", rrSummary.Code, rrSummary.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family, title, status, created_at FROM tasks WHERE tenant_id = $1 AND workspace_id = $2 ORDER BY created_at DESC")).
		WithArgs("ten_01", "ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family", "title", "status", "created_at"}).
			AddRow("task_01", "research.extract", "Task One", "approved", now))
	wsTasks := httptest.NewRequest(http.MethodGet, "/v1/workspaces/ws_01/tasks", nil)
	wsTasks.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrWSTasks := httptest.NewRecorder()
	h.ServeHTTP(rrWSTasks, wsTasks)
	if rrWSTasks.Code != http.StatusOK {
		t.Fatalf("workspace tasks expected 200 got %d body=%s", rrWSTasks.Code, rrWSTasks.Body.String())
	}

	wsLedger := httptest.NewRequest(http.MethodGet, "/v1/workspaces/ws_01/ledger", nil)
	wsLedger.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrLedger := httptest.NewRecorder()
	h.ServeHTTP(rrLedger, wsLedger)
	if rrLedger.Code != http.StatusOK {
		t.Fatalf("workspace ledger expected 200 got %d body=%s", rrLedger.Code, rrLedger.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT node_id, tenant_id, operator_id, public_key, runtime_type, runtime_version, status, last_heartbeat_at FROM nodes WHERE node_id = $1")).
		WithArgs("node_01").
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "tenant_id", "operator_id", "public_key", "runtime_type", "runtime_version", "status", "last_heartbeat_at"}).
			AddRow("node_01", "ten_01", "op_01", "pub", "hermes", "1.0.0", "active", now))
	getNode := httptest.NewRequest(http.MethodGet, "/v1/nodes/node_01", nil)
	getNode.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrNode := httptest.NewRecorder()
	h.ServeHTTP(rrNode, getNode)
	if rrNode.Code != http.StatusOK {
		t.Fatalf("get node expected 200 got %d body=%s", rrNode.Code, rrNode.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT lease_id, task_id, node_id, status, approval_state, granted_at, expires_at FROM leases WHERE node_id = $1 AND tenant_id = $2 ORDER BY granted_at DESC")).
		WithArgs("node_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"lease_id", "task_id", "node_id", "status", "approval_state", "granted_at", "expires_at"}).
			AddRow("lease_01", "task_01", "node_01", "claimed", "approved", now, now.Add(5*time.Minute)))
	getNodeLeases := httptest.NewRequest(http.MethodGet, "/v1/nodes/node_01/leases", nil)
	getNodeLeases.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrNodeLeases := httptest.NewRecorder()
	h.ServeHTTP(rrNodeLeases, getNodeLeases)
	if rrNodeLeases.Code != http.StatusOK {
		t.Fatalf("get node leases expected 200 got %d body=%s", rrNodeLeases.Code, rrNodeLeases.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, workspace_id, task_family, title, status, created_at FROM tasks WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2")).
		WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "workspace_id", "task_family", "title", "status", "created_at"}).
			AddRow("task_01", "ws_01", "research.extract", "Task One", "approved", now))
	feedReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/feed?limit=10", nil)
	feedReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrFeed := httptest.NewRecorder()
	h.ServeHTTP(rrFeed, feedReq)
	if rrFeed.Code != http.StatusOK {
		t.Fatalf("task feed expected 200 got %d body=%s", rrFeed.Code, rrFeed.Body.String())
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3")).
		WithArgs("cancelled", "task_01", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))
	cancelReq := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_01/cancel", "human:ten_01:user_admin:tenant_admin,approver", "idem_cancel", map[string]any{})
	rrCancel := httptest.NewRecorder()
	h.ServeHTTP(rrCancel, cancelReq)
	if rrCancel.Code != http.StatusOK {
		t.Fatalf("task cancel expected 200 got %d body=%s", rrCancel.Code, rrCancel.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT approval_id, task_id, status, requested_by, decided_by, decision_reason, created_at, decided_at FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
		WithArgs("apr_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"approval_id", "task_id", "status", "requested_by", "decided_by", "decision_reason", "created_at", "decided_at"}).
			AddRow("apr_01", "task_01", "pending", "user_admin", "", "", now, nil))
	approvalReq := httptest.NewRequest(http.MethodGet, "/v1/approvals/apr_01", nil)
	approvalReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,approver")
	rrApproval := httptest.NewRecorder()
	h.ServeHTTP(rrApproval, approvalReq)
	if rrApproval.Code != http.StatusOK {
		t.Fatalf("get approval expected 200 got %d body=%s", rrApproval.Code, rrApproval.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
		WithArgs("apr_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_01", "pending"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5")).
		WithArgs("denied", "user_admin", sqlmock.AnyArg(), "apr_01", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3")).
		WithArgs("denied", "task_01", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))

	denyReq := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_01/deny", "human:ten_01:user_admin:tenant_admin,approver", "idem_deny", map[string]any{})
	rrDeny := httptest.NewRecorder()
	h.ServeHTTP(rrDeny, denyReq)
	if rrDeny.Code != http.StatusOK {
		t.Fatalf("deny approval expected 200 got %d body=%s", rrDeny.Code, rrDeny.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestValidationBranchesCoverage(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	badRoleReq := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_admin:invalid_role", "idem_bad_role", map[string]any{
		"workspace_id": "ws_01",
		"task_family":  "research.extract",
		"title":        "Bad role",
	})
	rrBadRole := httptest.NewRecorder()
	h.ServeHTTP(rrBadRole, badRoleReq)
	if rrBadRole.Code != http.StatusForbidden {
		t.Fatalf("invalid role expected 403 got %d body=%s", rrBadRole.Code, rrBadRole.Body.String())
	}

	noCredReq := newJSONRequest(t, http.MethodPost, "/v1/leases/lease_01/claim", "node:ten_01:node_01", "idem_no_cred", map[string]any{})
	rrNoCred := httptest.NewRecorder()
	h.ServeHTTP(rrNoCred, noCredReq)
	if rrNoCred.Code != http.StatusUnauthorized {
		t.Fatalf("missing credential expected 401 got %d body=%s", rrNoCred.Code, rrNoCred.Body.String())
	}
}
