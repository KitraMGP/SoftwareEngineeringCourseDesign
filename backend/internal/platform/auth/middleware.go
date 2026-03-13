package auth

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/platform/httpx"
)

type principalContextKey string

const authPrincipalContextKey principalContextKey = "auth_principal"

func WithPrincipal(ctx context.Context, principal Principal) context.Context {
	return context.WithValue(ctx, authPrincipalContextKey, principal)
}

func PrincipalFromContext(ctx context.Context) (Principal, bool) {
	value, ok := ctx.Value(authPrincipalContextKey).(Principal)
	return value, ok
}

func Middleware(tokens *TokenManager) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if !strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				httpx.Error(w, r, httpx.Unauthorized("missing bearer token"))
				return
			}

			rawToken := strings.TrimSpace(authHeader[len("Bearer "):])
			if rawToken == "" {
				httpx.Error(w, r, httpx.Unauthorized("missing bearer token"))
				return
			}

			principal, err := tokens.ParseAccessToken(rawToken)
			if err != nil {
				httpx.Error(w, r, httpx.Unauthorized("invalid access token"))
				return
			}

			next.ServeHTTP(w, r.WithContext(WithPrincipal(r.Context(), principal)))
		})
	}
}

func RequireRole(role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			principal, ok := PrincipalFromContext(r.Context())
			if !ok {
				httpx.Error(w, r, httpx.Unauthorized("missing auth context"))
				return
			}
			if principal.Role != role {
				httpx.Error(w, r, httpx.Forbidden("insufficient permissions"))
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
