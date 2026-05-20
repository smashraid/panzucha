package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"panzucha/internal/handlers/dto"
	"panzucha/internal/handlers/mapper"
	"panzucha/internal/httputil"
	"panzucha/internal/logger"
	"panzucha/internal/services"
	"panzucha/internal/validation"

	"github.com/go-chi/chi/v5"
)

type ProductHandler struct {
	services services.ProductService
	logger   *logger.Logger
}

func NewProductHandler(s services.ProductService, log *logger.Logger) *ProductHandler {
	return &ProductHandler{services: s, logger: log}
}

// Create handles POST /products
func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)

	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    logger.MsgBusinessInvalidJSON,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    logger.MsgBusinessValidationFailed,
			Payload:    req,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product := mapper.FromCreateProductRequest(&req)
	if err := h.services.Create(r.Context(), product); err != nil {
		status := http.StatusInternalServerError
		msg := logger.MsgBusinessInternalError
		// Could inspect error type (e.g., conflict, bad request) but keep simple for now
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    msg,
			Payload:    req,
		})
		http.Error(w, msg, status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusCreated,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     info.UserID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessCreated,
		Payload:    req,
	})
	httputil.RespondJSON(w, http.StatusCreated, resp)
}

// GetByID handles GET /products/{id}
func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessInvalidIdentifier,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	product, err := h.services.GetByID(r.Context(), id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() != "product not found" {
			status = http.StatusInternalServerError
		}
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    err.Error(),
			Payload:    nil,
		})
		http.Error(w, err.Error(), status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     info.UserID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessRetrieved,
		Payload:    nil,
	})
	httputil.RespondJSON(w, http.StatusOK, resp)
}

// Update handles PUT /products/{id}
func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessInvalidIdentifier,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    logger.MsgBusinessInvalidJSON,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    logger.MsgBusinessValidationFailed,
			Payload:    req,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product := mapper.FromUpdateProductRequest(&req)
	product.ID = id
	if err := h.services.Update(r.Context(), product); err != nil {
		status := http.StatusInternalServerError
		msg := logger.MsgBusinessInternalError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
			msg = logger.MsgBusinessNotFound
		}
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    msg,
			Payload:    req,
		})
		http.Error(w, msg, status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     info.UserID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessUpdated,
		Payload:    req,
	})
	httputil.RespondJSON(w, http.StatusOK, resp)
}

// Delete handles DELETE /products/{id}
func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	id := chi.URLParam(r, "id")
	if id == "" {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusBadRequest,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessInvalidIdentifier,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	if err := h.services.Delete(r.Context(), id); err != nil {
		status := http.StatusInternalServerError
		msg := logger.MsgBusinessInternalError
		if err.Error() == "product not found" {
			status = http.StatusNotFound
			msg = logger.MsgBusinessNotFound
		}
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    msg,
			Payload:    nil,
		})
		http.Error(w, msg, status)
		return
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusNoContent,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     info.UserID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessDeleted,
		Payload:    nil,
	})
	w.WriteHeader(http.StatusNoContent)
}

// List handles GET /products
func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)

	products, err := h.services.List(r.Context())
	if err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusInternalServerError,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    logger.MsgBusinessListFailed,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessInternalError, http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		resp[i] = *mapper.FromDomainToResponse(&p)
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     info.UserID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessListed,
		Payload:    nil,
	})
	httputil.RespondJSON(w, http.StatusOK, resp)
}
