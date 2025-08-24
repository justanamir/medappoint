package api

import (
	"net/http"
	"runtime/debug"
)

// RecoverJSON wraps handlers and converts panics into JSON 500 errors.
func RecoverJSON(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				// optional: log stack trace server-side; here we return a clean JSON
				ErrorJSON(w, http.StatusInternalServerError, "internal server error", nil)
				_ = debug.Stack() // keep to avoid "imported and not used" if you log later
			}
		}()
		next.ServeHTTP(w, r)
	})
}
