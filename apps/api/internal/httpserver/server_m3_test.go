package httpserver

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestArtifactEndpointsM3(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
		WithArgs("ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_01"))
	uploadReq := newJSONRequest(t, http.MethodPost, "/v1/artifacts/upload-url", "node:ten_01:node_01:cred_01", "idem_art_upload", map[string]any{
		"workspace_id": "ws_01",
		"file_name":    "result.json",
		"media_type":   "application/json",
		"size_bytes":   123,
	})
	rrUpload := httptest.NewRecorder()
	h.ServeHTTP(rrUpload, uploadReq)
	if rrUpload.Code != http.StatusOK {
		t.Fatalf("upload-url expected 200 got %d body=%s", rrUpload.Code, rrUpload.Body.String())
	}

	var uploadResp map[string]any
	if err := json.Unmarshal(rrUpload.Body.Bytes(), &uploadResp); err != nil {
		t.Fatalf("parse upload response: %v", err)
	}
	artifactID, _ := uploadResp["artifact_id"].(string)
	blobRef, _ := uploadResp["blob_ref"].(string)
	if artifactID == "" || blobRef == "" {
		t.Fatalf("expected artifact_id/blob_ref in upload response")
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
		WithArgs("ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_01"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO artifacts")).
		WithArgs(artifactID, "ten_01", "ws_01", blobRef, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "application/json", int64(123), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	registerReq := newJSONRequest(t, http.MethodPost, "/v1/artifacts/register", "node:ten_01:node_01:cred_01", "idem_art_register", map[string]any{
		"artifact_id":  artifactID,
		"workspace_id": "ws_01",
		"blob_ref":     blobRef,
		"sha256":       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"media_type":   "application/json",
		"size_bytes":   123,
		"metadata":     map[string]any{"kind": "result"},
	})
	rrRegister := httptest.NewRecorder()
	h.ServeHTTP(rrRegister, registerReq)
	if rrRegister.Code != http.StatusOK {
		t.Fatalf("artifact register expected 200 got %d body=%s", rrRegister.Code, rrRegister.Body.String())
	}

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
		WithArgs(artifactID).
		WillReturnRows(sqlmock.NewRows([]string{"artifact_id", "tenant_id", "workspace_id", "blob_ref", "sha256", "media_type", "size_bytes", "metadata_json", "created_at"}).
			AddRow(artifactID, "ten_01", "ws_01", blobRef, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "application/json", int64(123), []byte(`{"kind":"result"}`), now))
	getReq := httptest.NewRequest(http.MethodGet, "/v1/artifacts/"+artifactID, nil)
	getReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, getReq)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("get artifact expected 200 got %d body=%s", rrGet.Code, rrGet.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
		WithArgs(artifactID).
		WillReturnRows(sqlmock.NewRows([]string{"artifact_id", "tenant_id", "workspace_id", "blob_ref", "sha256", "media_type", "size_bytes", "metadata_json", "created_at"}).
			AddRow(artifactID, "ten_01", "ws_01", blobRef, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "application/json", int64(123), []byte(`{"kind":"result"}`), now))
	downloadReq := httptest.NewRequest(http.MethodGet, "/v1/artifacts/"+artifactID+"/download-url", nil)
	downloadReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrDownload := httptest.NewRecorder()
	h.ServeHTTP(rrDownload, downloadReq)
	if rrDownload.Code != http.StatusOK {
		t.Fatalf("download-url expected 200 got %d body=%s", rrDownload.Code, rrDownload.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestArtifactRegisterRejectsInlinePayloadM3(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})

	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
	req := newJSONRequest(t, http.MethodPost, "/v1/artifacts/register", "node:ten_01:node_01:cred_01", "idem_inline_reject", map[string]any{
		"workspace_id": "ws_01",
		"blob_ref":     "s3://bucket/ten_01/ws_01/art_01_result.json",
		"sha256":       "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"media_type":   "application/json",
		"size_bytes":   10,
		"inline_bytes": "Zm9v",
	})
	rr := httptest.NewRecorder()
	srv.Handler().ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestSubmitResultSignedFlowM3(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	nodeToken := "node:ten_01:node_01:cred_01"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(nodeToken)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).
		WithArgs("lease_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_01", "task_01", "claimed"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
		WithArgs("task_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"workspace_id"}).AddRow("ws_01"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id FROM artifacts WHERE artifact_id = $1 AND tenant_id = $2 AND workspace_id = $3")).
		WithArgs("art_01", "ten_01", "ws_01").
		WillReturnRows(sqlmock.NewRows([]string{"artifact_id"}).AddRow("art_01"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO results")).
		WithArgs("res_01", "ten_01", "task_01", "lease_01", "node_01", "submitted", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO result_output_refs")).
		WithArgs("res_01", "art_01").
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO run_telemetry_refs")).
		WithArgs(sqlmock.AnyArg(), "ten_01", "lease_01", "res_01", "art_logs", "art_tools", "art_metrics").
		WillReturnResult(sqlmock.NewResult(1, 1))

	req := newJSONRequest(t, http.MethodPost, "/v1/results", nodeToken, "idem_result_m3", map[string]any{
		"result_id":   "res_01",
		"task_id":     "task_01",
		"lease_id":    "lease_01",
		"output_refs": []string{"art_01"},
		"signature":   "signed:node_01:lease_01:res_01",
		"telemetry_refs": map[string]any{
			"logs_artifact_id":       "art_logs",
			"tool_calls_artifact_id": "art_tools",
			"metrics_artifact_id":    "art_metrics",
		},
	})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("submit result expected 200 got %d body=%s", rr.Code, rr.Body.String())
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSubmitResultInvalidSignatureM3(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	nodeToken := "node:ten_01:node_01:cred_01"
	mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
		WithArgs("cred_01", "ten_01", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(nodeToken)))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).
		WithArgs("lease_01").
		WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_01", "task_01", "claimed"))

	req := newJSONRequest(t, http.MethodPost, "/v1/results", nodeToken, "idem_result_bad_sig", map[string]any{
		"result_id":   "res_01",
		"task_id":     "task_01",
		"lease_id":    "lease_01",
		"output_refs": []string{"art_01"},
		"signature":   "bad-signature",
	})
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
	}
}

func TestResultReadEndpointsM3(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()
	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	now := time.Now().UTC()
	mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "lease_id", "node_id", "status", "signature", "created_at"}).
			AddRow("res_01", "ten_01", "task_01", "lease_01", "node_01", "submitted", "signed:node_01:lease_01:res_01", now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id FROM result_output_refs WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"artifact_id"}).AddRow("art_01"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT logs_artifact_id, tool_calls_artifact_id, metrics_artifact_id FROM run_telemetry_refs WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"logs_artifact_id", "tool_calls_artifact_id", "metrics_artifact_id"}).AddRow("art_logs", "art_tools", "art_metrics"))

	getResultReq := httptest.NewRequest(http.MethodGet, "/v1/results/res_01", nil)
	getResultReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrResult := httptest.NewRecorder()
	h.ServeHTTP(rrResult, getResultReq)
	if rrResult.Code != http.StatusOK {
		t.Fatalf("get result expected 200 got %d body=%s", rrResult.Code, rrResult.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC")).
		WithArgs("task_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "lease_id", "node_id", "status", "signature", "created_at"}).
			AddRow("res_01", "ten_01", "task_01", "lease_01", "node_01", "submitted", "signed:node_01:lease_01:res_01", now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id FROM result_output_refs WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"artifact_id"}).AddRow("art_01"))

	taskResultsReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_01/results", nil)
	taskResultsReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
	rrTaskResults := httptest.NewRecorder()
	h.ServeHTTP(rrTaskResults, taskResultsReq)
	if rrTaskResults.Code != http.StatusOK {
		t.Fatalf("task results expected 200 got %d body=%s", rrTaskResults.Code, rrTaskResults.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestArtifactAndResultErrorBranchesM3(t *testing.T) {
	t.Run("upload url cross tenant forbidden", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()
		srv := newTestServer(t, dbMock, &inMemoryEventStore{})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_01", "ten_01", "node_01").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).
			WithArgs("ws_01").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_02"))

		req := newJSONRequest(t, http.MethodPost, "/v1/artifacts/upload-url", "node:ten_01:node_01:cred_01", "idem_m3_err_1", map[string]any{
			"workspace_id": "ws_01",
			"file_name":    "x.json",
			"media_type":   "application/json",
			"size_bytes":   1,
		})
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("register invalid sha", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()
		srv := newTestServer(t, dbMock, &inMemoryEventStore{})

		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_01", "ten_01", "node_01").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString("node:ten_01:node_01:cred_01")))

		req := newJSONRequest(t, http.MethodPost, "/v1/artifacts/register", "node:ten_01:node_01:cred_01", "idem_m3_err_2", map[string]any{
			"workspace_id": "ws_01",
			"blob_ref":     "s3://bucket/ten_01/ws_01/art_1.json",
			"sha256":       "bad",
			"media_type":   "application/json",
			"size_bytes":   1,
		})
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("get artifact forbidden", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()
		srv := newTestServer(t, dbMock, &inMemoryEventStore{})

		now := time.Now().UTC()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
			WithArgs("art_01").
			WillReturnRows(sqlmock.NewRows([]string{"artifact_id", "tenant_id", "workspace_id", "blob_ref", "sha256", "media_type", "size_bytes", "metadata_json", "created_at"}).
				AddRow("art_01", "ten_other", "ws_01", "s3://bucket/ten_other/ws_01/art_01.json", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "application/json", int64(1), []byte(`{}`), now))
		req := httptest.NewRequest(http.MethodGet, "/v1/artifacts/art_01", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin")
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("submit result invalid lease status", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()
		srv := newTestServer(t, dbMock, &inMemoryEventStore{})

		nodeToken := "node:ten_01:node_01:cred_01"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_01", "ten_01", "node_01").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(nodeToken)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).
			WithArgs("lease_01").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_01", "task_01", "released"))

		req := newJSONRequest(t, http.MethodPost, "/v1/results", nodeToken, "idem_m3_err_3", map[string]any{
			"result_id":   "res_01",
			"task_id":     "task_01",
			"lease_id":    "lease_01",
			"output_refs": []string{"art_01"},
			"signature":   "signed:node_01:lease_01:res_01",
		})
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("submit result output ref invalid", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()
		srv := newTestServer(t, dbMock, &inMemoryEventStore{})

		nodeToken := "node:ten_01:node_01:cred_01"
		mock.ExpectQuery(regexp.QuoteMeta("SELECT status, token_hash FROM node_credentials WHERE credential_id = $1 AND tenant_id = $2 AND node_id = $3")).
			WithArgs("cred_01", "ten_01", "node_01").
			WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(nodeToken)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).
			WithArgs("lease_01").
			WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_01", "task_01", "claimed"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"workspace_id"}).AddRow("ws_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id FROM artifacts WHERE artifact_id = $1 AND tenant_id = $2 AND workspace_id = $3")).
			WithArgs("art_bad", "ten_01", "ws_01").
			WillReturnError(sql.ErrNoRows)

		req := newJSONRequest(t, http.MethodPost, "/v1/results", nodeToken, "idem_m3_err_4", map[string]any{
			"result_id":   "res_01",
			"task_id":     "task_01",
			"lease_id":    "lease_01",
			"output_refs": []string{"art_bad"},
			"signature":   "signed:node_01:lease_01:res_01",
		})
		rr := httptest.NewRecorder()
		srv.Handler().ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})
}
