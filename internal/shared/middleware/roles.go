package middleware

import (
	"net/http"
	"panzucha/internal/shared/auth"
	"panzucha/internal/shared/httputil"
	"slices"
)

func RequireRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := auth.RolesFromContext(r.Context())
			if !ok {
				httputil.RespondJSON(w, http.StatusForbidden, httputil.APIError{
					Code:    "FORBIDDEN",
					Message: "insufficient permissions",
				})
				return
			}
			for _, rl := range roles {
				if slices.Contains(requiredRoles, rl) {
					next.ServeHTTP(w, r)
					return
				}
			}
			httputil.RespondJSON(w, http.StatusForbidden, httputil.APIError{
				Code:    "FORBIDDEN",
				Message: "insufficient permissions",
			})
		})
	}
}
