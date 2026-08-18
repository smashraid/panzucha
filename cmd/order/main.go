package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"panzucha/internal/order/handlers"
	"panzucha/internal/order/repositories/postgres"
	"panzucha/internal/order/services"
	"panzucha/internal/shared/config"
	"panzucha/internal/shared/idempotency"
	"panzucha/internal/shared/messaging"
	"panzucha/internal/shared/outbox"
	"panzucha/internal/shared/server"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load config
	cfg := config.Load()

	// 2. Setup logger (structured)
	stdLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(stdLogger)

	broker := messaging.NewRabbitMQBroker(cfg.RabbitMQURL, "order.events")
	if err := broker.Connect(); err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer broker.Close()

	//pub := publisher.New(broker)

	// 3. Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	validate := validator.New()
	transactor := postgres.NewPgxTransactor(pool)

	// 4. Repository -> Service -> Handler
	userRepo := postgres.NewPostgresUserRepository(pool)
	userService := services.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService, validate)

	productRepo := postgres.NewPostgresProductRepository(pool)
	productService := services.NewProductService(productRepo)
	productHandler := handlers.NewProductHandler(productService, validate)

	idempotencyRepo := idempotency.NewPostgresIdempotencyKeyRepository(pool)
	outboxRepo := outbox.NewPostgresOutboxRepository(pool)

	orderRepo := postgres.NewPostgresOrderRepository(pool)
	orderService := services.NewOrderService(transactor, orderRepo, productRepo, idempotencyRepo, outboxRepo)
	orderHandler := handlers.NewOrderHandler(orderService, validate)

	outboxCfg := outbox.Config{}
	outboxRelay := outbox.NewRelay(pool, outboxRepo, broker, outboxCfg)
	go outboxRelay.Start(ctx)

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
