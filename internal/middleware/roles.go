package middleware

import (
	"net/http"
	"panzucha/internal/auth"
)

func RequireRole(requiredRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			roles, ok := auth.RolesFromContext(r.Context())
			if !ok {
				http.Error(w, "forbidden", http.StatusForbidden)
				return
			}
			// Check if user has at least one of the required roles
			for _, rl := range roles {
				for _, allowed := range requiredRoles {
					if rl == allowed {
						next.ServeHTTP(w, r)
						return
					}
				}
			}
			http.Error(w, "insufficient permissions", http.StatusForbidden)
		})
	}
}
