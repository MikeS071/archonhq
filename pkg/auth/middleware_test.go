package auth

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestParseTokenAndRoleHelpers(t *testing.T) {
	human, err := ParseToken("human:ten_01:user_01:tenant_admin,approver")
	if err != nil {
		t.Fatalf("parse human: %v", err)
	}
	if human.Type != "human" || !human.HasRole("tenant_admin") || !human.HasAnyRole("developer", "approver") {
		t.Fatalf("unexpected human actor: %+v", human)
	}

	node, err := ParseToken("node:ten_01:node_01:cred_01")
	if err != nil {
		t.Fatalf("parse node: %v", err)
	}
	if node.Type != "node" || node.CredentialID != "cred_01" {
		t.Fatalf("unexpected node actor: %+v", node)
	}

	if _, err := ParseToken("bad"); err == nil {
		t.Fatalf("expected parse error")
	}
	if _, err := ParseToken("service:ten:user"); err == nil {
		t.Fatalf("expected unsupported type error")
	}
}

func TestMiddlewareGuards(t *testing.T) {
	humanProtected := RequireHuman(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if _, ok := ActorFromContext(r.Context()); !ok {
			t.Fatalf("expected actor in context")
		}
		w.WriteHeader(http.StatusOK)
	}))

	rrUnauthorized := httptest.NewRecorder()
	humanProtected.ServeHTTP(rrUnauthorized, httptest.NewRequest(http.MethodGet, "/", nil))
	if rrUnauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 got %d", rrUnauthorized.Code)
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer human:ten_01:user_01:operator")
	rr := httptest.NewRecorder()
	humanProtected.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", rr.Code)
	}

	nodeProtected := RequireNode(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	nodeReq := httptest.NewRequest(http.MethodGet, "/", nil)
	nodeReq.Header.Set("Authorization", "Bearer node:ten_01:node_01:cred_01")
	nodeRR := httptest.NewRecorder()
	nodeProtected.ServeHTTP(nodeRR, nodeReq)
	if nodeRR.Code != http.StatusOK {
		t.Fatalf("expected 200 got %d", nodeRR.Code)
	}
}
