package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"panzucha/internal/config"
	"panzucha/internal/handlers"
	"panzucha/internal/logger"
	repositories "panzucha/internal/repositories/postgres"

	"panzucha/internal/server"
	"panzucha/internal/services"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Setup logger (structured)
	stdLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(stdLogger)

	log := logger.New(cfg)
	// defer log.Close()

	// 3. Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Repository -> Service -> Handler
	userRepo := repositories.NewPostgresUserRepository(pool, log)
	userService := services.NewUserService(userRepo, log)
	userHandler := handlers.NewUserHandler(userService, log)

	productRepo := repositories.NewPostgresProductRepository(pool, log)
	productService := services.NewProductService(productRepo, log)
	productHandler := handlers.NewProductHandler(productService, log)

	// 5. Router (chi)
	r := server.NewRouter(cfg, productHandler, userHandler)

	// 6. HTTP server
	srv := &http.Server{
		Addr:    ":8080",
		Handler: r,
	}

	// 7. Graceful shutdown
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server failed", "error", err)
			os.Exit(1)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	ctxShutdown, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctxShutdown); err != nil {
		slog.Error("forced shutdown", "error", err)
	}
	slog.Info("server stopped")
}
