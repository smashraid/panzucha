package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"

	"panzucha/internal/handlers/dto"
	"panzucha/internal/httputil"
	"panzucha/internal/logger"
	"panzucha/internal/services"
	"panzucha/internal/validation"
)

type UserHandler struct {
	service services.UserService
	logger  *logger.Logger
}

func NewUserHandler(s services.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{service: s, logger: log}
}

// Register handles POST /users/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)

	var req dto.RegisterRequest
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

	user, err := h.service.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		if msg == "email already registered" {
			status = http.StatusConflict
		}
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: status,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     "", // user not created
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    msg,
			Payload:    req,
		})
		http.Error(w, msg, status)
		return
	}

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusCreated,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     user.ID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessCreated,
		Payload:    req,
	})
	httputil.RespondJSON(w, http.StatusCreated, resp)
}

// Login handles POST /users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)

	var req dto.LoginRequest
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

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		h.logger.LogAPI(logger.APILogParams{
			Ctx:        r.Context(),
			Method:     r.Method,
			Path:       r.URL.Path,
			StatusCode: http.StatusUnauthorized,
			Duration:   time.Since(info.StartTime),
			RequestID:  info.RequestID,
			UserID:     "", // unknown at this point
			ClientIP:   info.ClientIP,
			UserAgent:  info.UserAgent,
			Err:        err,
			Message:    "invalid credentials",
			Payload:    req,
		})
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     req.Email, // email as fallback; better to get from service
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    "user logged in",
		Payload:    req,
	})
	httputil.RespondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// GetProfile handles GET /users/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	userID := chi.URLParam(r, "id")

	if userID == "" {
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

	user, err := h.service.GetByID(r.Context(), userID)
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

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     user.ID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessRetrieved,
		Payload:    nil,
	})
	httputil.RespondJSON(w, http.StatusOK, resp)
}

// Update handles PUT /users/{id}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	info := ExtractRequestInfo(r)
	userID := chi.URLParam(r, "id")

	if userID == "" {
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

	var req dto.UpdateRequest
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

	user, err := h.service.Update(r.Context(), userID, req.Email, req.Name)
	if err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		if msg == logger.MsgBusinessNotFound {
			status = http.StatusNotFound
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

	resp := &dto.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	h.logger.LogAPI(logger.APILogParams{
		Ctx:        r.Context(),
		Method:     r.Method,
		Path:       r.URL.Path,
		StatusCode: http.StatusOK,
		Duration:   time.Since(info.StartTime),
		RequestID:  info.RequestID,
		UserID:     user.ID,
		ClientIP:   info.ClientIP,
		UserAgent:  info.UserAgent,
		Err:        nil,
		Message:    logger.MsgBusinessUpdated,
		Payload:    req,
	})
	httputil.RespondJSON(w, http.StatusOK, resp)
}
