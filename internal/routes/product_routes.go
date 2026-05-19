package routes

import (
	"panzucha/internal/handlers"
	"panzucha/internal/middleware"

	"github.com/go-chi/chi/v5"
)

func RegisterProductRoutes(r chi.Router, h *handlers.ProductHandler) {
	r.Route("/products", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(middleware.Authenticate)
			r.Post("/", h.Create)
			r.Get("/", h.List)
			r.Get("/{id}", h.GetByID)
			r.Put("/{id}", h.Update)
			r.Delete("/{id}", h.Delete)
		})
	})
}
