package auth

import (
	"context"
	"net/http"
	"strings"
)

type Actor struct {
	Type     string
	ID       string
	TenantID string
}

type ctxKey string

const actorCtxKey ctxKey = "actor"

func WithActor(ctx context.Context, actor Actor) context.Context {
	return context.WithValue(ctx, actorCtxKey, actor)
}

func ActorFromContext(ctx context.Context) (Actor, bool) {
	actor, ok := ctx.Value(actorCtxKey).(Actor)
	return actor, ok
}

func RequireHuman(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			http.Error(w, "missing clerk bearer token", http.StatusUnauthorized)
			return
		}
		ctx := WithActor(r.Context(), Actor{Type: "human", ID: "clerk_stub_user", TenantID: "tenant_stub"})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func RequireNode(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token := bearerToken(r)
		if token == "" {
			http.Error(w, "missing node bearer token", http.StatusUnauthorized)
			return
		}
		ctx := WithActor(r.Context(), Actor{Type: "node", ID: "node_stub", TenantID: "tenant_stub"})
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func bearerToken(r *http.Request) string {
	authz := r.Header.Get("Authorization")
	if authz == "" {
		return ""
	}
	parts := strings.SplitN(authz, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}
