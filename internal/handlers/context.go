package handlers

import (
	"net/http"
	"panzucha/internal/auth"
	"panzucha/internal/middleware"
	"time"
)

type RequestInfo struct {
	StartTime time.Time
	RequestID string
	ClientIP  string
	UserAgent string
	UserID    string
}

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
