package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	"panzucha/internal/contextkeys"
	"panzucha/internal/domain"
	"panzucha/internal/httputil"
	"panzucha/internal/services"
)

type OrderHandler struct {
	svc      services.OrderService
	validate *validator.Validate
}

func NewOrderHandler(
	svc services.OrderService,
	v *validator.Validate,
) *OrderHandler {
	return &OrderHandler{
		svc:      svc,
		validate: v,
	}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	// ── Idempotency key from header ────────────────────────────────────────
	// The key lives in the HTTP header — it never touches the request body
	// or the domain entity. This is the only place it is read.
	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorBody("Idempotency-Key header is required"))
		return
	}

	// ── Decode and validate request body ──────────────────────────────────
	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorBody("invalid JSON"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.RespondJSON(w, http.StatusUnprocessableEntity, httputil.ValidationErrorBody(err))
		return
	}

	// ── Read caller identity from context (set by auth middleware) ─────────
	userID, ok := contextkeys.GetUserID(r.Context())
	if !ok {
		httputil.RespondJSON(w, http.StatusUnauthorized, httputil.ErrorBody("unauthorized"))
		return
	}

	// ── Build service input — idempotency key is part of the input ────────
	input := services.CreateOrderInput{
		OrderID:        uuid.NewString(),
		UserID:         userID,
		Items:          toDomainOrderItems(req.Items),
		IdempotencyKey: idempotencyKey,
		CreatedBy:      userID,
	}

	order, err := h.svc.Create(r.Context(), input)
	if err != nil {
		// ErrConflict here means either:
		//   a) key is "processing" — concurrent duplicate in flight
		//   b) key is "completed"  — should not reach here (service replays)
		// Both are correctly surfaced as 409.
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusCreated, toOrderResponse(*order))
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, toOrderResponse(*order))
}

func (h *OrderHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := contextkeys.GetUserID(r.Context())
	if !ok {
		httputil.RespondJSON(w, http.StatusUnauthorized, httputil.ErrorBody("unauthorized"))
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	orders, err := h.svc.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		httputil.RespondError(w, err)
		return
	}

	resp := make([]orderResponse, len(orders))
	for i, o := range orders {
		resp[i] = toOrderResponse(o)
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req struct {
		Status string `json:"status" validate:"required"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		httputil.RespondJSON(w, http.StatusBadRequest, httputil.ErrorBody("invalid JSON"))
		return
	}
	if err := h.validate.Struct(req); err != nil {
		httputil.RespondJSON(w, http.StatusUnprocessableEntity, httputil.ValidationErrorBody(err))
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, domain.OrderStatus(req.Status)); err != nil {
		httputil.RespondError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
