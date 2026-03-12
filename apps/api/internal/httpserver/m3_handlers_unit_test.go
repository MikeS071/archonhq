package httpserver

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"

	"github.com/MikeS071/archonhq/pkg/auth"
	"github.com/MikeS071/archonhq/pkg/objectstore"
)

func TestM3HandlersDirectBranches(t *testing.T) {
	actorNode := auth.Actor{
		Type:         "node",
		ID:           "node_1",
		TenantID:     "ten_01",
		CredentialID: "cred_1",
		TokenRaw:     "node:ten_01:node_1:cred_1",
	}
	actorHuman := auth.Actor{
		Type:     "human",
		ID:       "user_1",
		TenantID: "ten_01",
		Roles:    map[string]struct{}{"tenant_admin": {}},
	}

	t.Run("artifact upload workspace not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).WithArgs("ws_missing").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/artifacts/upload-url", `{"workspace_id":"ws_missing","file_name":"r.json","media_type":"application/json","size_bytes":10}`, actorNode)
		s.handleArtifactUploadURLV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("artifact register invalid blob namespace", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id FROM workspaces WHERE workspace_id = $1")).WithArgs("ws_01").WillReturnRows(sqlmock.NewRows([]string{"tenant_id"}).AddRow("ten_01"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/artifacts/register", `{"workspace_id":"ws_01","blob_ref":"s3://bucket/ten_02/ws_01/a.json","sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa","media_type":"application/json","size_bytes":1}`, actorNode)
		s.handleArtifactRegisterV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("get artifact not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
			WithArgs("art_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/artifacts/art_missing", "", actorHuman)
		req.SetPathValue("artifact_id", "art_missing")
		s.handleGetArtifactV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("download artifact object store health failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		s.objectStore = &objectstore.Client{}
		now := time.Now().UTC()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
			WithArgs("art_01").
			WillReturnRows(sqlmock.NewRows([]string{"artifact_id", "tenant_id", "workspace_id", "blob_ref", "sha256", "media_type", "size_bytes", "metadata_json", "created_at"}).
				AddRow("art_01", "ten_01", "ws_01", "s3://bucket/ten_01/ws_01/art_01.json", "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "application/json", int64(1), []byte(`{}`), now))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/artifacts/art_01/download-url", "", actorHuman)
		req.SetPathValue("artifact_id", "art_01")
		s.handleArtifactDownloadURLV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("download artifact not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT artifact_id, tenant_id, workspace_id, blob_ref, sha256, media_type, size_bytes, metadata_json, created_at FROM artifacts WHERE artifact_id = $1")).
			WithArgs("art_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/artifacts/art_missing/download-url", "", actorHuman)
		req.SetPathValue("artifact_id", "art_missing")
		s.handleArtifactDownloadURLV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("get result not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE result_id = $1")).
			WithArgs("res_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/results/res_missing", "", actorHuman)
		req.SetPathValue("result_id", "res_missing")
		s.handleGetResultV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("task results query fail", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, lease_id, node_id, status, signature, created_at FROM results WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC")).
			WithArgs("task_1", "ten_01").
			WillReturnError(errors.New("query failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/tasks/task_1/results", "", actorHuman)
		req.SetPathValue("task_id", "task_1")
		s.handleTaskResultsV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("submit result lease not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).WithArgs("lease_missing").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/results", `{"result_id":"res_1","task_id":"task_1","lease_id":"lease_missing","signature":"signed:node_1:lease_missing:res_1"}`, actorNode)
		s.handleSubmitResult(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("submit result forbidden lease mismatch", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).WithArgs("lease_1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_other", "node_1", "task_1", "claimed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/results", `{"result_id":"res_1","task_id":"task_1","lease_id":"lease_1","signature":"signed:node_1:lease_1:res_1"}`, actorNode)
		s.handleSubmitResult(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("submit result task lookup missing", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).WithArgs("lease_1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_1", "task_1", "claimed"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id FROM tasks WHERE task_id = $1 AND tenant_id = $2")).WithArgs("task_1", "ten_01").WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/results", `{"result_id":"res_1","task_id":"task_1","lease_id":"lease_1","signature":"signed:node_1:lease_1:res_1"}`, actorNode)
		s.handleSubmitResult(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("submit result insert failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery("SELECT status, token_hash FROM node_credentials").WillReturnRows(sqlmock.NewRows([]string{"status", "token_hash"}).AddRow("active", hashString(actorNode.TokenRaw)))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT tenant_id, node_id, task_id, status FROM leases WHERE lease_id = $1")).WithArgs("lease_1").WillReturnRows(sqlmock.NewRows([]string{"tenant_id", "node_id", "task_id", "status"}).AddRow("ten_01", "node_1", "task_1", "claimed"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT workspace_id FROM tasks WHERE task_id = $1 AND tenant_id = $2")).WithArgs("task_1", "ten_01").WillReturnRows(sqlmock.NewRows([]string{"workspace_id"}).AddRow("ws_1"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO results")).WillReturnError(errors.New("insert failed"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/results", `{"result_id":"res_1","task_id":"task_1","lease_id":"lease_1","signature":"signed:node_1:lease_1:res_1"}`, actorNode)
		s.handleSubmitResult(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})
}
