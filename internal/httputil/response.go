package httputil

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"panzucha/internal/domain"

	"github.com/go-playground/validator/v10"
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
		http.Error(w, `{"code":"INTERNAL_ERROR","message":"internal encoding error"}`, http.StatusInternalServerError)
	}
}

func RespondError(w http.ResponseWriter, err error) {
	status, code, msg, details := mapDomainError(err)
	RespondJSON(w, status, APIError{Code: code, Message: msg, Details: details})
}

func mapDomainError(err error) (int, string, string, any) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		return http.StatusNotFound, "NOT_FOUND", "resource not found", nil
	case errors.Is(err, domain.ErrConflict):
		return http.StatusConflict, "CONFLICT", "resource conflict", nil
	case errors.Is(err, domain.ErrVersionConflict):
		return http.StatusConflict, "CONFLICT", "concurrent update detected, please retry", nil
	case errors.Is(err, domain.ErrUnauthorized):
		return http.StatusUnauthorized, "UNAUTHORIZED", "invalid credentials", nil
	case errors.Is(err, domain.ErrForbidden):
		return http.StatusForbidden, "FORBIDDEN", "insufficient permissions", nil
	case errors.Is(err, domain.ErrInsufficientStock):
		return http.StatusConflict, "INSUFFICIENT_STOCK", "insufficient stock for one or more items", nil
	case errors.Is(err, domain.ErrInvalidInput):
		if ve, ok := errors.AsType[validator.ValidationErrors](err); ok {
			fields := make(map[string]string, len(ve))
			for _, fe := range ve {
				fields[fe.Field()] = translateValidatorTag(fe.Tag())
			}
			return http.StatusUnprocessableEntity, "VALIDATION_FAILED", "invalid input parameters", fields
		}
		// Non-validator invalid input → sanitize. Never leak err.Error()
		return http.StatusBadRequest, "INVALID_INPUT", "invalid request parameters", nil
	default:
		// Unexpected error → log in handler, return generic 500
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
	case "email":
		return "must be a valid email"
	default:
		return "invalid value"
	}
}
