package logger

type Category string
type SubCategory string
type StatusCategory string
type PerformanceBucket string

const (
	// Categories
	CategoryAPI      Category = "api"
	CategoryDatabase Category = "database"
	CategoryBusiness Category = "business"
	CategoryAuth     Category = "auth"
	CategorySystem   Category = "system"

	// SubCategories - API
	SubAPIRequest  SubCategory = "http_request"
	SubAPIResponse SubCategory = "http_response"

	// SubCategories - Database
	SubDBSelect SubCategory = "db_select"
	SubDBInsert SubCategory = "db_insert"
	SubDBUpdate SubCategory = "db_update"
	SubDBDelete SubCategory = "db_delete"

	// SubCategories - Business
	SubValidation SubCategory = "validation"
	SubProcess    SubCategory = "process"
)

// StatusCategory from HTTP status code
func GetStatusCategory(statusCode int) string {
	switch {
	case statusCode >= 200 && statusCode < 300:
		return "2xx_Success"
	case statusCode >= 300 && statusCode < 400:
		return "3xx_Redirect"
	case statusCode >= 400 && statusCode < 500:
		return "4xx_ClientError"
	case statusCode >= 500 && statusCode < 600:
		return "5xx_ServerError"
	default:
		return "Unknown"
	}
}

// PerformanceBucket from duration in milliseconds
func GetPerformanceBucket(durationMs int64) string {
	switch {
	case durationMs < 100:
		return "Fast"
	case durationMs < 500:
		return "Normal"
	case durationMs < 1000:
		return "Slow"
	case durationMs < 5000:
		return "VerySlow"
	default:
		return "Critical"
	}
}

// Database operation constants
const (
	DBSelect = "db_select"
	DBInsert = "db_insert"
	DBUpdate = "db_update"
	DBDelete = "db_delete"
)

// API message constants (already discussed)
const (
	MsgBusinessInvalidJSON          = "invalid JSON"
	MsgBusinessValidationFailed     = "validation failed"
	MsgBusinessCreated              = "resource created"
	MsgBusinessUpdated              = "resource updated"
	MsgBusinessDeleted              = "resource deleted"
	MsgBusinessRetrieved            = "resource retrieved"
	MsgBusinessListed               = "resources listed"
	MsgBusinessInternalError        = "internal server error"
	MsgBusinessUnauthorized         = "unauthorized access"
	MsgBusinessForbidden            = "access forbidden"
	MsgBusinessCreateFailed         = "failed to create resource"
	MsgBusinessUpdateFailed         = "failed to update resource"
	MsgBusinessDeleteFailed         = "failed to delete resource"
	MsgBusinessGetFailed            = "failed to get resource"
	MsgBusinessListFailed           = "failed to list resources"
	MsgBusinessNotFound             = "resource not found"
	MsgBusinessDatabaseError        = "database error"
	MsgBusinessSuccess              = "operation completed successfully"
	MsgBusinessInvalidIdentifier    = "invalid resource identifier"
	MsgIdempotencyNotFound          = "missing idempotency key"
	MsgIdempotencyValidationFailed  = "idempotency check failed"
	MsgIdempotencyReservationFailed = "idempotency reservation failed"
	MsgIdempotencyCreateFailed      = "failed to complete idempotency key"
	MsgIdempotencyProcessing        = "request already in progress"
)

// Business sub‑category constants (optional but recommended)
const (
	BusinessEntityCreate = "entity_creation"
	BusinessEntityUpdate = "entity_update"
	BusinessEntityDelete = "entity_delete"
	BusinessEntityGet    = "entity_get"
	BusinessEntityList   = "entity_list"
	BusinessExistEntity  = "entity_exist"
	// similar for user, order, etc.
)
