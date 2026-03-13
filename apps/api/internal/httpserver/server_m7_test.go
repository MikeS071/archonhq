package httpserver

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestM7DecomposeAndMarketEndpoints(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
		WithArgs("task_code_1").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
			AddRow("task_code_1", "ten_01", "ws_01", "code.patch", "Patch app", "approved", "ast_patch_v1"))

	decomposeReq := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_code_1/decompose", "human:ten_01:user_admin:tenant_admin,developer", "idem_decompose_1", map[string]any{
		"max_children": 2,
	})
	rrDecompose := httptest.NewRecorder()
	h.ServeHTTP(rrDecompose, decomposeReq)
	if rrDecompose.Code != http.StatusOK {
		t.Fatalf("decompose expected 200 got %d body=%s", rrDecompose.Code, rrDecompose.Body.String())
	}
	if !strings.Contains(rrDecompose.Body.String(), `"merge_strategy":"ast_patch_v1"`) {
		t.Fatalf("decompose expected ast_patch_v1 body=%s", rrDecompose.Body.String())
	}
	if !strings.Contains(rrDecompose.Body.String(), `code_patch_merge_storm_v1`) {
		t.Fatalf("decompose expected simulation entrypoint body=%s", rrDecompose.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
		WithArgs("task_code_1").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
			AddRow("task_code_1", "ten_01", "ws_01", "code.patch", "Patch app", "approved", "ast_patch_v1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT quote_id, strategy_name, quote_json, created_at FROM price_quotes WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1")).
		WithArgs("task_code_1", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "strategy_name", "quote_json", "created_at"}).
			AddRow("quote_01", "fixed_plus_bid", []byte(`{"rate_value":1.25}`), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT rate_snapshot_id, strategy_name, rate_value, metadata_json, created_at FROM rate_snapshots WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1")).
		WithArgs("task_code_1", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"rate_snapshot_id", "strategy_name", "rate_value", "metadata_json", "created_at"}).
			AddRow("rate_01", "fixed_plus_bid", 1.25, []byte(`{"estimated_net":2.11}`), now))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM results WHERE task_id = $1 AND tenant_id = $2")).
		WithArgs("task_code_1", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(3))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COUNT(*) FROM verifications v WHERE v.tenant_id = $1 AND v.result_id IN (SELECT result_id FROM results WHERE task_id = $2 AND tenant_id = $1) AND v.decision = 'accepted'")).
		WithArgs("ten_01", "task_code_1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT reduction_id, created_at FROM reductions WHERE task_id = $1 AND tenant_id = $2 ORDER BY created_at DESC LIMIT 1")).
		WithArgs("task_code_1", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"reduction_id", "created_at"}).AddRow("red_01", now))

	marketReq := httptest.NewRequest(http.MethodGet, "/v1/tasks/task_code_1/market", nil)
	marketReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,developer")
	rrMarket := httptest.NewRecorder()
	h.ServeHTTP(rrMarket, marketReq)
	if rrMarket.Code != http.StatusOK {
		t.Fatalf("market expected 200 got %d body=%s", rrMarket.Code, rrMarket.Body.String())
	}
	if !strings.Contains(rrMarket.Body.String(), `"verification_acceptance_rate"`) {
		t.Fatalf("market expected acceptance rate signal body=%s", rrMarket.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM7AutoModeReductionAndAutosearchGuardrails(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, status, payload_json FROM approval_requests WHERE approval_id = $1 AND tenant_id = $2")).
		WithArgs("apr_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "status", "payload_json"}).
			AddRow("task_auto_1", "pending", []byte(`{"reason":"initial"}`)))
	mock.ExpectExec(regexp.QuoteMeta("UPDATE approval_requests SET payload_json = $1 WHERE approval_id = $2 AND tenant_id = $3")).
		WithArgs(sqlmock.AnyArg(), "apr_01", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))

	autoReq := newJSONRequest(t, http.MethodPost, "/v1/approvals/apr_01/auto-mode", "human:ten_01:user_approver:approver", "idem_auto_mode_1", map[string]any{
		"max_iterations":   5,
		"budget_limit_jw":  12.5,
		"require_approval": true,
		"approval_granted": true,
	})
	rrAuto := httptest.NewRecorder()
	h.ServeHTTP(rrAuto, autoReq)
	if rrAuto.Code != http.StatusOK {
		t.Fatalf("auto-mode expected 200 got %d body=%s", rrAuto.Code, rrAuto.Body.String())
	}
	if !strings.Contains(rrAuto.Body.String(), `"max_iterations":5`) {
		t.Fatalf("auto-mode expected max_iterations body=%s", rrAuto.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
		WithArgs("task_auto_1").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
			AddRow("task_auto_1", "ten_01", "ws_01", "autosearch.self_improve", "Improve benchmark", "approved", "topk_rank_v1"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO reductions (reduction_id, tenant_id, task_id, strategy, output_state_ref, decision_json) VALUES ($1,$2,$3,$4,$5,$6)")).
		WithArgs("red_m7_1", "ten_01", "task_auto_1", "topk_rank_v1", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	reductionReq := newJSONRequest(t, http.MethodPost, "/v1/reductions", "human:ten_01:user_admin:tenant_admin,developer", "idem_reduce_1", map[string]any{
		"reduction_id": "red_m7_1",
		"task_id":      "task_auto_1",
		"strategy":     "topk_rank_v1",
		"candidates": []map[string]any{
			{"result_id": "res_01", "score": 0.81, "patch_ops": []string{"op_a"}},
			{"result_id": "res_02", "score": 0.74, "patch_ops": []string{"op_b"}},
		},
		"autosearch": map[string]any{
			"enabled":          true,
			"experiment_id":    "exp_m7_1",
			"max_iterations":   3,
			"budget_limit_jw":  8.0,
			"require_approval": false,
			"approval_granted": true,
			"min_accept_score": 0.62,
		},
	})
	rrReduction := httptest.NewRecorder()
	h.ServeHTTP(rrReduction, reductionReq)
	if rrReduction.Code != http.StatusOK {
		t.Fatalf("reduction expected 200 got %d body=%s", rrReduction.Code, rrReduction.Body.String())
	}
	if !strings.Contains(rrReduction.Body.String(), `"reduction_id":"red_m7_1"`) {
		t.Fatalf("reduction expected id body=%s", rrReduction.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT reduction_id, tenant_id, task_id, strategy, output_state_ref, decision_json, created_at FROM reductions WHERE reduction_id = $1")).
		WithArgs("red_m7_1").
		WillReturnRows(sqlmock.NewRows([]string{"reduction_id", "tenant_id", "task_id", "strategy", "output_state_ref", "decision_json", "created_at"}).
			AddRow("red_m7_1", "ten_01", "task_auto_1", "topk_rank_v1", "state_01", []byte(`{"status":"accepted","lineage":{"task_id":"task_auto_1"}}`), now))

	getReductionReq := httptest.NewRequest(http.MethodGet, "/v1/reductions/red_m7_1", nil)
	getReductionReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,developer")
	rrGetReduction := httptest.NewRecorder()
	h.ServeHTTP(rrGetReduction, getReductionReq)
	if rrGetReduction.Code != http.StatusOK {
		t.Fatalf("get reduction expected 200 got %d body=%s", rrGetReduction.Code, rrGetReduction.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, tenant_id, workspace_id, task_family, title, status, merge_strategy FROM tasks WHERE task_id = $1")).
		WithArgs("task_auto_1").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "tenant_id", "workspace_id", "task_family", "title", "status", "merge_strategy"}).
			AddRow("task_auto_1", "ten_01", "ws_01", "autosearch.self_improve", "Improve benchmark", "approved", "topk_rank_v1"))

	decomposeGuardrailReq := newJSONRequest(t, http.MethodPost, "/v1/tasks/task_auto_1/decompose", "human:ten_01:user_admin:tenant_admin,developer", "idem_decompose_guardrail", map[string]any{
		"autosearch": map[string]any{
			"require_approval": true,
			"approval_granted": false,
		},
	})
	rrGuardrail := httptest.NewRecorder()
	h.ServeHTTP(rrGuardrail, decomposeGuardrailReq)
	if rrGuardrail.Code != http.StatusConflict {
		t.Fatalf("decompose guardrail expected 409 got %d body=%s", rrGuardrail.Code, rrGuardrail.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestM7VerificationEndpoints(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()
	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id FROM results WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id"}).
			AddRow("res_01", "ten_01", "task_auto_1"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
		WithArgs("task_auto_1", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_family"}).AddRow("autosearch.self_improve"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO verifications (verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)")).
		WithArgs("ver_m7_1", "ten_01", "res_01", "benchmark/eval", "verifier_01", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	createVerificationReq := newJSONRequest(t, http.MethodPost, "/v1/verifications", "human:ten_01:user_admin:tenant_admin,developer", "idem_verification_1", map[string]any{
		"verification_id": "ver_m7_1",
		"result_id":       "res_01",
		"verifier_type":   "benchmark/eval",
		"verifier_id":     "verifier_01",
		"iterative": map[string]any{
			"experiment_id":        "exp_01",
			"iteration":            1,
			"candidate_id":         "cand_1",
			"benchmark_delta_norm": 0.8,
			"eval_reproducibility": 0.9,
			"rollback_safety":      0.9,
			"search_novelty":       0.6,
			"compute_efficiency":   0.7,
		},
	})
	rrCreate := httptest.NewRecorder()
	h.ServeHTTP(rrCreate, createVerificationReq)
	if rrCreate.Code != http.StatusOK {
		t.Fatalf("create verification expected 200 got %d body=%s", rrCreate.Code, rrCreate.Body.String())
	}
	if !strings.Contains(rrCreate.Body.String(), `"verification_id":"ver_m7_1"`) {
		t.Fatalf("create verification expected id body=%s", rrCreate.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT verification_id, tenant_id, result_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE verification_id = $1")).
		WithArgs("ver_m7_1").
		WillReturnRows(sqlmock.NewRows([]string{"verification_id", "tenant_id", "result_id", "verifier_type", "verifier_id", "score", "decision", "report_json", "created_at"}).
			AddRow("ver_m7_1", "ten_01", "res_01", "benchmark/eval", "verifier_01", 0.84, "accepted", []byte(`{"lineage":{"task_id":"task_auto_1"}}`), now))

	getVerificationReq := httptest.NewRequest(http.MethodGet, "/v1/verifications/ver_m7_1", nil)
	getVerificationReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,developer")
	rrGet := httptest.NewRecorder()
	h.ServeHTTP(rrGet, getVerificationReq)
	if rrGet.Code != http.StatusOK {
		t.Fatalf("get verification expected 200 got %d body=%s", rrGet.Code, rrGet.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT verification_id, verifier_type, verifier_id, score, decision, report_json, created_at FROM verifications WHERE result_id = $1 AND tenant_id = $2 ORDER BY created_at DESC")).
		WithArgs("res_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"verification_id", "verifier_type", "verifier_id", "score", "decision", "report_json", "created_at"}).
			AddRow("ver_m7_1", "benchmark/eval", "verifier_01", 0.84, "accepted", []byte(`{"lineage":{"task_id":"task_auto_1"}}`), now))

	listReq := httptest.NewRequest(http.MethodGet, "/v1/results/res_01/verifications", nil)
	listReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,developer")
	rrList := httptest.NewRecorder()
	h.ServeHTTP(rrList, listReq)
	if rrList.Code != http.StatusOK {
		t.Fatalf("list verifications expected 200 got %d body=%s", rrList.Code, rrList.Body.String())
	}
	if !strings.Contains(rrList.Body.String(), `"verifications"`) {
		t.Fatalf("list verifications expected payload body=%s", rrList.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
