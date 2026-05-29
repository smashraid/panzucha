package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"

	"panzucha/internal/contextkeys"
	"panzucha/internal/domain"
	"panzucha/internal/httputil"
	"panzucha/internal/services"
)

type UserHandler struct {
	svc      services.UserService
	validate *validator.Validate
}

func NewUserHandler(s services.UserService, v *validator.Validate) *UserHandler {
	return &UserHandler{svc: s, validate: v}
}

// Register handles POST /users/register
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "register: invalid json", "err", err)
		httputil.RespondError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "register: validation failed", "err", err)
		httputil.RespondError(w, err)
		return
	}

	user, err := h.svc.Register(r.Context(), req.Email, req.Name, req.Password)
	if err != nil {
		if errors.Is(err, domain.ErrConflict) {
			slog.WarnContext(r.Context(), "register: email already exists", "email", req.Email)
		} else {
			slog.ErrorContext(r.Context(), "register: failed", "err", err, "email", req.Email)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "register: user created", "user_id", user.ID)
	httputil.RespondJSON(w, http.StatusCreated, toUserResponse(user))
}

// Login handles POST /users/login
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "login: invalid json", "err", err)
		httputil.RespondError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "login: validation failed", "err", err)
		httputil.RespondError(w, err)
		return
	}

	token, err := h.svc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		// Log at Warn, not Error — wrong credentials is expected, not exceptional.
		// Never log the password or the specific failure reason (user not found vs wrong password).
		slog.WarnContext(r.Context(), "login: authentication failed", "email", req.Email)
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "login: success", "email", req.Email)
	httputil.RespondJSON(w, http.StatusOK, map[string]string{"token": token})
}

// GetProfile handles GET /users/{id}
func (h *UserHandler) GetProfile(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	user, err := h.svc.GetByID(r.Context(), userID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "get_profile: failed", "err", err, "user_id", userID)
		}
		httputil.RespondError(w, err)
		return
	}

	httputil.RespondJSON(w, http.StatusOK, toUserResponse(user))
}

// Update handles PATCH /users/{id}
// updatedBy is read from the auth context — never from the request body.
func (h *UserHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		slog.WarnContext(r.Context(), "update_user: invalid json", "err", err)
		httputil.RespondError(w, domain.ErrInvalidInput)
		return
	}
	if err := h.validate.Struct(req); err != nil {
		slog.WarnContext(r.Context(), "update_user: validation failed", "err", err)
		httputil.RespondError(w, err)
		return
	}

	// updatedBy comes from the auth context — the caller's verified identity.
	// Passing req.Email as updatedBy (as in the original) would mean "the email
	// being updated" is recorded as who made the change, which is wrong.
	callerID, ok := contextkeys.GetUserID(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "update_user: missing auth context", "user_id", userID)
		httputil.RespondError(w, domain.ErrUnauthorized)
		return
	}

	user, err := h.svc.Update(r.Context(), userID, req.Email, req.Name, callerID)
	if err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "update_user: failed", "err", err, "user_id", userID)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "update_user: success", "user_id", userID, "updated_by", callerID)
	httputil.RespondJSON(w, http.StatusOK, toUserResponse(user))
}

// Delete handles DELETE /users/{id}
func (h *UserHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")

	callerID, ok := contextkeys.GetUserID(r.Context())
	if !ok {
		slog.WarnContext(r.Context(), "delete_user: missing auth context", "user_id", userID)
		httputil.RespondError(w, domain.ErrUnauthorized)
		return
	}

	if err := h.svc.Delete(r.Context(), userID); err != nil {
		if !errors.Is(err, domain.ErrNotFound) {
			slog.ErrorContext(r.Context(), "delete_user: failed", "err", err, "user_id", userID)
		}
		httputil.RespondError(w, err)
		return
	}

	slog.InfoContext(r.Context(), "delete_user: success", "user_id", userID, "deleted_by", callerID)
	w.WriteHeader(http.StatusNoContent)
}
