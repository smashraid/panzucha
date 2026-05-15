package logger

import (
	"context"
	"log/slog"
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
	l.sender.Send(entry)
}

// API Logging Helper
func (l *Logger) LogAPI(ctx context.Context, method, path string, statusCode int, duration time.Duration, requestID, userID, clientIP, userAgent, message string) {
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
		Message:     message,
	}

	level := "INFO"
	if statusCode >= 400 {
		level = "ERROR"
		entry.Message = "API request failed"
	} else {
		entry.Message = "API request completed"
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
func (l *Logger) LogRequestPayload(ctx context.Context, method, path string, payload interface{}) {
	entry := LogEntry{
		Category:    string(CategoryAPI),
		SubCategory: string(SubAPIRequest),
		HTTPMethod:  method,
		HTTPPath:    path,
		Custom:      map[string]interface{}{"payload": payload},
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
