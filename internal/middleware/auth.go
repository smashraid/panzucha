package middleware

import (
	"log/slog"
	"net/http"
	"panzucha/internal/auth"
	"panzucha/internal/httputil"
	"strings"
)

func Authenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			slog.WarnContext(r.Context(), "auth: missing authorization header",
				"path", r.URL.Path,
			)
			httputil.RespondJSON(w, http.StatusUnauthorized, httputil.APIError{
				Code:    "UNAUTHORIZED",
				Message: "missing authorization header",
			})
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
			slog.WarnContext(r.Context(), "auth: malformed authorization header",
				"path", r.URL.Path,
			)
			httputil.RespondJSON(w, http.StatusUnauthorized, httputil.APIError{
				Code:    "UNAUTHORIZED",
				Message: "invalid authorization header format",
			})
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			slog.WarnContext(r.Context(), "auth: invalid token",
				"path", r.URL.Path,
				"method", r.Method,
			)
			httputil.RespondJSON(w, http.StatusUnauthorized, httputil.APIError{
				Code:    "UNAUTHORIZED",
				Message: "invalid or expired token",
			})
			return
		}

		ctx := auth.ContextWithUser(r.Context(), claims.UserID, claims.Roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
