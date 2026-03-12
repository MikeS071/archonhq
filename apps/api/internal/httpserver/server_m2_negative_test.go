package httpserver

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func newHTTPServerForTest(t *testing.T) (http.Handler, sqlmock.Sqlmock, *sql.DB) {
	t.Helper()
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	return srv.Handler(), mock, dbMock
}

func TestM2NegativeBranches(t *testing.T) {
	t.Run("create tenant invalid signup", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/tenants", "human:ten_01:user_1:operator", "idem_1", map[string]any{"name": "Acme", "signup_mode": "bad"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("create tenant cross tenant forbidden", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/tenants", "human:ten_01:user_1:operator", "idem_2", map[string]any{"tenant_id": "ten_02", "name": "Acme", "signup_mode": "open"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("get tenant cross tenant forbidden", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/ten_02", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:operator")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("patch tenant forbidden role", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPatch, "/v1/tenants/ten_01", "human:ten_01:user_1:operator", "idem_3", map[string]any{"name": "x"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("patch tenant invalid signup", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		now := time.Now().UTC()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, name, signup_mode, status, created_at FROM tenants WHERE tenant_id = $1")).
			WithArgs("ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "name", "signup_mode", "status", "created_at"}).AddRow("ten_01", "Acme", "open", "active", now))

		req := newJSONRequest(t, http.MethodPatch, "/v1/tenants/ten_01", "human:ten_01:user_1:tenant_admin", "idem_4", map[string]any{"signup_mode": "bad"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("expectations: %v", err)
		}
	})

	t.Run("get members cross tenant", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := httptest.NewRequest(http.MethodGet, "/v1/tenants/ten_02/members", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:tenant_admin")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("create workspace invalid", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/workspaces", "human:ten_01:user_1:operator", "idem_5", map[string]any{"tenant_id": "ten_01"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("get workspace not found", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id, tenant_id, name, status, created_at FROM workspaces WHERE workspace_id = $1")).
			WithArgs("ws_x").
			WillReturnError(sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/v1/workspaces/ws_x", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:tenant_admin")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("register-intent runtime not allowed", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/nodes/register-intent", "human:ten_01:user_1:operator", "idem_6", map[string]any{"tenant_id": "ten_01", "operator_id": "op_1", "runtime_type": "other"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("register bad signature", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, operator_id, challenge_nonce, expires_at, status FROM node_registration_challenges WHERE challenge_id = $1")).
			WithArgs("ch_1").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "operator_id", "challenge_nonce", "expires_at", "status"}).
				AddRow("ten_01", "op_1", "challenge_ok", time.Now().UTC().Add(time.Minute), "pending"))
		req := newJSONRequest(t, http.MethodPost, "/v1/nodes/register", "human:ten_01:user_1:operator", "idem_7", map[string]any{
			"challenge_id":    "ch_1",
			"node_id":         "node_1",
			"tenant_id":       "ten_01",
			"operator_id":     "op_1",
			"public_key":      "pk",
			"runtime_type":    "hermes",
			"runtime_version": "1.0.0",
			"signature":       "bad",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("heartbeat node mismatch", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/nodes/node_2/heartbeat", "node:ten_01:node_1:cred_1", "idem_8", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("get node not found", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT node_id, tenant_id, operator_id, public_key, runtime_type, runtime_version, status, last_heartbeat_at FROM nodes WHERE node_id = $1")).
			WithArgs("node_x").
			WillReturnError(sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/v1/nodes/node_x", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:tenant_admin")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("task create invalid family", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_1:operator", "idem_9", map[string]any{"workspace_id": "ws_1", "task_family": "bad", "title": "x"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("task create cross tenant workspace forbidden", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
			WithArgs("ws_1").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_02"))
		req := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_1:operator", "idem_9b", map[string]any{
			"workspace_id": "ws_1",
			"task_family":  "research.extract",
			"title":        "x",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("task feed query error", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, workspace_id, task_family, title, status, created_at FROM tasks WHERE tenant_id = $1 ORDER BY created_at DESC LIMIT $2")).
			WithArgs("ten_01", 50).
			WillReturnError(sql.ErrConnDone)
		req := httptest.NewRequest(http.MethodGet, "/v1/tasks/feed", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:operator")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("task cancel forbidden", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_1/cancel", "human:ten_01:user_1:developer", "idem_10", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("approval queue query error", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT approval_id, task_id, status, created_at, decided_at FROM approval_requests WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC")).
			WithArgs("ten_01", "pending").
			WillReturnError(sql.ErrConnDone)
		req := httptest.NewRequest(http.MethodGet, "/v1/approvals/queue", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:approver")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("get approval not found", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT approval_id, task_id, status, requested_by, decided_by, decision_reason, created_at, decided_at FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_x", "ten_01").
			WillReturnError(sql.ErrNoRows)
		req := httptest.NewRequest(http.MethodGet, "/v1/approvals/apr_x", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_1:approver")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("approve invalid approval id", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/approvals/bad/approve", "human:ten_01:user_1:approver", "idem_11", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("deny forbidden role", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_1/deny", "human:ten_01:user_1:developer", "idem_12", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("create lease invalid payload", func(t *testing.T) {
		h, _, db := newHTTPServerForTest(t)
		defer db.Close()
		req := newJSONRequest(t, http.MethodPost, "/v1/leases", "human:ten_01:user_1:approver", "idem_13", map[string]any{"task_id": "task_1"})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("create lease task not approved", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("awaiting_approval"))
		req := newJSONRequest(t, http.MethodPost, "/v1/leases", "human:ten_01:user_1:approver", "idem_13b", map[string]any{
			"task_id": "task_1",
			"node_id": "node_1",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("extend lease invalid json", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_1", "ten_01", "node_1").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_1:cred_1")))
		req := httptest.NewRequest(http.MethodPost, "/v1/leases/lease_1/extend", strings.NewReader("{bad"))
		req.Header.Set("Authorization", "Bearer node:ten_01:node_1:cred_1")
		req.Header.Set("Idempotency-Key", "idem_14")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("claim lease revoked credential", func(t *testing.T) {
		h, mock, db := newHTTPServerForTest(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_2", "ten_01", "node_1").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("revoked", hashString("node:ten_01:node_1:cred_2")))
		req := newJSONRequest(t, http.MethodPost, "/v1/leases/lease_1/claim", "node:ten_01:node_1:cred_2", "idem_15", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusUnauthorized {
			t.Fatalf("expected 401 got %d", rr.Code)
		}
	})
}

func TestHelperFunctionsDirectCoverage(t *testing.T) {
	if got := approvalIDFromTaskID("other"); !strings.HasPrefix(got, "apr_") {
		t.Fatalf("expected generated approval id, got %s", got)
	}
	if got := taskIDFromApprovalID("other"); got != "" {
		t.Fatalf("expected empty task id for invalid approval id")
	}
}
