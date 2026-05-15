package handlers

import (
	"encoding/json"
	"net/http"
	"panzucha/internal/logger"
	"panzucha/internal/middleware"
	"panzucha/internal/services"
	"time"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	service services.UserService
	logger  *logger.Logger
}

func NewUserHandler(s services.UserService, log *logger.Logger) *UserHandler {
	return &UserHandler{service: s, logger: log}
}

type registerRequest struct {
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type updateRequest struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	requestID := middleware.GetRequestID(r.Context())

	clientIP := r.RemoteAddr
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		clientIP = xff
	}
	userAgent := r.UserAgent()

	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.LogAPI(r.Context(), r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start),
			requestID, "", clientIP, userAgent, "invalid json")
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		h.logger.LogAPI(r.Context(), r.Method, r.URL.Path, http.StatusBadRequest, time.Since(start),
			requestID, "", clientIP, userAgent, err.Error())
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	h.logger.LogAPI(r.Context(), r.Method, r.URL.Path, http.StatusCreated, time.Since(start),
		requestID, user.ID, clientIP, userAgent, "user registered")
	respondJSON(w, http.StatusCreated, user)
}

func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	token, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		http.Error(w, "invalid credentials", http.StatusUnauthorized)
		return
	}

	respondJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	// The user ID would come from the authenticated context (JWT middleware)
	userID := chi.URLParam(r, "id")
	if userID == "" {
		http.Error(w, "missing user id", http.StatusBadRequest)
		return
	}
	user, err := h.service.GetByID(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.service.Update(r.Context(), userID, req.Email, req.Name)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	respondJSON(w, http.StatusOK, user)
}

// Helper
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}
