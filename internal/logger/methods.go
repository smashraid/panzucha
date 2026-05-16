package logger

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"
)

func (l *Logger) log(ctx context.Context, level string, entry LogEntry) {
	// Set defaults
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	entry.Level = level
	entry.Logger = l.service

	// Send to stdout (structured)
	l.slog.Log(ctx, parseLevel(level), entry.Message, "entry", entry)

	// Async send to Logstash
	if l.sender != nil {
		l.sender.Send(entry)
	}
}

// LogAPI records an HTTP API request/response.
// If err is not nil, the log is marked as ERROR and includes error details + stack trace.
func (l *Logger) LogAPI(ctx context.Context, method, path string, statusCode int, duration time.Duration, requestID, userID, clientIP, userAgent string, err error, message string, payload any) {
	ms := duration.Milliseconds()
	entry := LogEntry{
		Category:    string(CategoryAPI),
		SubCategory: string(SubAPIResponse),
		HTTPMethod:  method,
		HTTPPath:    path,
		HTTPStatus:  statusCode,
		StatusCat:   GetStatusCategory(statusCode),
		DurationMs:  ms,
		Performance: GetPerformanceBucket(ms),
		RequestID:   requestID,
		UserID:      userID,
		ClientIP:    clientIP,
		UserAgent:   userAgent,
	}

	level := "INFO"
	if statusCode >= 400 || err != nil {
		level = "ERROR"
		entry.Message = message
		if err != nil {
			entry.Error = err.Error()
			// Capture stack trace for internal errors (500)
			if statusCode >= 500 {
				// Clean up the stack (remove logger internal frames)
				stack := string(debug.Stack())
				lines := strings.Split(stack, "\n")
				// Keep only frames after the logger call (simplified)
				entry.ErrorStack = strings.Join(lines, "\n")
			}
		}
	} else {
		entry.Message = message
	}

	if payload != nil && l.env == "development" { // only in dev, or based on header
		entry.RequestPayload = payload
	}

	l.log(ctx, level, entry)
}

// Database Logging Helper
func (l *Logger) LogDB(ctx context.Context, operation, table string, duration time.Duration, rowsAffected int64, err error) {
	ms := duration.Milliseconds()
	entry := LogEntry{
		Category:       string(CategoryDatabase),
		SubCategory:    operation, // db_select, db_insert, etc.
		DBOperation:    operation,
		DBTable:        table,
		DurationMs:     ms,
		Performance:    GetPerformanceBucket(ms),
		DBRowsAffected: rowsAffected,
	}

	level := "INFO"
	entry.Message = "Database operation completed"
	if err != nil {
		level = "ERROR"
		entry.Message = "Database operation failed"
		entry.Error = err.Error()
	}
	l.log(ctx, level, entry)
}

// Business Logging Helper
func (l *Logger) LogBusiness(ctx context.Context, subCategory, entityType, entityID, message string, err error) {
	entry := LogEntry{
		Category:    string(CategoryBusiness),
		SubCategory: subCategory,
		EntityType:  entityType,
		EntityID:    entityID,
		Message:     message,
	}

	level := "INFO"
	if err != nil {
		level = "ERROR"
		entry.Error = err.Error()
	}
	l.log(ctx, level, entry)
}

// Request/Response Payload Logging (debug only, be careful with sensitive data)
func (l *Logger) LogRequestPayload(ctx context.Context, method, path string, payload any) {
	entry := LogEntry{
		Category:    string(CategoryAPI),
		SubCategory: string(SubAPIRequest),
		HTTPMethod:  method,
		HTTPPath:    path,
		Custom:      map[string]any{"payload": payload},
		Message:     "API request payload",
	}
	l.log(ctx, "DEBUG", entry)
}

func parseLevel(level string) slog.Level {
	switch level {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
