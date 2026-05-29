package middleware

import (
	"log/slog"
	"net/http"
	"panzucha/internal/auth"
	"time"

	"github.com/go-chi/chi/v5/middleware"
)

func StructuredLogger(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			reqID := middleware.GetReqID(r.Context())

			// Extract real client IP (reuse your ClientIP middleware logic)
			clientIP := r.RemoteAddr
			if realIP := r.Header.Get("X-Real-IP"); realIP != "" {
				clientIP = realIP
			}

			// Redact sensitive query params (optional but recommended)
			path := r.URL.Path
			if r.URL.RawQuery != "" {
				// Log query presence but not content (PII risk)
				path += "?[query_params_redacted]"
			}

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				attrs := []slog.Attr{
					slog.String("request_id", reqID),
					slog.String("method", r.Method),
					slog.String("path", path),
					slog.String("client_ip", clientIP),
					slog.String("user_agent", r.UserAgent()),
					slog.Int("status", ww.Status()),
					slog.Duration("latency", time.Since(start)),
					slog.Int("bytes_written", ww.BytesWritten()),
				}

				// Pull user ID from auth context if present
				if userID, ok := auth.UserIDFromContext(r.Context()); ok {
					attrs = append(attrs, slog.String("user_id", userID))
				}

				level := slog.LevelInfo
				switch {
				case ww.Status() >= 500:
					level = slog.LevelError
				case ww.Status() >= 400:
					level = slog.LevelWarn
				}

				logger.LogAttrs(r.Context(), level, "http_request", attrs...)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}
