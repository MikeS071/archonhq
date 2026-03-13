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

	"github.com/MikeS071/archonhq/pkg/auth"
)

func TestM5ListNodesAndPolicies(t *testing.T) {
	actorAdmin := auth.Actor{Type: "human", ID: "user_admin", TenantID: "ten_01", Roles: map[string]struct{}{"tenant_admin": {}}}
	actorOperator := auth.Actor{Type: "human", ID: "user_op", TenantID: "ten_01", Roles: map[string]struct{}{"operator": {}}}

	t.Run("list nodes success", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		rows := sqlmock.NewRows([]string{"node_id", "operator_id", "runtime_type", "runtime_version", "status", "last_heartbeat_at", "active_leases"}).
			AddRow("node_1", "op_1", "hermes", "1.9", "healthy", time.Now().UTC(), 2)

		mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  n.node_id,")).WithArgs("ten_01", 200).WillReturnRows(rows)

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/nodes", "", actorOperator)
		s.handleListNodesV2(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "node_1") {
			t.Fatalf("expected node in response: %s", rr.Body.String())
		}
	})

	t.Run("list nodes bad limit", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/nodes?limit=bad", "", actorOperator)
		s.handleListNodesV2(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("list nodes query error", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT\n  n.node_id,")).WithArgs("ten_01", 200).WillReturnError(errors.New("boom"))

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/nodes", "", actorOperator)
		s.handleListNodesV2(rr, req)

		if rr.Code != http.StatusInternalServerError {
			t.Fatalf("expected 500 got %d", rr.Code)
		}
	})

	t.Run("get policies forbidden", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()

		actorDeveloper := auth.Actor{Type: "human", ID: "dev", TenantID: "ten_01", Roles: map[string]struct{}{"developer": {}}}
		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/policies", "", actorDeveloper)
		s.handleGetPoliciesV2(rr, req)

		if rr.Code != http.StatusForbidden {
			t.Fatalf("expected 403 got %d", rr.Code)
		}
	})

	t.Run("get policies success", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		policyJSON := `{"provider":"OpenAI","model":"gpt-5","max_usd_per_task":2.5,"retries":2,"requires_approval":true}`
		rows := sqlmock.NewRows([]string{"policy_id", "tenant_id", "workspace_id", "family", "version", "policy_json"}).
			AddRow("pol_1", "ten_01", sql.NullString{String: "", Valid: false}, sql.NullString{String: "provider", Valid: true}, 1, []byte(policyJSON))
		mock.ExpectQuery(regexp.QuoteMeta("SELECT policy_id, tenant_id, workspace_id, family, version, policy_json FROM policy_bundles WHERE tenant_id = $1 ORDER BY family, policy_id")).
			WithArgs("ten_01").WillReturnRows(rows)

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodGet, "/v1/policies", "", actorOperator)
		s.handleGetPoliciesV2(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d", rr.Code)
		}
		if !strings.Contains(rr.Body.String(), "OpenAI") {
			t.Fatalf("expected policy payload in response: %s", rr.Body.String())
		}
	})

	t.Run("create policy success", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		mock.ExpectExec(regexp.QuoteMeta("INSERT INTO policy_bundles (policy_id, tenant_id, workspace_id, family, version, policy_json) VALUES ($1,$2,$3,$4,$5,$6)")).
			WithArgs(sqlmock.AnyArg(), "ten_01", sql.NullString{}, "provider", 1, sqlmock.AnyArg()).
			WillReturnResult(sqlmock.NewResult(1, 1))

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/policies", `{"family":"provider","provider":"OpenAI","model":"gpt-5","max_usd_per_task":2.2,"retries":2,"requires_approval":true}`, actorAdmin)
		s.handleCreatePolicyV2(rr, req)

		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d (%s)", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "gpt-5") {
			t.Fatalf("expected model in response: %s", rr.Body.String())
		}
	})

	t.Run("create policy invalid request", func(t *testing.T) {
		s, _, db := newServerWithMock(t)
		defer db.Close()

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPost, "/v1/policies", `{"provider":"OpenAI"}`, actorAdmin)
		s.handleCreatePolicyV2(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Fatalf("expected 400 got %d", rr.Code)
		}
	})

	t.Run("patch policy not found", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		mock.ExpectQuery(regexp.QuoteMeta("SELECT policy_id, tenant_id, workspace_id, family, version, policy_json FROM policy_bundles WHERE policy_id = $1 AND tenant_id = $2")).
			WithArgs("pol_missing", "ten_01").WillReturnError(sql.ErrNoRows)

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPatch, "/v1/policies/pol_missing", `{"version":2}`, actorAdmin)
		req.SetPathValue("policy_id", "pol_missing")
		s.handlePatchPolicyV2(rr, req)
		if rr.Code != http.StatusNotFound {
			t.Fatalf("expected 404 got %d", rr.Code)
		}
	})

	t.Run("patch policy success", func(t *testing.T) {
		s, mock, db := newServerWithMock(t)
		defer db.Close()

		currentJSON := []byte(`{"provider":"OpenAI","model":"gpt-5","max_usd_per_task":2.2,"retries":2,"requires_approval":true}`)
		mock.ExpectQuery(regexp.QuoteMeta("SELECT policy_id, tenant_id, workspace_id, family, version, policy_json FROM policy_bundles WHERE policy_id = $1 AND tenant_id = $2")).
			WithArgs("pol_1", "ten_01").
			WillReturnRows(sqlmock.NewRows([]string{"policy_id", "tenant_id", "workspace_id", "family", "version", "policy_json"}).
				AddRow("pol_1", "ten_01", sql.NullString{}, sql.NullString{String: "provider", Valid: true}, 1, currentJSON))
		mock.ExpectExec(regexp.QuoteMeta("UPDATE policy_bundles SET workspace_id = $1, family = $2, version = $3, policy_json = $4 WHERE policy_id = $5 AND tenant_id = $6")).
			WithArgs(sql.NullString{}, sql.NullString{String: "provider", Valid: true}, 2, sqlmock.AnyArg(), "pol_1", "ten_01").
			WillReturnResult(sqlmock.NewResult(1, 1))

		rr := httptest.NewRecorder()
		req := reqWithActor(http.MethodPatch, "/v1/policies/pol_1", `{"version":2,"model":"gpt-5.1"}`, actorAdmin)
		req.SetPathValue("policy_id", "pol_1")
		s.handlePatchPolicyV2(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected 200 got %d (%s)", rr.Code, rr.Body.String())
		}
		if !strings.Contains(rr.Body.String(), "gpt-5.1") {
			t.Fatalf("expected patched model in response: %s", rr.Body.String())
		}
	})
}
