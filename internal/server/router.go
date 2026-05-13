package server

import (
	"panzucha/internal/config"
	"panzucha/internal/handlers"
	mymiddleware "panzucha/internal/middleware"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(cfg *config.Config, productHandler *handlers.ProductHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	if cfg.Environment == "development" {
		r.Use(middleware.Logger)
	} else {
		r.Use(mymiddleware.LoggingMiddleware)
	}

	r.Route("/products", func(r chi.Router) {
		r.Post("/", productHandler.Create)
		// r.Get("/", productHandler.List)
		// r.Get("/{id}", productHandler.Get)
		// r.Put("/{id}", productHandler.Update)
		// r.Delete("/{id}", productHandler.Delete)
	})

	return r
}
