package routes

import (
	"panzucha/internal/handlers"

	"github.com/go-chi/chi/v5"
)

func RegisterProductRoutes(r chi.Router, h *handlers.ProductHandler) {
	r.Route("/products", func(r chi.Router) {
		r.Post("/", h.Create)
		// r.Get("/", h.List)
		// r.Get("/{id}", h.GetByID)
		// r.Put("/{id}", h.Update)
		// r.Delete("/{id}", h.Delete)
	})
}
