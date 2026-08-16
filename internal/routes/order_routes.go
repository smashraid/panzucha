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
