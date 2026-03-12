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
)

func TestM4HandlersDirectErrorBranches(t *testing.T) {
	actorFinance := auth.Actor{Type: "human", ID: "user_fin", TenantID: "ten_01", Roles: map[string]struct{}{"tenant_admin": {}, "finance_viewer": {}}}
	actorAuditor := auth.Actor{Type: "human", ID: "user_aud", TenantID: "ten_01", Roles: map[string]struct{}{"auditor": {}}}

	t.Run("pricing quote forbidden role", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/quote", `{"task_id":"task_1"}`, actorAuditor)
		s.handlePricingQuoteV2(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("pricing quote task not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_missing", "ten_01").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/quote", `{"task_id":"task_missing","base_rate":1}`, actorFinance)
		s.handlePricingQuoteV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("pricing quote rate resolution error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family"}).AddRow("task_01", "research.extract"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/quote", `{"task_id":"task_01","base_rate":-1}`, actorFinance)
		s.handlePricingQuoteV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("pricing quote insert error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family"}).AddRow("task_01", "research.extract"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO price_quotes")).WillReturnError(errors.New("insert fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/quote", `{"task_id":"task_01","base_rate":1.2}`, actorFinance)
		s.handlePricingQuoteV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("pricing bids invalid payload", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/bids", `{}`, actorFinance)
		s.handlePricingBidsV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("pricing bids task lookup error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_01", "ten_01").
			WillReturnError(errors.New("db fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/bids", `{"task_id":"task_01","base_rate":1}`, actorFinance)
		s.handlePricingBidsV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("pricing bids insert error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
			WithArgs("task_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family"}).AddRow("task_01", "research.extract"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO rate_snapshots")).WillReturnError(errors.New("insert fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/pricing/bids", `{"task_id":"task_01","base_rate":1}`, actorFinance)
		s.handlePricingBidsV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("reliability lookup not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT snapshot_id, tenant_id, subject_type, subject_id, family, window_name, rf_value, components_json, created_at FROM reliability_snapshots WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 ORDER BY created_at DESC LIMIT 1")).
			WithArgs("ten_01", "node", "node_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/reliability/subjects/node/node_missing", "", actorFinance)
		req.SetPathValue("subject_type", "node")
		req.SetPathValue("subject_id", "node_missing")
		s.handleReliabilitySubjectV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("ledger account cross tenant forbidden", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id, tenant_id, owner_type, owner_id, currency, status, created_at FROM ledger_accounts WHERE account_id = $1")).
			WithArgs("acct_1").
			WillReturnRows(sqlmock.NewRows([]string{"account_id", "tenant_id", "owner_type", "owner_id", "currency", "status", "created_at"}).
				AddRow("acct_1", "ten_other", "operator", "op_1", "JWUSD", "active", time.Now().UTC()))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/ledger/accounts/acct_1", "", actorFinance)
		req.SetPathValue("account_id", "acct_1")
		s.handleGetLedgerAccountV2(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("ledger entries query error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT entry_id, event_type, result_id, raw_jw, credited_jw, rate, gross_amount, reserve_amount, net_amount, status, created_at FROM ledger_entries WHERE tenant_id = $1 AND account_id = $2 ORDER BY created_at DESC")).
			WithArgs("ten_01", "acct_1").
			WillReturnError(errors.New("query fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/ledger/accounts/acct_1/entries", "", actorFinance)
		req.SetPathValue("account_id", "acct_1")
		s.handleGetLedgerAccountEntriesV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("operator earnings account not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/operators/op_missing/earnings-summary", "", actorFinance)
		req.SetPathValue("operator_id", "op_missing")
		s.handleOperatorEarningsSummaryV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("operator reserve holds query error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT reserve_hold_id, ledger_entry_id, status, release_after, released_at FROM reserve_holds WHERE tenant_id = $1 AND ledger_entry_id IN (SELECT entry_id FROM ledger_entries WHERE account_id = (SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3)) ORDER BY release_after DESC")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnError(errors.New("query fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/operators/op_01/reserve-holds", "", actorFinance)
		req.SetPathValue("operator_id", "op_01")
		s.handleOperatorReserveHoldsV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("release reserve hold forbidden", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/reserve-holds/hold_1/release", `{}`, actorAuditor)
		req.SetPathValue("reserve_hold_id", "hold_1")
		s.handleReleaseReserveHoldV2(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("release reserve hold not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectExec(regexp.QuoteMeta("UPDATE reserve_holds SET status = $1, released_at = $2 WHERE reserve_hold_id = $3 AND tenant_id = $4")).
			WithArgs("released", sqlmock.AnyArg(), "hold_missing", "ten_01").
			WillReturnResult(sqlmock.NewResult(0, 0))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/reserve-holds/hold_missing/release", `{}`, actorFinance)
		req.SetPathValue("reserve_hold_id", "hold_missing")
		s.handleReleaseReserveHoldV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("release reserve hold missing id", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/reserve-holds//release", `{}`, actorFinance)
		s.handleReleaseReserveHoldV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("settlement invalid role", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/ledger/settlements", `{"result_id":"res_01"}`, actorAuditor)
		s.handlePostSettlementV2(rr, req)
		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("settlement result not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1")).
			WithArgs("res_missing").
			WillReturnError(sql.ErrNoRows)
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/ledger/settlements", `{"result_id":"res_missing","rate":1}`, actorFinance)
		s.handlePostSettlementV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("settlement account resolution failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1")).
			WithArgs("res_01").
			WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "node_id"}).AddRow("res_01", "ten_01", "task_01", "node_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT operator_id FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
			WithArgs("node_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"operator_id"}).AddRow("op_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnError(errors.New("db fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/ledger/settlements", `{"result_id":"res_01","rate":1}`, actorFinance)
		s.handlePostSettlementV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("settlement engine validation failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1")).
			WithArgs("res_01").
			WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "node_id"}).AddRow("res_01", "ten_01", "task_01", "node_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT operator_id FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
			WithArgs("node_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"operator_id"}).AddRow("op_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow("acct_01"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/ledger/settlements", `{"result_id":"res_01","rate":0}`, actorFinance)
		s.handlePostSettlementV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("settlement ledger entry insert failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1")).
			WithArgs("res_01").
			WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "node_id"}).AddRow("res_01", "ten_01", "task_01", "node_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT operator_id FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
			WithArgs("node_01", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"operator_id"}).AddRow("op_01"))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow("acct_01"))
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ledger_entries")).WillReturnError(errors.New("insert fail"))
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/ledger/settlements", `{"result_id":"res_01","rate":1}`, actorFinance)
		s.handlePostSettlementV2(rr, req)
		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})
}

func TestGetOrCreateOperatorAccount(t *testing.T) {
	actorFinance := auth.Actor{Type: "human", ID: "user_fin", TenantID: "ten_01", Roles: map[string]struct{}{"tenant_admin": {}, "finance_viewer": {}}}

	t.Run("lookup existing account", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow("acct_01"))

		req := reqWithActor(http.MethodGet, "/", "", actorFinance)
		got, err := s.getOrCreateOperatorAccount(req, "ten_01", "op_01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got != "acct_01" {
			t.Fatalf("unexpected account id: %s", got)
		}
	})

	t.Run("create new account", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ledger_accounts")).WillReturnResult(sqlmock.NewResult(1, 1))

		req := reqWithActor(http.MethodGet, "/", "", actorFinance)
		got, err := s.getOrCreateOperatorAccount(req, "ten_01", "op_01")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got == "" {
			t.Fatalf("expected generated account id")
		}
	})

	t.Run("insert failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnError(sql.ErrNoRows)
		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ledger_accounts")).WillReturnError(errors.New("insert fail"))

		req := reqWithActor(http.MethodGet, "/", "", actorFinance)
		if _, err := s.getOrCreateOperatorAccount(req, "ten_01", "op_01"); err == nil {
			t.Fatalf("expected error")
		}
	})

	t.Run("lookup query failure", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()
		mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
			WithArgs("ten_01", "operator", "op_01").
			WillReturnError(errors.New("query fail"))

		req := reqWithActor(http.MethodGet, "/", "", actorFinance)
		if _, err := s.getOrCreateOperatorAccount(req, "ten_01", "op_01"); err == nil {
			t.Fatalf("expected error")
		}
	})
}
