package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"panzucha/internal/domain"
	"panzucha/internal/httputil"
	"panzucha/internal/services"
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
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "get_product: failed", "err", err, "product_id", id)
		}
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
		slog.ErrorContext(r.Context(), "list_products: failed", "err", err, "limit", limit, "offset", offset)
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toProductListResponse(products, limit, offset))
}

func (h *ProductHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "create_product: invalid json", "err", err)
		httputil.RespondError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "create_product: validation failed", "err", err)
		httputil.RespondError(w, err)
		return
	}

	p := toDomainProduct(req)
	p.ID = uuid.NewString()

	if err := h.svc.Create(r.Context(), &p); err != nil {
		// Log safe fields only — never log the full payload which may contain
		// sensitive pricing or inventory data not meant for log storage.
		slog.ErrorContext(r.Context(), "create_product: failed", "err", err, "product_name", req.Name)
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "create_product: success", "product_id", p.ID)
	httputil.RespondJSON(w, http.StatusCreated, toProductResponse(p))
}

func (h *ProductHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req updateProductRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "update_product: invalid json", "err", err, "product_id", id)
		httputil.RespondError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "update_product: validation failed", "err", err, "product_id", id)
		httputil.RespondError(w, err)
		return
	}

	existing, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "update_product: fetch failed", "err", err, "product_id", id)
		}
		httputil.RespondError(w, err)
		return
	}

	updated := applyProductUpdate(*existing, req)
	if err := h.svc.Update(r.Context(), &updated); err != nil {
		slog.ErrorContext(r.Context(), "update_product: failed", "err", err, "product_id", id)
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "update_product: success", "product_id", id)
	httputil.RespondJSON(w, http.StatusOK, toProductResponse(updated))
}

func (h *ProductHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.svc.Delete(r.Context(), id); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "delete_product: failed", "err", err, "product_id", id)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "delete_product: success", "product_id", id)
	w.WriteHeader(http.StatusNoContent)
}
