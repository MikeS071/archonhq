package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseTokenDefaultsAndBearerParsing(t *testing.T) {
	a, err := ParseToken("human:ten_01:user_1")
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if !a.HasRole("operator") {
		t.Fatalf("expected default operator role")
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Token abc")
	if token := bearerToken(req); token != "" {
		t.Fatalf("expected empty token for invalid scheme")
	}
}

func TestMiddlewareTypeMismatch(t *testing.T) {
	nodeProtected := RequireNode(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	reqHuman := httptest.NewRequest(http.MethodGet, "/", nil)
	reqHuman.Header.Set("Authorization", "Bearer human:ten_01:user_1:operator")
	rr := httptest.NewRecorder()
	nodeProtected.ServeHTTP(rr, reqHuman)
	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rr.Code)
	}
}
