package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"panzucha/internal/httputil"
	"panzucha/internal/logger"
	"panzucha/internal/services"
	"panzucha/internal/validation"
)

type UserHandler struct {
	svc        services.UserService
	validation *validator.Validate
}

func NewUserHandler(s services.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{svc: s, validation: v}
}

// Register handles POST /users/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		if msg == "email already registered" {
			status = http.StatusConflict
		}
		http.Error(w, msg, status)
		return
	}

	resp := &userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	httputil.RespondJSON(w, http.StatusCreated, resp)
}

// Login handles POST /users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}
	httputil.RespondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// GetProfile handles GET /users/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	if userID == "" {
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		status := http.StatusNotFound
		msg := logger.MsgBusinessNotFound
		http.Error(w, msg, status)
		return
	}

	resp := &userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}

// Update handles PUT /users/{id}
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	if userID == "" {
		http.Error(w, logger.MsgBusinessInvalidIdentifier, http.StatusBadRequest)
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, logger.MsgBusinessInvalidJSON, http.StatusBadRequest)
		return
	}

	if err := validation.Validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	user, err := h.svc.Update(r.Context(), userID, req.Email, req.Name, req.Email)
	if err != nil {
		status := http.StatusBadRequest
		msg := err.Error()
		if msg == logger.MsgBusinessNotFound {
			status = http.StatusNotFound
		}

		http.Error(w, msg, status)
		return
	}

	resp := &userResponse{
		ID:        user.ID,
		Email:     user.Email,
		Name:      user.Name,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}
	httputil.RespondJSON(w, http.StatusOK, resp)
}
