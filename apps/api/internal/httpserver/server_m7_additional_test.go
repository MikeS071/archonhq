package httpserver

import (
	"database/sql"
	"errors"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestM7HandlerErrorPaths(t *testing.T) {
	t.Run("decompose forbidden role", func(t *testing.T) {
		dbMock, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		req := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_1/decompose", "human:ten_01:user_a:auditor", "idem_m7_err_1", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("decompose invalid max children", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_1").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
				AddRow("task_1", "ten_01", "ws_01", "code.patch", "Patch app", "approved", "ast_patch_v1"))

		req := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_1/decompose", "human:ten_01:user_a:developer", "idem_m7_err_2", map[string]any{
			"max_children": 99,
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

	t.Run("auto-mode invalid iterations", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status, payload_json FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "status", "payload_json"}).AddRow("task_1", "pending", []byte(`{}`)))

		req := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_1/auto-mode", "human:ten_01:user_a:approver", "idem_m7_err_3", map[string]any{
			"max_iterations": 26,
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

	t.Run("create verification missing fields", func(t *testing.T) {
		dbMock, _, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		req := newJSONRequest(t, http.MethodPost, "/v1/verifications", "human:ten_01:user_a:developer", "idem_m7_err_4", map[string]any{
			"result_id": "",
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
	})

	t.Run("get verification not found", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE verification_id = $1")).
			WithArgs("ver_missing").
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/v1/verifications/ver_missing", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_a:developer")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("result verifications query failure", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT verification_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE result_id = $1 AND tenant_id = $2 ORDER BY created_at DESC")).
			WithArgs("res_1", "ten_01").
			WillReturnError(errors.New("boom"))

		req := httptest.NewRequest(http.MethodGet, "/v1/results/res_1/verifications", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_a:developer")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("reduction unsupported strategy", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_1").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
				AddRow("task_1", "ten_01", "ws_01", "code.patch", "Patch app", "approved", "ast_patch_v1"))

		req := newJSONRequest(t, http.MethodPost, "/v1/reductions", "human:ten_01:user_a:developer", "idem_m7_err_5", map[string]any{
			"task_id":    "task_1",
			"strategy":   "bad_strategy",
			"candidates": []map[string]any{{"result_id": "res_1", "score": 0.8}},
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d body=%s", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "merge_strategy_unsupported") {
			t.Fatalf("expected merge_strategy_unsupported body=%s", rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("reduction candidates required", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_1").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
				AddRow("task_1", "ten_01", "ws_01", "code.patch", "Patch app", "approved", "ast_patch_v1"))

		req := newJSONRequest(t, http.MethodPost, "/v1/reductions", "human:ten_01:user_a:developer", "idem_m7_err_6", map[string]any{
			"task_id":    "task_1",
			"candidates": []map[string]any{},
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

	t.Run("get reduction not found", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT reduction_id, tenant_id, task_id, strategy, output_state_ref, decision_json, created_at FROM reductions WHERE reduction_id = $1")).
			WithArgs("red_missing").
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/v1/reductions/red_missing", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_a:developer")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("task market not found", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_missing").
			WillReturnError(sql.ErrNoRows)

		req := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_missing/market", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_a:developer")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("decompose autosearch invalid loop policy", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_auto_1").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
				AddRow("task_auto_1", "ten_01", "ws_01", "autosearch.self_improve", "Improve", "approved", "topk_rank_v1"))

		req := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_auto_1/decompose", "human:ten_01:user_a:developer", "idem_m7_err_7", map[string]any{
			"autosearch": map[string]any{
				"max_iterations": 26,
			},
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

	t.Run("auto-mode conflict for non-pending status", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status, payload_json FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
			WithArgs("apr_closed", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "status", "payload_json"}).AddRow("task_1", "denied", []byte(`{}`)))

		req := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_closed/auto-mode", "human:ten_01:user_a:approver", "idem_m7_err_8", map[string]any{})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("create verification cross-tenant forbidden", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id FROM results WHERE result_id = $1")).
			WithArgs("res_other").
			WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id"}).AddRow("res_other", "ten_02", "task_2"))

		req := newJSONRequest(t, http.MethodPost, "/v1/verifications", "human:ten_01:user_a:developer", "idem_m7_err_9", map[string]any{
			"result_id":     "res_other",
			"verifier_type": "schema",
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

	t.Run("get verification cross-tenant forbidden", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE verification_id = $1")).
			WithArgs("ver_other").
			WillReturnRows(sqlmock.NewRows([]string{"verification_id", "tenant_id", "result_id", "verifier_type", "verifier_id", "score", "decision", "report_json", "created_at"}).
				AddRow("ver_other", "ten_02", "res_other", "schema", "ver_1", 0.8, "accepted", []byte(`{}`), time.Now().UTC()))

		req := httptest.NewRequest(http.MethodGet, "/v1/verifications/ver_other", nil)
		req.Header.Set("Authorization", "Bearer human:ten_01:user_a:developer")
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})

	t.Run("reduction autosearch requires approval", func(t *testing.T) {
		dbMock, mock, err := sqlmock.New()
		if err != nil {
			t.Fatalf("sqlmock new: %v", err)
		}
		defer dbMock.Close()

		srv := newTestServer(t, dbMock, &inMemoryEventStore{})
		h := srv.Handler()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
			WithArgs("task_auto_2").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
				AddRow("task_auto_2", "ten_01", "ws_01", "autosearch.self_improve", "Improve", "approved", "topk_rank_v1"))

		req := newJSONRequest(t, http.MethodPost, "/v1/reductions", "human:ten_01:user_a:developer", "idem_m7_err_10", map[string]any{
			"task_id": "task_auto_2",
			"candidates": []map[string]any{
				{"result_id": "res_1", "score": 0.7},
			},
			"autosearch": map[string]any{
				"enabled":          true,
				"require_approval": true,
				"approval_granted": false,
			},
		})
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusConflict {
			t.Fatalf("expected 409 got %d body=%s", rr.Code, rr.Body.String())
		}
		if err := mock.ExpectationsWereMet(); err != nil {
			t.Fatalf("sql expectations: %v", err)
		}
	})
}

func TestM7HelperFunctions(t *testing.T) {
	defaultReq := buildAutosearchLoopRequest(nil)
	if defaultReq.Policy.MaxIterations != defaultAutoMaxIterations {
		t.Fatalf("unexpected default max iterations %d", defaultReq.Policy.MaxIterations)
	}
	if len(defaultReq.Candidates) == 0 {
		t.Fatalf("expected default candidates")
	}

	customReq := buildAutosearchLoopRequest(&boundedAutosearchInput{
		ExperimentID:             "exp_custom",
		MaxIterations:            2,
		BudgetLimitJW:            3.5,
		RequireApproval:          true,
		ApprovalGranted:          true,
		MinAcceptScore:           0.7,
		CandidateBenchmarkDeltas: []float64{0.1, 0.2},
	})
	if customReq.ExperimentID != "exp_custom" {
		t.Fatalf("unexpected experiment id %q", customReq.ExperimentID)
	}
	if len(customReq.Candidates) != 2 {
		t.Fatalf("expected 2 custom candidates got %d", len(customReq.Candidates))
	}

	codePatchEntrypoints := buildSimulationEntrypoints("code.patch", "ast_patch_v1")
	if len(codePatchEntrypoints) < 2 {
		t.Fatalf("expected code patch simulation entrypoints")
	}
	autosearchEntrypoints := buildSimulationEntrypoints("autosearch.self_improve", "topk_rank_v1")
	if len(autosearchEntrypoints) < 2 {
		t.Fatalf("expected autosearch simulation entrypoints")
	}

	if got := nonEmpty("x", "y"); got != "x" {
		t.Fatalf("nonEmpty returned %q", got)
	}
	if got := nonEmpty("", "y"); got != "y" {
		t.Fatalf("nonEmpty fallback returned %q", got)
	}
}
