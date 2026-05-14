package routes

import (
	"panzucha/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func RegisterUserRoutes(r chi.Router, h *handlers.UserHandler) {
	r.Route("/users", func(r chi.Router) {
		r.Post("/register", h.Register)
		r.Post("/login", h.Login)
		r.Get("/{id}", h.GetProfile)
		r.Put("/{id}", h.Update)
		// future: add password reset, etc.
	})
}
