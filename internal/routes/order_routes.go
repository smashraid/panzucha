package routes

import (
	"panzucha/internal/handlers"
	"panzucha/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterOrderRoutes(r chi.Router, h *handlers.OrderHandler) {
	r.Route("/orders", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Post("/", h.Create)
			r.Get("/{id}", h.GetByID)
			r.Put("/{id}/status", h.UpdateStatus)
		})
	})
	// List orders for a user: /users/{userID}/orders
	r.Route("/users/{userID}/orders", func(r chi.Router) {
		r.Use(middleware.Authenticate)
		r.Get("/", h.ListByUser)
	})
}
