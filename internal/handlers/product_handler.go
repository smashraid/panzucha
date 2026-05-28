package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

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
}

func NewProductHandler(s services.ProductService) *ProductHandler {
	return &ProductHandler{services: s}
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	product := mapper.FromCreateProductRequest(&req)
	if err := h.services.Create(r.Context(), product); err != nil {
		status := http.StatusInternalServerError
		msg := logger.MsgBusinessInternalError
		http.Error(w, msg, status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	httputil.RespondJSON(w, http.StatusCreated, resp)
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	product, err := h.services.GetByID(r.Context(), id)
	if err != nil {
		status := http.StatusNotFound
		if err.Error() != "product not found" {
			status = http.StatusInternalServerError
		}
		http.Error(w, err.Error(), status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	var req dto.UpdateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
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
		http.Error(w, msg, status)
		return
	}

	resp := mapper.FromDomainToResponse(product)
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if id == "" {
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
		http.Error(w, msg, status)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit := 10
	offset := 0
	if lStr := r.URL.Query().Get("limit"); lStr != "" {
		if l, err := strconv.Atoi(lStr); err == nil && l > 0 {
			limit = l
		}
	}
	if oStr := r.URL.Query().Get("offset"); oStr != "" {
		if o, err := strconv.Atoi(oStr); err == nil && o >= 0 {
			offset = o
		}
	}

	products, err := h.services.List(r.Context(), limit, offset)
	if err != nil {
		http.Error(w, logger.MsgBusinessInternalError, http.StatusInternalServerError)
		return
	}

	resp := make([]dto.ProductResponse, len(products))
	for i, p := range products {
		resp[i] = *mapper.FromDomainToResponse(&p)
	}

	httputil.RespondJSON(w, http.StatusOK, resp)
}
