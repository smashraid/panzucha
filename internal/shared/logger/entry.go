package logger

type LogEntry struct {
	// Standard fields
	Timestamp string `json:"@timestamp"`
	Level     string `json:"level"`
	Message   string `json:"message"`
	Logger    string `json:"logger"`

	// Core categorization
	Category    string `json:"category"`
	SubCategory string `json:"sub_category"`

	// Performance
	DurationMs  int64  `json:"duration_ms,omitempty"`
	Performance string `json:"performance,omitempty"`

	// HTTP specific
	HTTPMethod string `json:"http_method,omitempty"`
	HTTPPath   string `json:"http_path,omitempty"`
	HTTPStatus int    `json:"http_status,omitempty"`
	StatusCat  string `json:"status_category,omitempty"`
	ClientIP   string `json:"client_ip,omitempty"`
	UserAgent  string `json:"user_agent,omitempty"`
	RequestID  string `json:"request_id,omitempty"`
	UserID     string `json:"user_id,omitempty"`

	// Database specific
	DBOperation    string `json:"db_operation,omitempty"`
	DBTable        string `json:"db_table,omitempty"`
	DBRowsAffected int64  `json:"db_rows_affected,omitempty"`

	// Business context
	EntityID   string `json:"entity_id,omitempty"`
	EntityType string `json:"entity_type,omitempty"`

	// Error fields
	Error      string `json:"error,omitempty"`
	ErrorStack string `json:"error_stack,omitempty"`

	RequestPayload any `json:"request_payload,omitempty"`
	// Additional custom fields
	Custom map[string]any `json:"custom,omitempty"`
}
