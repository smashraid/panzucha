package handlers

import (
	"encoding/json"
	"net/http"
	"panzucha/internal/handlers/dto"
	"panzucha/internal/handlers/mapper"
	"panzucha/internal/logger"
	"panzucha/internal/services"
	"time"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

type ProductHandler struct {
	services services.ProductService
	logger   *logger.Logger
}

func NewProductHandler(s services.ProductService, log *logger.Logger) *ProductHandler {
	return &ProductHandler{services: s, logger: log}
}

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
			Message:    "invalid JSON",
			Payload:    nil,
		})
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if err := validate.Struct(req); err != nil {
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
			Message:    err.Error(),
			Payload:    req,
		})
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// 3. Map to domain model
	product := mapper.FromCreateProductRequest(&req)

	// 4. Call service (which uses repository – still domain model)
	if err := h.services.Create(r.Context(), product); err != nil {
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
			Message:    err.Error(),
			Payload:    req,
		})
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 5. Map domain model to response DTO
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
		Message:    "product created",
		Payload:    req,
	})

	// 6. Send JSON response
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}
