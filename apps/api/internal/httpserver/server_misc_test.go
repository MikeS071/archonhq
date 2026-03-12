package httpserver

import (
	"net/http"
	"net/http/httptest"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
)

func TestServerMiscHandlersAndMiddleware(t *testing.T) {
	dbMock, _, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock new: %v", err)
	}
	defer dbMock.Close()

	eventStore := &inMemoryEventStore{}
	srv := newTestServer(t, dbMock, eventStore)
	h := srv.Handler()

	resultReq := newJSONRequest(t, http.MethodPost, "/v1/results", "node:ten_01:node_01", "idem_result_1", map[string]any{
		"result_id":   "res_01",
		"task_id":     "task_01",
		"lease_id":    "lease_01",
		"output_refs": []string{"art_1"},
		"signature":   "signed:node_01:lease_01:res_01",
	})
	rrResult := httptest.NewRecorder()
	h.ServeHTTP(rrResult, resultReq)
	if rrResult.Code != http.StatusUnauthorized {
		t.Fatalf("submit result expected 401 without credential setup got %d body=%s", rrResult.Code, rrResult.Body.String())
	}

	legacyCreateReq := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_1:operator", "idem_legacy_task_1", map[string]any{
		"workspace_id": "ws_1",
		"task_family":  "research.extract",
		"title":        "legacy handler",
	})
	rrLegacyCreate := httptest.NewRecorder()
	srv.handleCreateTask(rrLegacyCreate, legacyCreateReq)
	if rrLegacyCreate.Code != http.StatusOK {
		t.Fatalf("legacy create task expected 200 got %d body=%s", rrLegacyCreate.Code, rrLegacyCreate.Body.String())
	}

	notImpl := httptest.NewRequest(http.MethodGet, "/v1/pricing/rate-cards", nil)
	rrNotImpl := httptest.NewRecorder()
	h.ServeHTTP(rrNotImpl, notImpl)
	if rrNotImpl.Code != http.StatusNotImplemented {
		t.Fatalf("not implemented route expected 501 got %d body=%s", rrNotImpl.Code, rrNotImpl.Body.String())
	}

	healthReq := httptest.NewRequest(http.MethodGet, "/health", nil)
	rrHealth := httptest.NewRecorder()
	h.ServeHTTP(rrHealth, healthReq)
	if rrHealth.Code != http.StatusServiceUnavailable {
		t.Fatalf("health expected 503 when nats disconnected got %d body=%s", rrHealth.Code, rrHealth.Body.String())
	}

	missingIdem := newJSONRequest(t, http.MethodPost, "/v1/tasks", "human:ten_01:user_1:operator", "placeholder", map[string]any{
		"workspace_id": "ws_1",
		"task_family":  "research.extract",
		"title":        "x",
	})
	missingIdem.Header.Del("Idempotency-Key")
	rrMissingIdem := httptest.NewRecorder()
	h.ServeHTTP(rrMissingIdem, missingIdem)
	if rrMissingIdem.Code != http.StatusBadRequest {
		t.Fatalf("missing idempotency expected 400 got %d body=%s", rrMissingIdem.Code, rrMissingIdem.Body.String())
	}

	rootReq := httptest.NewRequest(http.MethodGet, "/", nil)
	rootReq.Header.Set("X-Correlation-ID", "corr_fixed")
	rrRoot := httptest.NewRecorder()
	h.ServeHTTP(rrRoot, rootReq)
	if rrRoot.Code != http.StatusOK {
		t.Fatalf("root expected 200 got %d", rrRoot.Code)
	}
	if got := rrRoot.Header().Get("X-Correlation-ID"); got != "corr_fixed" {
		t.Fatalf("expected correlation id echo, got %q", got)
	}
}
