package server

import (
	"panzucha/internal/config"
	"panzucha/internal/handlers"
	mymiddleware "panzucha/internal/middleware"
	"panzucha/internal/routes"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func NewRouter(
	cfg *config.Config,
	productHandler *handlers.ProductHandler,
	userHandler *handlers.UserHandler,
) *chi.Mux {
	r := chi.NewRouter()

	// Global middlewares
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	if cfg.Environment == "development" {
		r.Use(middleware.Logger)
	} else {
		r.Use(mymiddleware.LoggingMiddleware)
	}

	// Register domain routes
	r.Route("/api/v1", func(r chi.Router) {
		routes.RegisterProductRoutes(r, productHandler)
		routes.RegisterUserRoutes(r, userHandler)
	})

	return r
}
