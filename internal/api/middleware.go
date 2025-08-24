package api

import (
	"context"
	"net/http"
	"strings"

	"github.com/justanamir/medappoint/internal/auth"
	"github.com/justanamir/medappoint/internal/config"
)

type ctxKey string

const (
	ctxUserID ctxKey = "uid"
	ctxRole   ctxKey = "role"
)

// WithAuth validates "Authorization: Bearer <jwt>" and attaches user to context.
func WithAuth(cfg config.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := r.Header.Get("Authorization")
			if h == "" || !strings.HasPrefix(strings.ToLower(h), "bearer ") {
				ErrorJSON(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}
			token := strings.TrimSpace(h[len("Bearer "):])

			claims, err := auth.ParseJWT(cfg.JWTSecret, token)
			if err != nil {
				ErrorJSON(w, http.StatusUnauthorized, "invalid token", nil)
				return
			}

			ctx := context.WithValue(r.Context(), ctxUserID, claims.UserID)
			ctx = context.WithValue(ctx, ctxRole, claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Helpers
func UserIDFromCtx(r *http.Request) (int64, bool) {
	v := r.Context().Value(ctxUserID)
	id, ok := v.(int64)
	return id, ok
}
func RoleFromCtx(r *http.Request) (string, bool) {
	v := r.Context().Value(ctxRole)
	s, ok := v.(string)
	return s, ok
}
