package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-playground/validator/v10"

	shareddomain "panzucha/internal/shared/domain"
)

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func RespondJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("failed to encode JSON response", "err", err)
		http.Error(w,
			`{"code":"INTERNAL_ERROR","message":"internal encoding error"}`,
			http.StatusInternalServerError,
		)
	}
}

// RespondError maps any error to a structured JSON response.
func RespondError(w http.ResponseWriter, err error) {
	// Step 1 — validator.ValidationErrors are not domain errors.
	// errors.As unwraps the chain, so this works even if the validator error
	// is wrapped with fmt.Errorf("%w", ...) somewhere upstream.
	if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			fields[fe.Field()] = translateValidatorTag(fe.Tag())
		}
		RespondJSON(w, http.StatusUnprocessableEntity, APIError{
			Code:    "VALIDATION_FAILED",
			Message: "invalid input parameters",
			Details: fields,
		})
		return
	}

	// Step 2 — domain sentinel errors.
	status, code, msg, details := mapDomainError(err)
	RespondJSON(w, status, APIError{
		Code:    code,
		Message: msg,
		Details: details,
	})
}

// mapDomainError translates domain sentinel errors into HTTP status codes.
// validator.ValidationErrors are handled upstream in RespondError —
// they never reach this function.
func mapDomainError(err error) (status int, code string, msg string, details any) {
	switch {
	case errors.Is(err, shareddomain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "resource not found", nil

	case errors.Is(err, shareddomain.ErrConflict):
		return http.StatusConflict, "CONFLICT", "resource conflict", nil

	case errors.Is(err, shareddomain.ErrVersionConflict):
		return http.StatusConflict, "VERSION_CONFLICT", "concurrent update detected, please retry", nil

	case errors.Is(err, shareddomain.ErrInsufficientStock):
		return http.StatusConflict, "INSUFFICIENT_STOCK", "insufficient stock for one or more items", nil

	case errors.Is(err, shareddomain.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials", nil

	case errors.Is(err, shareddomain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", "insufficient permissions", nil

	case errors.Is(err, shareddomain.ErrInvalidInput):
		// Business rule violation from the service layer.
		// Never leak err.Error() — may contain internal details.
		return http.StatusBadRequest, "INVALID_INPUT", "invalid request parameters", nil

	default:
		// Unknown error — already logged in the handler with full context.
		// Generic message so internal details never reach the client.
		return http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error", nil
	}
}

func translateValidatorTag(tag string) string {
	switch tag {
	case "required":
		return "is required"
	case "min":
		return "too short"
	case "max":
		return "too long"
	case "gt":
		return "must be greater than 0"
	case "gte":
		return "must be zero or greater"
	case "email":
		return "must be a valid email"
	case "uuid4":
		return "must be a valid UUID"
	case "dive":
		return "contains invalid items"
	default:
		return "invalid value"
	}
}
