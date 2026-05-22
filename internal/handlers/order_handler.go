package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"panzucha/internal/auth"
	"panzucha/internal/domain"
	"panzucha/internal/handlers/dto"
	"panzucha/internal/handlers/mapper"
	"panzucha/internal/httputil"
	"panzucha/internal/logger"
	"panzucha/internal/services"
	"panzucha/internal/validation"
)

type OrderHandler struct {
	orderService   services.OrderService
	productService services.ProductService
	userService    services.UserService
	logger         *logger.Logger
}

func NewOrderHandler(
	orderService services.OrderService,
	productService services.ProductService,
	userService services.UserService,
	log *logger.Logger,
) *OrderHandler {
	return &OrderHandler{
		orderService:   orderService,
		productService: productService,
		userService:    userService,
		logger:         log,
	}
}

// Create handles POST /orders
func (h *OrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)

	var req dto.CreateOrderRequest
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

	// Verify user exists
	_, err := h.userService.GetByID(r.Context(), info.UserID)
	if err != nil {
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
			Message:    "user not found",
			Payload:    req,
		})
		http.Error(w, "user not found", http.StatusBadRequest)
		return
	}

	// Verify product exists and get its price
	product, err := h.productService.GetByID(r.Context(), req.ProductID)
	if err != nil {
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
			Message:    "product not found",
			Payload:    req,
		})
		http.Error(w, "product not found", http.StatusBadRequest)
		return
	}

	// Create order domain object
	order := mapper.FromCreateOrderRequest(&req, info.UserID)
	order.TotalPrice = product.Price * float64(order.Quantity)

	if err := h.orderService.Create(r.Context(), order); err != nil {
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
			Message:    "failed to create order",
			Payload:    req,
		})
		http.Error(w, "failed to create order", http.StatusInternalServerError)
		return
	}

	resp := mapper.FromDomainToOrderResponse(order)
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

// GetByID handles GET /orders/{id}
func (h *OrderHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	orderID := chi.URLParam(r, "id")

	if orderID == "" {
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

	order, err := h.orderService.GetByID(r.Context(), orderID)
	if err != nil {
		status := http.StatusNotFound
		msg := logger.MsgBusinessNotFound
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

	// Ensure the authenticated user owns the order (or is admin)
	if order.UserID != info.UserID && !isAdmin(r.Context()) {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusForbidden,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessForbidden,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessForbidden, http.StatusForbidden)
		return
	}

	resp := mapper.FromDomainToOrderResponse(order)
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

// ListByUser handles GET /users/{userID}/orders
func (h *OrderHandler) ListByUser(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	userID := chi.URLParam(r, "userID")

	if userID == "" {
		userID = info.UserID // default to authenticated user if not provided
	}

	// Only allow admin or the user themselves
	if userID != info.UserID && !isAdmin(r.Context()) {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusForbidden,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessForbidden,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessForbidden, http.StatusForbidden)
		return
	}

	orders, err := h.orderService.GetByUserID(r.Context(), userID)
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

	resp := make([]dto.OrderResponse, len(orders))
	for i, o := range orders {
		resp[i] = *mapper.FromDomainToOrderResponse(&o)
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

// UpdateStatus handles PUT /orders/{id}/status
func (h *OrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	orderID := chi.URLParam(r, "id")

	if orderID == "" {
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

	var req dto.UpdateOrderStatusRequest
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

	// Admin-only or role-based – for simplicity, only admin can update status
	if !isAdmin(r.Context()) {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusForbidden,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     info.UserID,
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        nil,
			Message:    logger.MsgBusinessForbidden,
			Payload:    nil,
		})
		http.Error(w, logger.MsgBusinessForbidden, http.StatusForbidden)
		return
	}

	if err := h.orderService.UpdateStatus(r.Context(), orderID, req.Status); err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		if err == domain.ErrNotFound {
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
	httputil.RespondJSON(w, http.StatusOK, map[string]string{"status": "updated"})
}

// Helper to check if user has admin role (implement based on your auth context)
func isAdmin(ctx context.Context) bool {
	roles, ok := auth.RolesFromContext(ctx)
	if !ok {
		return false
	}
	for _, r := range roles {
		if r == "admin" {
			return true
		}
	}
	return false
}
