package handlers

import (
	"errors"
	"net/http"
	"panzucha/internal/auth"
	"panzucha/internal/middleware"
	"regexp"
	"strings"
	"time"
)

type RequestInfo struct {
	StartTime time.Time
	RequestID string
	ClientIP  string
	UserAgent string
	UserID    string
}

var (
	ErrMissingIdempotencyKey = errors.New("idempotency key required")
	ErrInvalidIdempotencyKey = errors.New("idempotency key format invalid")
	// UUID-like or alphanumeric with dashes: 36-128 chars
	idempotencyKeyPattern = regexp.MustCompile(`^[a-zA-Z0-9\-_]{36,128}$`)
)

func ExtractRequestInfo(r *http.Request) *RequestInfo {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())

	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}
	userAgent := r.UserAgent()

	userID, _ := auth.UserIDFromContext(r.Context())

	return &RequestInfo{
		StartTime: start,
		RequestID: requestID,
		ClientIP:  clientIP,
		UserAgent: userAgent,
		UserID:    userID,
	}
}

func ExtractIdempotencyKey(r *http.Request) (string, error) {
	key := r.Header.Get("Idempotency-Key")
	if key == "" {
		return "", ErrMissingIdempotencyKey
	}

	// Trim whitespace (clients sometimes send "key " with trailing space)
	key = strings.TrimSpace(key)

	// Validate format: prevent injection, overly long keys, or empty after trim
	if !idempotencyKeyPattern.MatchString(key) {
		return "", ErrInvalidIdempotencyKey
	}

	return key, nil
}
