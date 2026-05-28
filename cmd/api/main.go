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
	"panzucha/internal/messaging"
	repositories "panzucha/internal/repositories/postgres"

	"panzucha/internal/server"
	"panzucha/internal/services"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10"
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

	rabbitMQ := messaging.NewRabbitMQ(cfg.RabbitMQURL)
	if err := rabbitMQ.Connect(); err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}

	// 3. Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	validate := validator.New()

	// 4. Repository -> Service -> Handler
	userRepo := repositories.NewPostgresUserRepository(pool)
	userService := services.NewUserService(userRepo, log)
	userHandler := handlers.NewUserHandler(userService, log)

	productRepo := repositories.NewPostgresProductRepository(pool)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService, validate)

	idempotencyRepo := repositories.NewPostgresIdempotencyKeyRepository(pool, log)
	idempotencyService := services.NewIdempotencyService(idempotencyRepo)

	orderRepo := repositories.NewPostgresOrderRepository(pool)
	orderService := services.NewOrderService(orderRepo, log)
	orderHandler := handlers.NewOrderHandler(orderService, productService, userService, idempotencyService, log)

	// 5. Router (chi)
	r, telemetryShutdown := server.NewRouter(cfg, productHandler, userHandler, orderHandler)

	defer telemetryShutdown()

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
