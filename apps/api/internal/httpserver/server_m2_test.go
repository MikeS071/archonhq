package httpserver

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/MikeS071/archonhq/pkg/db"
	"github.com/MikeS071/archonhq/pkg/events"
	natsclient "github.com/MikeS071/archonhq/pkg/nats"
	"github.com/MikeS071/archonhq/pkg/objectstore"
	redisclient "github.com/MikeS071/archonhq/pkg/redis"
)

type inMemoryEventStore struct {
	events []events.Envelope
}

func (s *inMemoryEventStore) Append(_ context.Context, env events.Envelope) error {
	s.events = append(s.events, env)
	return nil
}

func newTestServer(t *testing.T, sqlDB *sql.DB, eventStore events.Store) *Server {
	t.Helper()
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return New(
		logger,
		&db.Postgres{DB: sqlDB},
		&natsclient.Client{},
		&redisclient.Client{URL: "redis://test"},
		&objectstore.Client{Config: objectstore.Config{Endpoint: "http://object", Bucket: "bucket"}},
		eventStore,
	)
}

func newJSONRequest(t *testing.T, method, path, token, idem string, body any) *http.Request {
	t.Helper()
	payload, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Idempotency-Key", idem)
	return req
}

func TestTenantsAndWorkspacesFlow(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	eventStore := &inMemoryEventStore{}
	srv := newTestServer(t, dbMock, eventStore)
	h := srv.Handler()

	now := time.Now().UTC()

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tenants")).WithArgs("ten_01", "Acme Corp", "approval", "active").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO memberships")).WithArgs(sqlmock.AnyArg(), "ten_01", "user_admin", "tenant_admin").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, name, signup_mode, status, created_at FROM tenants WHERE tenant_id = $1")).WithArgs("ten_01").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "name", "signup_mode", "status", "created_at"}).
		AddRow("ten_01", "Acme Corp", "approval", "active", now))

	createTenant := newJSONRequest(t, http.MethodPost, "/v1/tenants", "human:ten_01:user_admin:tenant_admin,operator", "idem_ten_1", map[string]any{
		"tenant_id":   "ten_01",
		"name":        "Acme Corp",
		"signup_mode": "approval",
	})
	rrCreateTenant := httptest.NewRecorder()
	h.ServeHTTP(rrCreateTenant, createTenant)
	if rrCreateTenant.Code != http.StatusOK {
		t.Fatalf("create tenant expected 200 got %d body=%s", rrCreateTenant.Code, rrCreateTenant.Body.String())
	}

	getTenant := httptest.NewRequest(http.MethodGet, "/v1/tenants/ten_01", nil)
	getTenant.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator")
	rrGetTenant := httptest.NewRecorder()
	h.ServeHTTP(rrGetTenant, getTenant)
	if rrGetTenant.Code != http.StatusOK {
		t.Fatalf("get tenant expected 200 got %d body=%s", rrGetTenant.Code, rrGetTenant.Body.String())
	}

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO workspaces")).WithArgs("ws_01", "ten_01", "Main Workspace", "active").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id, tenant_id, name, status, created_at FROM workspaces WHERE workspace_id = $1 AND tenant_id = $2")).WithArgs("ws_01", "ten_01").WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "tenant_id", "name", "status", "created_at"}).
		AddRow("ws_01", "ten_01", "Main Workspace", "active", now))

	createWorkspace := newJSONRequest(t, http.MethodPost, "/v1/workspaces", "human:ten_01:user_admin:tenant_admin,operator", "idem_ws_1", map[string]any{
		"workspace_id": "ws_01",
		"tenant_id":    "ten_01",
		"name":         "Main Workspace",
	})
	rrCreateWS := httptest.NewRecorder()
	h.ServeHTTP(rrCreateWS, createWorkspace)
	if rrCreateWS.Code != http.StatusOK {
		t.Fatalf("create workspace expected 200 got %d body=%s", rrCreateWS.Code, rrCreateWS.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
	if len(eventStore.events) == 0 {
		t.Fatalf("expected events to be written")
	}
}

func TestNodeRegistrationAndHeartbeatFlow(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	challenge := "challenge_abc"

	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO node_registration_challenges")).WithArgs("ch_01", "ten_01", "op_01", "hermes", sqlmock.AnyArg(), sqlmock.AnyArg(), "pending").WillReturnResult(sqlmock.NewResult(1, 1))

	registerIntent := newJSONRequest(t, http.MethodPost, "/v1/nodes/register-intent", "human:ten_01:user_admin:tenant_admin,operator", "idem_node_intent", map[string]any{
		"challenge_id": "ch_01",
		"tenant_id":    "ten_01",
		"operator_id":  "op_01",
		"runtime_type": "hermes",
	})
	rrIntent := httptest.NewRecorder()
	h.ServeHTTP(rrIntent, registerIntent)
	if rrIntent.Code != http.StatusOK {
		t.Fatalf("register-intent expected 200 got %d body=%s", rrIntent.Code, rrIntent.Body.String())
	}
	var intentResp map[string]any
	if err := json.Unmarshal(rrIntent.Body.Bytes(), &intentResp); err != nil {
		t.Fatalf("parse register-intent response: %v", err)
	}
	challenge, _ = intentResp["challenge"].(string)
	if challenge == "" {
		t.Fatalf("expected challenge in register-intent response")
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, operator_id, challenge_nonce, expires_at, status FROM node_registration_challenges WHERE challenge_id = $1")).WithArgs("ch_01").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "operator_id", "challenge_nonce", "expires_at", "status"}).
		AddRow("ten_01", "op_01", challenge, time.Now().UTC().Add(10*time.Minute), "pending"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE node_registration_challenges SET status = $1 WHERE challenge_id = $2")).WithArgs("used", "ch_01").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO nodes")).WithArgs("node_01", "ten_01", "op_01", "pub_key", "hermes", "1.0.0", "active").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO node_credentials")).WithArgs(sqlmock.AnyArg(), "ten_01", "node_01", sqlmock.AnyArg(), "active").WillReturnResult(sqlmock.NewResult(1, 1))

	register := newJSONRequest(t, http.MethodPost, "/v1/nodes/register", "human:ten_01:user_admin:tenant_admin,operator", "idem_node_reg", map[string]any{
		"challenge_id":    "ch_01",
		"node_id":         "node_01",
		"tenant_id":       "ten_01",
		"operator_id":     "op_01",
		"public_key":      "pub_key",
		"runtime_type":    "hermes",
		"runtime_version": "1.0.0",
		"signature":       "signed:" + challenge,
	})
	rrRegister := httptest.NewRecorder()
	h.ServeHTTP(rrRegister, register)
	if rrRegister.Code != http.StatusOK {
		t.Fatalf("register expected 200 got %d body=%s", rrRegister.Code, rrRegister.Body.String())
	}

	var registerResp map[string]any
	if err := json.Unmarshal(rrRegister.Body.Bytes(), &registerResp); err != nil {
		t.Fatalf("parse register response: %v", err)
	}
	nodeToken, _ := registerResp["node_token"].(string)
	if nodeToken == "" {
		t.Fatalf("expected node_token in response")
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs(sqlmock.AnyArg(), "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(nodeToken)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE nodes SET last_heartbeat_at = $1 WHERE node_id = $2 AND tenant_id = $3")).WithArgs(sqlmock.AnyArg(), "node_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))

	heartbeat := newJSONRequest(t, http.MethodPost, "/v1/nodes/node_01/heartbeat", nodeToken, "idem_hb_1", map[string]any{})
	rrHeartbeat := httptest.NewRecorder()
	h.ServeHTTP(rrHeartbeat, heartbeat)
	if rrHeartbeat.Code != http.StatusOK {
		t.Fatalf("heartbeat expected 200 got %d body=%s", rrHeartbeat.Code, rrHeartbeat.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestTaskApprovalLeaseLifecycle(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).WithArgs("ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_01"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO tasks")).WithArgs("task_01", "ten_01", "ws_01", "research.extract", "Collect signals", "", "awaiting_approval", sqlmock.AnyArg(), "always_required", "user_admin").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO approval_requests")).WithArgs("apr_01", "ten_01", "task_01", "pending", "user_admin").WillReturnResult(sqlmock.NewResult(1, 1))

	createTask := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_admin:tenant_admin,operator,approver", "idem_task_1", map[string]any{
		"task_id":       "task_01",
		"workspace_id":  "ws_01",
		"task_family":   "research.extract",
		"title":         "Collect signals",
		"description":   "",
		"approval_mode": "always_required",
	})
	rrTask := httptest.NewRecorder()
	h.ServeHTTP(rrTask, createTask)
	if rrTask.Code != http.StatusOK {
		t.Fatalf("create task expected 200 got %d body=%s", rrTask.Code, rrTask.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).WithArgs("apr_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "status"}).AddRow("task_01", "pending"))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE approval_requests SET status = $1, decided_by = $2, decided_at = $3 WHERE approval_id = $4 AND tenant_id = $5")).WithArgs("approved", "user_admin", sqlmock.AnyArg(), "apr_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE tasks SET status = $1 WHERE task_id = $2 AND tenant_id = $3")).WithArgs("approved", "task_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))

	approve := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_01/approve", "human:ten_01:user_admin:tenant_admin,operator,approver", "idem_apr_1", map[string]any{})
	rrApprove := httptest.NewRecorder()
	h.ServeHTTP(rrApprove, approve)
	if rrApprove.Code != http.StatusOK {
		t.Fatalf("approve expected 200 got %d body=%s", rrApprove.Code, rrApprove.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM tasks WHERE task_id = $1 AND tenant_id = $2")).WithArgs("task_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("approved"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status FROM nodes WHERE node_id = $1 AND tenant_id = $2")).WithArgs("node_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("active"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO leases")).WithArgs("lease_01", "ten_01", "task_01", "node_01", 1, "granted", "approved", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(1, 1))

	createLease := newJSONRequest(t, http.MethodPost, "/v1/leases", "human:ten_01:user_admin:tenant_admin,operator,approver", "idem_lease_1", map[string]any{
		"lease_id":           "lease_01",
		"task_id":            "task_01",
		"node_id":            "node_01",
		"expires_in_seconds": 600,
		"execution_policy":   map[string]any{"allowed_backends": []string{"docker"}},
	})
	rrCreateLease := httptest.NewRecorder()
	h.ServeHTTP(rrCreateLease, createLease)
	if rrCreateLease.Code != http.StatusOK {
		t.Fatalf("create lease expected 200 got %d body=%s", rrCreateLease.Code, rrCreateLease.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WithArgs("claimed", "lease_01", "node_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))

	claim := newJSONRequest(t, http.MethodPost, "/v1/leases/lease_01/claim", "node:ten_01:node_01:cred_01", "idem_claim_1", map[string]any{})
	rrClaim := httptest.NewRecorder()
	h.ServeHTTP(rrClaim, claim)
	if rrClaim.Code != http.StatusOK {
		t.Fatalf("claim expected 200 got %d body=%s", rrClaim.Code, rrClaim.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET status = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WithArgs("released", "lease_01", "node_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))

	release := newJSONRequest(t, http.MethodPost, "/v1/leases/lease_01/release", "node:ten_01:node_01:cred_01", "idem_release_1", map[string]any{})
	rrRelease := httptest.NewRecorder()
	h.ServeHTTP(rrRelease, release)
	if rrRelease.Code != http.StatusOK {
		t.Fatalf("release expected 200 got %d body=%s", rrRelease.Code, rrRelease.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE leases SET expires_at = $1 WHERE lease_id = $2 AND node_id = $3 AND tenant_id = $4")).WithArgs(sqlmock.AnyArg(), "lease_01", "node_01", "ten_01").WillReturnResult(sqlmock.NewResult(1, 1))

	extend := newJSONRequest(t, http.MethodPost, "/v1/leases/lease_01/extend", "node:ten_01:node_01:cred_01", "idem_extend_1", map[string]any{
		"extend_seconds": 300,
	})
	rrExtend := httptest.NewRecorder()
	h.ServeHTTP(rrExtend, extend)
	if rrExtend.Code != http.StatusOK {
		t.Fatalf("extend expected 200 got %d body=%s", rrExtend.Code, rrExtend.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT approval_id, task_id, status, created_at, decided_at FROM approval_requests WHERE tenant_id = $1 AND status = $2 ORDER BY created_at DESC")).WithArgs("ten_01", "pending").WillReturnRows(sqlmock.NewRows([]string{"approval_id", "task_id", "status", "created_at", "decided_at"}))

	queueReq := httptest.NewRequest(http.MethodGet, "/v1/approvals/queue", nil)
	queueReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator,approver")
	rrQueue := httptest.NewRecorder()
	h.ServeHTTP(rrQueue, queueReq)
	if rrQueue.Code != http.StatusOK {
		t.Fatalf("queue expected 200 got %d body=%s", rrQueue.Code, rrQueue.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, description, status, created_at FROM tasks WHERE task_id = $1 AND tenant_id = $2")).WithArgs("task_01", "ten_01").WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "description", "status", "created_at"}).
		AddRow("task_01", "ten_01", "ws_01", "research.extract", "Collect signals", "", "approved", now))

	getTask := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_01", nil)
	getTask.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator,approver")
	rrGetTask := httptest.NewRecorder()
	h.ServeHTTP(rrGetTask, getTask)
	if rrGetTask.Code != http.StatusOK {
		t.Fatalf("get task expected 200 got %d body=%s", rrGetTask.Code, rrGetTask.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
