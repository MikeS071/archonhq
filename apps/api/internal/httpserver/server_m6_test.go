package httpserver

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestPaperclipSyncAndStatusEndpointsM6(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  w.workspace_id,")).WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "name", "tasks_total", "pending_approvals"}).
			AddRow("ws_01", "Main Workspace", 2, 1))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  aq.approval_id,")).WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"approval_id", "task_id", "status", "created_at", "workspace_id", "title", "task_status"}).
			AddRow("apr_01", "task_01", "pending", now, "ws_01", "Review claim", "awaiting_approval"))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT node_id, operator_id, runtime_type, runtime_version, status, last_heartbeat_at")).WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "operator_id", "runtime_type", "runtime_version", "status", "last_heartbeat_at"}).
			AddRow("node_01", "op_01", "hermes", "1.9", "healthy", now))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT subject_type, subject_id, family, window_name, rf_value, created_at")).WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"subject_type", "subject_id", "family", "window_name", "rf_value", "created_at"}).
			AddRow("operator", "op_01", "fleet", "last_30d", 0.96, now))

	mock.ExpectQuery(regexp.QuoteMeta("SELECT entry_id, account_id, result_id, credited_jw::text, net_amount::text, status, created_at")).WithArgs("ten_01", 10).
		WillReturnRows(sqlmock.NewRows([]string{"entry_id", "account_id", "result_id", "credited_jw", "net_amount", "status", "created_at"}).
			AddRow("entry_01", "acct_01", "res_01", "1.50000000", "2.25000000", "posted", now))

	syncReq := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human:ten_01:user_admin:tenant_admin,operator", "idem_paperclip_sync_1", map[string]any{
		"sync_id": "pcsync_01",
		"limit":   10,
	})
	rrSync := httptest.NewRecorder()
	h.ServeHTTP(rrSync, syncReq)
	if rrSync.Code != http.StatusAccepted {
		t.Fatalf("paperclip sync expected 202 got %d body=%s", rrSync.Code, rrSync.Body.String())
	}
	if !strings.Contains(rrSync.Body.String(), `"status":"completed"`) {
		t.Fatalf("paperclip sync expected completed status body=%s", rrSync.Body.String())
	}
	if !strings.Contains(rrSync.Body.String(), `"source_of_truth":"postgres"`) {
		t.Fatalf("paperclip sync expected postgres source body=%s", rrSync.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id, event_type, payload_json, occurred_at")).WithArgs("ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"event_id", "event_type", "payload_json", "occurred_at"}).
			AddRow("evt_01", "integration.paperclip_sync_completed", []byte(`{"sync_id":"pcsync_01","status":"completed","external_ref":"paperclip_stub_20260101T000000Z","surface_counts":{"workspace_summaries":1}}`), now))

	statusReq := httptest.NewRequest(http.MethodGet, "/v1/integrations/paperclip/status", nil)
	statusReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator")
	rrStatus := httptest.NewRecorder()
	h.ServeHTTP(rrStatus, statusReq)
	if rrStatus.Code != http.StatusOK {
		t.Fatalf("paperclip status expected 200 got %d body=%s", rrStatus.Code, rrStatus.Body.String())
	}
	if !strings.Contains(rrStatus.Body.String(), `"status":"completed"`) {
		t.Fatalf("paperclip status expected completed body=%s", rrStatus.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPaperclipStatusNeverSyncedM6(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id, event_type, payload_json, occurred_at")).WithArgs("ten_01").
		WillReturnError(sqlmock.ErrCancelled)

	req := httptest.NewRequest(http.MethodGet, "/v1/integrations/paperclip/status", nil)
	req.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator")
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusInternalServerError {
		t.Fatalf("paperclip status db error expected 500 got %d body=%s", rr.Code, rr.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT event_id, event_type, payload_json, occurred_at")).WithArgs("ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"event_id", "event_type", "payload_json", "occurred_at"}))

	rr2 := httptest.NewRecorder()
	h.ServeHTTP(rr2, req)
	if rr2.Code != http.StatusOK {
		t.Fatalf("paperclip status never synced expected 200 got %d body=%s", rr2.Code, rr2.Body.String())
	}
	if !strings.Contains(rr2.Body.String(), `"status":"never_synced"`) {
		t.Fatalf("paperclip status expected never_synced body=%s", rr2.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPaperclipSyncValidationAndFailuresM6(t *testing.T) {
	t.Run("forbidden role", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()

		req := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human:ten_01:user_dev:developer", "idem_paperclip_sync_forbidden", map[string]any{
			"sync_id": "pcsync_forbidden",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("invalid json payload", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()

		req := httptest.NewRequest(http.MethodPost, "/v1/integrations/paperclip/sync", strings.NewReader("{"))
		req.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator")
		req.Header.Set("Idempotency-Key", "idem_paperclip_sync_bad_json")
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("invalid limit", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()

		req := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human:ten_01:user_admin:tenant_admin,operator", "idem_paperclip_sync_bad_limit", map[string]any{
			"limit": 501,
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("projection query failure returns integration_sync_failed", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  w.workspace_id,")).WithArgs("ten_01", 50).WillReturnError(errors.New("db unavailable"))

		req := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human:ten_01:user_admin:tenant_admin,operator", "idem_paperclip_sync_qerr", map[string]any{
			"dry_run": true,
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), `"code":"integration_sync_failed"`) {
			t.Fatalf("expected integration_sync_failed body=%s", rr.Body.String())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("connector guardrail failure returns bad gateway", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  w.workspace_id,")).WithArgs("", 50).
			WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "name", "tasks_total", "pending_approvals"}))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  aq.approval_id,")).WithArgs("", 50).
			WillReturnRows(sqlmock.NewRows([]string{"approval_id", "task_id", "status", "created_at", "workspace_id", "title", "task_status"}))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT node_id, operator_id, runtime_type, runtime_version, status, last_heartbeat_at")).WithArgs("", 50).
			WillReturnRows(sqlmock.NewRows([]string{"node_id", "operator_id", "runtime_type", "runtime_version", "status", "last_heartbeat_at"}))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT subject_type, subject_id, family, window_name, rf_value, created_at")).WithArgs("", 50).
			WillReturnRows(sqlmock.NewRows([]string{"subject_type", "subject_id", "family", "window_name", "rf_value", "created_at"}))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT entry_id, account_id, result_id, credited_jw::text, net_amount::text, status, created_at")).WithArgs("", 50).
			WillReturnRows(sqlmock.NewRows([]string{"entry_id", "account_id", "result_id", "credited_jw", "net_amount", "status", "created_at"}))

		req := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human::user_admin:tenant_admin,operator", "idem_paperclip_sync_connector_err", map[string]any{
			"sync_id": "pcsync_connector_fail",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadGateway {
			t.Fatalf("expected 502 got %d body=%s", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), `"code":"integration_sync_failed"`) {
			t.Fatalf("expected integration_sync_failed body=%s", rr.Body.String())
		}

		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})
}

func TestPaperclipSyncDryRunDefaultLimitM6(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  w.workspace_id,")).WithArgs("ten_01", 50).
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id", "name", "tasks_total", "pending_approvals"}).
			AddRow("ws_01", "Main Workspace", 2, 1))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  aq.approval_id,")).WithArgs("ten_01", 50).
		WillReturnRows(sqlmock.NewRows([]string{"approval_id", "task_id", "status", "created_at", "workspace_id", "title", "task_status"}).
			AddRow("apr_01", "task_01", "pending", now, "ws_01", "Review claim", "awaiting_approval"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT node_id, operator_id, runtime_type, runtime_version, status, last_heartbeat_at")).WithArgs("ten_01", 50).
		WillReturnRows(sqlmock.NewRows([]string{"node_id", "operator_id", "runtime_type", "runtime_version", "status", "last_heartbeat_at"}).
			AddRow("node_01", "op_01", "hermes", "1.9", "healthy", now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT subject_type, subject_id, family, window_name, rf_value, created_at")).WithArgs("ten_01", 50).
		WillReturnRows(sqlmock.NewRows([]string{"subject_type", "subject_id", "family", "window_name", "rf_value", "created_at"}).
			AddRow("operator", "op_01", "fleet", "last_30d", 0.96, now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT entry_id, account_id, result_id, credited_jw::text, net_amount::text, status, created_at")).WithArgs("ten_01", 50).
		WillReturnRows(sqlmock.NewRows([]string{"entry_id", "account_id", "result_id", "credited_jw", "net_amount", "status", "created_at"}).
			AddRow("entry_01", "acct_01", "res_01", "1.50000000", "2.25000000", "posted", now))

	req := newJSONRequest(t, http.MethodPost, "/v1/integrations/paperclip/sync", "human:ten_01:user_admin:tenant_admin,operator", "idem_paperclip_sync_dry_run", map[string]any{
		"dry_run": true,
	})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusAccepted {
		t.Fatalf("expected 202 got %d body=%s", rr.Code, rr.Body.String())
	}

	var body map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}

	syncID, _ := body["sync_id"].(string)
	if !strings.HasPrefix(syncID, "pcsync_") {
		t.Fatalf("expected generated sync_id prefix pcsync_ got %q", syncID)
	}
	if status, _ := body["status"].(string); status != "dry_run" {
		t.Fatalf("expected dry_run status got %v", body["status"])
	}
	if source, _ := body["source_of_truth"].(string); source != "postgres" {
		t.Fatalf("expected postgres source_of_truth got %v", body["source_of_truth"])
	}
	if authoritative, _ := body["paperclip_authoritative"].(bool); authoritative {
		t.Fatalf("paperclip_authoritative must be false")
	}
	externalRef, _ := body["external_ref"].(string)
	if !strings.HasPrefix(externalRef, "paperclip://dry-run/"+syncID) {
		t.Fatalf("expected dry run external_ref prefix for sync_id %q got %q", syncID, externalRef)
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestPaperclipStatusFromEventM6(t *testing.T) {
	tests := []struct {
		name      string
		eventType string
		payload   map[string]any
		want      string
	}{
		{
			name:      "completed uses payload status when present",
			eventType: "integration.paperclip_sync_completed",
			payload:   map[string]any{"status": "completed"},
			want:      "completed",
		},
		{
			name:      "completed defaults when status missing",
			eventType: "integration.paperclip_sync_completed",
			payload:   map[string]any{},
			want:      "completed",
		},
		{
			name:      "failed event",
			eventType: "integration.paperclip_sync_failed",
			payload:   map[string]any{"error": "boom"},
			want:      "failed",
		},
		{
			name:      "requested event",
			eventType: "integration.paperclip_sync_requested",
			payload:   map[string]any{},
			want:      "requested",
		},
		{
			name:      "unknown event",
			eventType: "integration.paperclip_sync_other",
			payload:   map[string]any{},
			want:      "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := paperclipStatusFromEvent(tt.eventType, tt.payload); got != tt.want {
				t.Fatalf("paperclipStatusFromEvent(%q)=%q want %q", tt.eventType, got, tt.want)
			}
		})
	}
}

func TestCollectPaperclipProjectionPayloadErrorsM6(t *testing.T) {
	cases := []struct {
		name        string
		failingStep int
	}{
		{name: "workspace query fails", failingStep: 0},
		{name: "approvals query fails", failingStep: 1},
		{name: "fleet query fails", failingStep: 2},
		{name: "reliability query fails", failingStep: 3},
		{name: "settlements query fails", failingStep: 4},
	}

	type querySpec struct {
		pattern string
		cols    []string
	}

	specs := []querySpec{
		{
			pattern: "SELECT\n  w.workspace_id,",
			cols:    []string{"workspace_id", "name", "tasks_total", "pending_approvals"},
		},
		{
			pattern: "SELECT\n  aq.approval_id,",
			cols:    []string{"approval_id", "task_id", "status", "created_at", "workspace_id", "title", "task_status"},
		},
		{
			pattern: "SELECT node_id, operator_id, runtime_type, runtime_version, status, last_heartbeat_at",
			cols:    []string{"node_id", "operator_id", "runtime_type", "runtime_version", "status", "last_heartbeat_at"},
		},
		{
			pattern: "SELECT subject_type, subject_id, family, window_name, rf_value, created_at",
			cols:    []string{"subject_type", "subject_id", "family", "window_name", "rf_value", "created_at"},
		},
		{
			pattern: "SELECT entry_id, account_id, result_id, credited_jw::text, net_amount::text, status, created_at",
			cols:    []string{"entry_id", "account_id", "result_id", "credited_jw", "net_amount", "status", "created_at"},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			dbMock, mock, err := sqlmock.New()
			if err != nil {
				t.Fatalf("sqlmock new: %v", err)
			}
			defer dbMock.Close()

			srv := newTestServer(t, dbMock, &inMemoryEventStore{})
			req := httptest.NewRequest(http.MethodGet, "/unused", nil)

			for i, spec := range specs {
				expect := mock.ExpectQuery(regexp.QuoteMeta(spec.pattern)).WithArgs("ten_01", 50)
				if i == tc.failingStep {
					expect.WillReturnError(errors.New("query failed"))
					break
				}
				expect.WillReturnRows(sqlmock.NewRows(spec.cols))
			}

			_, err = srv.collectPaperclipProjectionPayload(req, "ten_01", 50)
			if err == nil {
				t.Fatalf("expected collectPaperclipProjectionPayload to fail at step %d", tc.failingStep)
			}

			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("sql expectations: %v", err)
			}
		})
	}
}
