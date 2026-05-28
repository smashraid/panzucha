package httputil

import (
	"encoding/json"
	"errors"
	"net/http"
	"panzucha/internal/domain"

	"github.com/go-playground/validator/v10"
)

// RespondJSON writes a JSON response with the given status code and data.
// func RespondJSON(w http.ResponseWriter, status int, data any) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(status)
// 	if data != nil {
// 		json.NewEncoder(w).Encode(data)
// 	}
// }

// RespondRaw writes a raw response body with the given status code and content type.
func RespondRaw(w http.ResponseWriter, statusCode int, body []byte) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	w.Write(body)
}

func RespondJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}

func RespondError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, domain.ErrNotFound):
		RespondJSON(w, http.StatusNotFound, ErrorBody("resource not found"))
	case errors.Is(err, domain.ErrConflict):
		RespondJSON(w, http.StatusConflict, ErrorBody("conflict"))
	case errors.Is(err, domain.ErrVersionConflict):
		RespondJSON(w, http.StatusConflict, ErrorBody("concurrent update detected, please retry"))
	case errors.Is(err, domain.ErrUnauthorized):
		RespondJSON(w, http.StatusUnauthorized, ErrorBody("unauthorized"))
	case errors.Is(err, domain.ErrForbidden):
		RespondJSON(w, http.StatusForbidden, ErrorBody("forbidden"))
	case errors.Is(err, domain.ErrInvalidInput):
		RespondJSON(w, http.StatusUnprocessableEntity, ErrorBody(err.Error()))
	default:
		RespondJSON(w, http.StatusInternalServerError, ErrorBody("internal server error"))
	}
}

type apiError struct {
	Error string `json:"error"`
}

func ErrorBody(msg string) apiError { return apiError{Error: msg} }

func ValidationErrorBody(err error) any {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		fields := make(map[string]string, len(ve))
		for _, fe := range ve {
			fields[fe.Field()] = fe.Tag()
		}
		return map[string]any{"error": "validation failed", "fields": fields}
	}
	return ErrorBody("validation failed")
}
