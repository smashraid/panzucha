package logger

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"go.opentelemetry.io/otel/trace"
)

type APILogParams struct {
	Ctx        context.Context
	Method     string
	Path       string
	StatusCode int
	Duration   time.Duration
	RequestID  string
	UserID     string
	ClientIP   string
	UserAgent  string
	Err        error
	Message    string
	Payload    any
	Custom     map[string]any
}

type DBLogParams struct {
	Ctx          context.Context
	Operation    string
	Table        string
	Duration     time.Duration
	RowsAffected int64
	Err          error
	Custom       map[string]any
}

type BusinessLogParams struct {
	Ctx         context.Context
	SubCategory string
	EntityType  string
	EntityID    string
	Message     string
	Err         error
}

func (l *Logger) log(ctx context.Context, level string, entry LogEntry) {
	// Set defaults
	entry.Timestamp = time.Now().UTC().Format(time.RFC3339Nano)
	entry.Level = level
	entry.Logger = l.service

	// Send to stdout (structured)
	l.slog.Log(ctx, parseLevel(level), entry.Message, "entry", entry)

	// Async send to Logstash
	// if l.sender != nil {
	// 	l.sender.Send(entry)
	// }
}

// LogAPI records an HTTP API request/response.
// If err is not nil, the log is marked as ERROR and includes error details + stack trace.
func (l *Logger) LogAPI(params APILogParams) {
	ctx := params.Ctx
	spanCtx := trace.SpanContextFromContext(ctx)
	if spanCtx.IsValid() {
		if params.Custom == nil {
			params.Custom = make(map[string]any)
		}
		params.Custom["trace_id"] = spanCtx.TraceID().String()
		params.Custom["span_id"] = spanCtx.SpanID().String()
	}
	ms := params.Duration.Milliseconds()
	entry := LogEntry{
		Category:    string(CategoryAPI),
		SubCategory: string(SubAPIResponse),
		HTTPMethod:  params.Method,
		HTTPPath:    params.Path,
		HTTPStatus:  params.StatusCode,
		StatusCat:   GetStatusCategory(params.StatusCode),
		DurationMs:  ms,
		Performance: GetPerformanceBucket(ms),
		RequestID:   params.RequestID,
		UserID:      params.UserID,
		ClientIP:    params.ClientIP,
		UserAgent:   params.UserAgent,
		Custom:      params.Custom,
	}

	level := "INFO"
	if params.StatusCode >= 400 || params.Err != nil {
		level = "ERROR"
		entry.Message = params.Message
		if params.Err != nil {
			entry.Error = params.Err.Error()
			// Capture stack trace for internal errors (500)
			if params.StatusCode >= 500 {
				// Clean up the stack (remove logger internal frames)
				stack := string(debug.Stack())
				lines := strings.Split(stack, "\n")
				// Keep only frames after the logger call (simplified)
				entry.ErrorStack = strings.Join(lines, "\n")
			}
		}
	} else {
		entry.Message = params.Message
	}

	if params.Payload != nil && l.env == "development" { // only in dev, or based on header
		entry.RequestPayload = params.Payload
	}

	l.log(ctx, level, entry)
}

// Database Logging Helper
func (l *Logger) LogDB(params DBLogParams) {
	ms := params.Duration.Milliseconds()
	entry := LogEntry{
		Category:       string(CategoryDatabase),
		SubCategory:    params.Operation,
		DBOperation:    params.Operation,
		DBTable:        params.Table,
		DurationMs:     ms,
		Performance:    GetPerformanceBucket(ms),
		DBRowsAffected: params.RowsAffected,
		Message:        "Database operation completed",
	}

	level := "INFO"

	if params.Custom != nil {
		entry.Custom = params.Custom
	}

	if params.Err != nil {
		level = "ERROR"
		entry.Message = "Database operation failed"
		entry.Error = params.Err.Error()
	}
	l.log(params.Ctx, level, entry)
}

// Business Logging Helper
func (l *Logger) LogBusiness(params BusinessLogParams) {
	entry := LogEntry{
		Category:    string(CategoryBusiness),
		SubCategory: params.SubCategory,
		EntityType:  params.EntityType,
		EntityID:    params.EntityID,
		Message:     params.Message,
	}

	level := "INFO"
	if params.Err != nil {
		level = "ERROR"
		entry.Error = params.Err.Error()
	}
	l.log(params.Ctx, level, entry)
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
