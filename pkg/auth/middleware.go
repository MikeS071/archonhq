package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"
)

type Actor struct {
	Type         string
	ID           string
	TenantID     string
	CredentialID string
	Roles        map[string]struct{}
	TokenRaw     string
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
		actor, err := ParseToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if actor.Type != "human" {
			http.Error(w, "human token required", http.StatusUnauthorized)
			return
		}
		ctx := WithActor(r.Context(), actor)
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
		actor, err := ParseToken(token)
		if err != nil {
			http.Error(w, err.Error(), http.StatusUnauthorized)
			return
		}
		if actor.Type != "node" {
			http.Error(w, "node token required", http.StatusUnauthorized)
			return
		}
		ctx := WithActor(r.Context(), actor)
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

func ParseToken(token string) (Actor, error) {
	parts := strings.Split(token, ":")
	if len(parts) < 3 {
		return Actor{}, fmt.Errorf("invalid bearer token format")
	}

	switch parts[0] {
	case "human":
		actor := Actor{
			Type:     "human",
			TenantID: parts[1],
			ID:       parts[2],
			Roles:    map[string]struct{}{},
			TokenRaw: token,
		}
		if len(parts) >= 4 {
			for _, role := range strings.Split(parts[3], ",") {
				role = strings.TrimSpace(role)
				if role != "" {
					actor.Roles[role] = struct{}{}
				}
			}
		}
		if len(actor.Roles) == 0 {
			actor.Roles["operator"] = struct{}{}
		}
		return actor, nil
	case "node":
		actor := Actor{
			Type:     "node",
			TenantID: parts[1],
			ID:       parts[2],
			Roles:    map[string]struct{}{},
			TokenRaw: token,
		}
		if len(parts) >= 4 {
			actor.CredentialID = parts[3]
		}
		return actor, nil
	default:
		return Actor{}, fmt.Errorf("unsupported bearer token type")
	}
}

func (a Actor) HasRole(role string) bool {
	_, ok := a.Roles[role]
	return ok
}

func (a Actor) HasAnyRole(roles ...string) bool {
	for _, role := range roles {
		if a.HasRole(role) {
			return true
		}
	}
	return false
}
