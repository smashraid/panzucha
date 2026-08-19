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

	"panzucha/internal/order/domain"
	"panzucha/internal/order/services"
	"panzucha/internal/shared/auth"
	shareddomain "panzucha/internal/shared/domain"
	"panzucha/internal/shared/httputil"
	"panzucha/internal/shared/middleware"
)

type OrderHandler struct {
	svc      services.OrderService
	validate *validator.Validate
}

func NewOrderHandler(svc services.OrderService, v *validator.Validate) *OrderHandler {
	return &OrderHandler{svc: svc, validate: v}
}

func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	idempotencyKey, ok := middleware.GetIdempotencyKey(r.Context())
	if !ok {
		slog.ErrorContext(r.Context(), "idempotency key missing in context")
		httputil.RespondError(w, shareddomain.ErrInvalidInput)
		return
	}

	var req createOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "create_order: invalid json", "err", err)
		httputil.RespondError(w, shareddomain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "create_order: validation failed", "err", err)
		httputil.RespondError(w, err)
		return
	}

	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "create_order: missing auth context")
		httputil.RespondError(w, shareddomain.ErrUnauthorized)
		return
	}

	input := services.CreateOrderInput{
		OrderID:        uuid.NewString(),
		UserID:         userID,
		Items:          toDomainOrderItems(req.Items),
		IdempotencyKey: idempotencyKey,
		CreatedBy:      userID,
	}

	order, err := h.svc.Create(r.Context(), input)
	if err != nil {
		switch {
		case errors.Is(err, shareddomain.ErrConflict):
			// Warn — concurrent duplicate is expected behaviour, not a bug.
			slog.WarnContext(r.Context(), "create_order: idempotency conflict",
				"idempotency_key", idempotencyKey, "user_id", userID)
		case errors.Is(err, shareddomain.ErrInsufficientStock):
			slog.WarnContext(r.Context(), "create_order: insufficient stock", "user_id", userID)
		default:
			slog.ErrorContext(r.Context(), "create_order: failed",
				"err", err, "user_id", userID, "idempotency_key", idempotencyKey)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "create_order: success",
		"order_id", order.ID, "user_id", userID, "total", order.TotalAmount)
	httputil.RespondJSON(w, http.StatusCreated, toOrderResponse(*order))
}

func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	order, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		if !errors.Is(err, shareddomain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "get_order: failed", "err", err, "order_id", id)
		}
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toOrderResponse(*order))
}

func (h *OrderHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := auth.UserIDFromContext(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "list_orders: missing auth context")
		httputil.RespondError(w, shareddomain.ErrUnauthorized)
		return
	}

	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	orders, err := h.svc.ListByUser(r.Context(), userID, limit, offset)
	if err != nil {
		slog.ErrorContext(r.Context(), "list_orders: failed",
			"err", err, "user_id", userID, "limit", limit, "offset", offset)
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
		slog.WarnContext(r.Context(), "update_order_status: invalid json", "err", err, "order_id", id)
		httputil.RespondError(w, shareddomain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "update_order_status: validation failed", "err", err, "order_id", id)
		httputil.RespondError(w, err)
		return
	}

	if err := h.svc.UpdateStatus(r.Context(), id, domain.OrderStatus(req.Status)); err != nil {
		if !errors.Is(err, shareddomain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "update_order_status: failed",
				"err", err, "order_id", id, "status", req.Status)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "update_order_status: success", "order_id", id, "status", req.Status)
	w.WriteHeader(http.StatusNoContent)
}

func (h *OrderHandler) RegisterOrderRoutes(r chi.Router) {
	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Get("/{id}", h.GetByID)
			r.Put("/{id}/status", h.UpdateStatus)

			r.Group(func(r chi.Router) {
				r.Use(middleware.RequireIdempotencyKey)
				r.Post("/", h.Create)
			})
			//r.With(middleware.RequireIdempotencyKey).Post("/", h.Create)
		})
	})
	// List orders for a user: /users/{userID}/orders
	r.Route("/users/{userID}/orders", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Get("/", h.ListByUser)
	})
}
