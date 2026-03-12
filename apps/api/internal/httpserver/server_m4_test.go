package httpserver

import (
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"
	"time"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestPricingEndpointsM4(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
		WithArgs("task_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family"}).AddRow("task_01", "research.extract"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO price_quotes")).
		WithArgs("quote_01", "ten_01", "task_01", "fixed_plus_bid", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	quoteReq := newJSONRequest(t, http.MethodPost, "/v1/pricing/quote", "human:ten_01:user_admin:tenant_admin,operator", "idem_pricing_quote_1", map[string]any{
		"quote_id":          "quote_01",
		"task_id":           "task_01",
		"base_rate":         2.0,
		"bid_adjustment":    0.5,
		"predicted_raw_jw":  10.0,
		"quality_factor":    0.8,
		"reliability_rf":    0.9,
		"reserve_ratio":     0.15,
		"strategy_override": "fixed_plus_bid",
	})
	rrQuote := httptest.NewRecorder()
	h.ServeHTTP(rrQuote, quoteReq)
	if rrQuote.Code != http.StatusOK {
		t.Fatalf("pricing quote expected 200 got %d body=%s", rrQuote.Code, rrQuote.Body.String())
	}

	rateCardsReq := httptest.NewRequest(http.MethodGet, "/v1/pricing/rate-cards", nil)
	rateCardsReq.Header.Set("Authorization", "Bearer human:ten_01:user_admin:tenant_admin,operator")
	rrRateCards := httptest.NewRecorder()
	h.ServeHTTP(rrRateCards, rateCardsReq)
	if rrRateCards.Code != http.StatusOK {
		t.Fatalf("rate cards expected 200 got %d body=%s", rrRateCards.Code, rrRateCards.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT task_id, task_family FROM tasks WHERE task_id = $1 AND tenant_id = $2")).
		WithArgs("task_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"task_id", "task_family"}).AddRow("task_01", "research.extract"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO rate_snapshots")).
		WithArgs("rate_01", "ten_01", "task_01", "res_01", "fixed_plus_bid", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	bidsReq := newJSONRequest(t, http.MethodPost, "/v1/pricing/bids", "human:ten_01:user_admin:tenant_admin,operator", "idem_pricing_bid_1", map[string]any{
		"rate_snapshot_id": "rate_01",
		"task_id":          "task_01",
		"result_id":        "res_01",
		"base_rate":        2.0,
		"bid_adjustment":   0.25,
		"quality_factor":   0.85,
		"reliability_rf":   0.93,
	})
	rrBids := httptest.NewRecorder()
	h.ServeHTTP(rrBids, bidsReq)
	if rrBids.Code != http.StatusOK {
		t.Fatalf("pricing bids expected 200 got %d body=%s", rrBids.Code, rrBids.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}

func TestSettlementReliabilityAndLedgerEndpointsM4(t *testing.T) {
	dbMock, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	srv := newTestServer(t, dbMock, &inMemoryEventStore{})
	h := srv.Handler()

	now := time.Now().UTC()

	mock.ExpectQuery(regexp.QuoteMeta("SELECT result_id, tenant_id, task_id, node_id FROM results WHERE result_id = $1")).
		WithArgs("res_01").
		WillReturnRows(sqlmock.NewRows([]string{"result_id", "tenant_id", "task_id", "node_id"}).AddRow("res_01", "ten_01", "task_01", "node_01"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT operator_id FROM nodes WHERE node_id = $1 AND tenant_id = $2")).
		WithArgs("node_01", "ten_01").
		WillReturnRows(sqlmock.NewRows([]string{"operator_id"}).AddRow("op_01"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
		WithArgs("ten_01", "operator", "op_01").
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow("acct_01"))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO ledger_entries")).
		WithArgs("entry_01", "ten_01", "acct_01", "ledger.settlement_posted", "res_01", sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), sqlmock.AnyArg(), "posted", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO reserve_holds")).
		WithArgs("hold_01", "ten_01", "entry_01", "held", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO reliability_snapshots")).
		WithArgs("rel_node_01", "ten_01", "node", "node_01", "task", "last_30d", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(regexp.QuoteMeta("INSERT INTO reliability_snapshots")).
		WithArgs("rel_op_01", "ten_01", "operator", "op_01", "fleet", "last_30d", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(1, 1))

	settlementReq := newJSONRequest(t, http.MethodPost, "/v1/ledger/settlements", "human:ten_01:user_finance:tenant_admin,finance_viewer", "idem_settle_1", map[string]any{
		"entry_id":         "entry_01",
		"reserve_hold_id":  "hold_01",
		"result_id":        "res_01",
		"rate":             2.0,
		"task_difficulty":  "standard",
		"reserve_ratio":    0.1,
		"node_snapshot_id": "rel_node_01",
		"op_snapshot_id":   "rel_op_01",
		"metering": map[string]any{
			"cpu_sec":   100,
			"gpu_class": "cpu-only",
		},
		"quality_inputs": map[string]any{
			"validity":          1.0,
			"verifier_score":    1.0,
			"acceptance_signal": 1.0,
			"novelty":           1.0,
			"latency_score":     1.0,
		},
		"rf_last_100":           0.95,
		"rf_last_30d":           0.95,
		"rf_lifetime":           0.95,
		"release_after_seconds": 3600,
	})
	rrSettlement := httptest.NewRecorder()
	h.ServeHTTP(rrSettlement, settlementReq)
	if rrSettlement.Code != http.StatusOK {
		t.Fatalf("ledger settlement expected 200 got %d body=%s", rrSettlement.Code, rrSettlement.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id, tenant_id, owner_type, owner_id, currency, status, created_at FROM ledger_accounts WHERE account_id = $1")).
		WithArgs("acct_01").
		WillReturnRows(sqlmock.NewRows([]string{"account_id", "tenant_id", "owner_type", "owner_id", "currency", "status", "created_at"}).
			AddRow("acct_01", "ten_01", "operator", "op_01", "JWUSD", "active", now))
	getAccount := httptest.NewRequest(http.MethodGet, "/v1/ledger/accounts/acct_01", nil)
	getAccount.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrAccount := httptest.NewRecorder()
	h.ServeHTTP(rrAccount, getAccount)
	if rrAccount.Code != http.StatusOK {
		t.Fatalf("get ledger account expected 200 got %d body=%s", rrAccount.Code, rrAccount.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT entry_id, event_type, result_id, raw_jw, credited_jw, rate, gross_amount, reserve_amount, net_amount, status, created_at FROM ledger_entries WHERE tenant_id = $1 AND account_id = $2 ORDER BY created_at DESC")).
		WithArgs("ten_01", "acct_01").
		WillReturnRows(sqlmock.NewRows([]string{"entry_id", "event_type", "result_id", "raw_jw", "credited_jw", "rate", "gross_amount", "reserve_amount", "net_amount", "status", "created_at"}).
			AddRow("entry_01", "ledger.settlement_posted", "res_01", 0.2, 0.2, 2.0, 0.4, 0.04, 0.36, "posted", now))
	getEntries := httptest.NewRequest(http.MethodGet, "/v1/ledger/accounts/acct_01/entries", nil)
	getEntries.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrEntries := httptest.NewRecorder()
	h.ServeHTTP(rrEntries, getEntries)
	if rrEntries.Code != http.StatusOK {
		t.Fatalf("get ledger entries expected 200 got %d body=%s", rrEntries.Code, rrEntries.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3")).
		WithArgs("ten_01", "operator", "op_01").
		WillReturnRows(sqlmock.NewRows([]string{"account_id"}).AddRow("acct_01"))
	mock.ExpectQuery(regexp.QuoteMeta("SELECT COALESCE(SUM(credited_jw),0), COALESCE(SUM(net_amount),0), COALESCE(SUM(reserve_amount),0) FROM ledger_entries WHERE tenant_id = $1 AND account_id = $2")).
		WithArgs("ten_01", "acct_01").
		WillReturnRows(sqlmock.NewRows([]string{"credited", "net", "reserve"}).AddRow(12.5, 30.0, 2.0))
	earningsReq := httptest.NewRequest(http.MethodGet, "/v1/operators/op_01/earnings-summary", nil)
	earningsReq.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrEarnings := httptest.NewRecorder()
	h.ServeHTTP(rrEarnings, earningsReq)
	if rrEarnings.Code != http.StatusOK {
		t.Fatalf("operator earnings expected 200 got %d body=%s", rrEarnings.Code, rrEarnings.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT reserve_hold_id, ledger_entry_id, status, release_after, released_at FROM reserve_holds WHERE tenant_id = $1 AND ledger_entry_id IN (SELECT entry_id FROM ledger_entries WHERE account_id = (SELECT account_id FROM ledger_accounts WHERE tenant_id = $1 AND owner_type = $2 AND owner_id = $3)) ORDER BY release_after DESC")).
		WithArgs("ten_01", "operator", "op_01").
		WillReturnRows(sqlmock.NewRows([]string{"reserve_hold_id", "ledger_entry_id", "status", "release_after", "released_at"}).
			AddRow("hold_01", "entry_01", "held", now.Add(time.Hour), nil))
	reserveReq := httptest.NewRequest(http.MethodGet, "/v1/operators/op_01/reserve-holds", nil)
	reserveReq.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrReserve := httptest.NewRecorder()
	h.ServeHTTP(rrReserve, reserveReq)
	if rrReserve.Code != http.StatusOK {
		t.Fatalf("reserve holds expected 200 got %d body=%s", rrReserve.Code, rrReserve.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT snapshot_id, tenant_id, subject_type, subject_id, family, window_name, rf_value, components_json, created_at FROM reliability_snapshots WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 ORDER BY created_at DESC LIMIT 1")).
		WithArgs("ten_01", "node", "node_01").
		WillReturnRows(sqlmock.NewRows([]string{"snapshot_id", "tenant_id", "subject_type", "subject_id", "family", "window_name", "rf_value", "components_json", "created_at"}).
			AddRow("rel_node_01", "ten_01", "node", "node_01", "task", "last_30d", 0.95, []byte(`{"quality_factor":1}`), now))
	reliabilityReq := httptest.NewRequest(http.MethodGet, "/v1/reliability/subjects/node/node_01", nil)
	reliabilityReq.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrReliability := httptest.NewRecorder()
	h.ServeHTTP(rrReliability, reliabilityReq)
	if rrReliability.Code != http.StatusOK {
		t.Fatalf("reliability subject expected 200 got %d body=%s", rrReliability.Code, rrReliability.Body.String())
	}

	mock.ExpectQuery(regexp.QuoteMeta("SELECT snapshot_id, tenant_id, subject_type, subject_id, family, window_name, rf_value, components_json, created_at FROM reliability_snapshots WHERE tenant_id = $1 AND subject_type = $2 AND subject_id = $3 ORDER BY created_at DESC LIMIT 1")).
		WithArgs("ten_01", "operator", "op_01").
		WillReturnRows(sqlmock.NewRows([]string{"snapshot_id", "tenant_id", "subject_type", "subject_id", "family", "window_name", "rf_value", "components_json", "created_at"}).
			AddRow("rel_op_01", "ten_01", "operator", "op_01", "fleet", "last_30d", 0.97, []byte(`{"quality_factor":1}`), now))
	operatorReliabilityReq := httptest.NewRequest(http.MethodGet, "/v1/operators/op_01/reliability", nil)
	operatorReliabilityReq.Header.Set("Authorization", "Bearer human:ten_01:user_finance:tenant_admin,finance_viewer")
	rrOperatorReliability := httptest.NewRecorder()
	h.ServeHTTP(rrOperatorReliability, operatorReliabilityReq)
	if rrOperatorReliability.Code != http.StatusOK {
		t.Fatalf("operator reliability expected 200 got %d body=%s", rrOperatorReliability.Code, rrOperatorReliability.Body.String())
	}

	mock.ExpectExec(regexp.QuoteMeta("UPDATE reserve_holds SET status = $1, released_at = $2 WHERE reserve_hold_id = $3 AND tenant_id = $4")).
		WithArgs("released", sqlmock.AnyArg(), "hold_01", "ten_01").
		WillReturnResult(sqlmock.NewResult(1, 1))
	releaseReq := newJSONRequest(t, http.MethodPost, "/v1/reserve-holds/hold_01/release", "human:ten_01:user_finance:tenant_admin,finance_viewer", "idem_release_1", map[string]any{})
	rrRelease := httptest.NewRecorder()
	h.ServeHTTP(rrRelease, releaseReq)
	if rrRelease.Code != http.StatusOK {
		t.Fatalf("reserve release expected 200 got %d body=%s", rrRelease.Code, rrRelease.Body.String())
	}

	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("sql expectations: %v", err)
	}
}
