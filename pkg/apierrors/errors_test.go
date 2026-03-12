package apierrors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWrite(t *testing.T) {
	rr := httptest.NewRecorder()
	Write(rr, http.StatusForbidden, "forbidden", "nope", "corr_1", map[string]any{"k": "v"})

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403 got %d", rr.Code)
	}

	var resp ErrorResponse
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if resp.Error.Code != "forbidden" || resp.Error.CorrelationID != "corr_1" {
		t.Fatalf("unexpected payload: %+v", resp)
	}
}
