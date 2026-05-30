package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"panzucha/internal/httputil"
	"regexp"
	"strings"
)

const (
	idempotencyKeyCtxKey contextKey = "idempotency_key"
)

// Regex: UUID-like or alphanumeric with dashes/underscores, 36-128 chars
var idempotencyKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_]{36,128}$`)

func RequireIdempotencyKey(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSpace(r.Header.Get("Idempotency-Key"))

		if key == "" || !idempotencyKeyPattern.MatchString(key) {
			slog.WarnContext(r.Context(), "invalid idempotency key",
				"key_preview", keyPreview(key),
				"path", r.URL.Path,
				"method", r.Method,
			)
			// Direct HTTP response — no domain error needed
			httputil.RespondJSON(w, http.StatusBadRequest, map[string]string{
				"error": "Idempotency-Key header is required and must be 36-128 alphanumeric characters",
			})
			return
		}

		// Inject into context for downstream handlers
		ctx := context.WithValue(r.Context(), idempotencyKeyCtxKey, key)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// GetIdempotencyKey extracts the validated key from context.
// Returns ("", false) if not set — handler should treat as internal error.
func GetIdempotencyKey(ctx context.Context) (string, bool) {
	key, ok := ctx.Value(idempotencyKeyCtxKey).(string)
	return key, ok
}

// keyPreview returns a safe, truncated version for logging.
// - Empty key → "<missing>"
// - Short key → "<invalid>"
// - Valid key → first 8 chars + "..."
func keyPreview(key string) string {
	if key == "" {
		return "<missing>"
	}
	key = strings.TrimSpace(key)
	if len(key) < 8 {
		return "<invalid>"
	}
	return key[:8] + "..."
}
