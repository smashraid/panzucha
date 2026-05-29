package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"panzucha/internal/httputil"
	"panzucha/internal/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ProductHandler struct {
	svc      services.ProductService
	validate *validator.Validate
}

func NewProductHandler(s services.ProductService, v *validator.Validate) *ProductHandler {
	return &ProductHandler{svc: s, validate: v}
}

func (h *ProductHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	product, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toProductResponse(*product))
}

func (h *ProductHandler) List(w http.ResponseWriter, r *http.Request) {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	products, err := h.svc.List(r.Context(), limit, offset)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toProductListResponse(products, limit, offset))
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorBody("invalid JSON"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.RespondJSON(w, http.StatusUnprocessableEntity, httputil.ValidationErrorBody(err))
		return
	}

	// Convert DTO → domain entity. The handler tills business fields;
	// the service assigns ID and persists.
	p := toDomainProduct(req)
	p.ID = uuid.NewString()

	if err := h.svc.Create(r.Context(), &p); err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, toProductResponse(p))
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorBody("invalid JSON"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.RespondJSON(w, http.StatusUnprocessableEntity, httputil.ValidationErrorBody(err))
		return
	}

	// Fetch current state, apply patch, persist.
	existing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	updated := applyProductUpdate(*existing, req)
	if err := h.svc.Update(r.Context(), &updated); err != nil {
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toProductResponse(updated))
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		httputil.RespondError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
