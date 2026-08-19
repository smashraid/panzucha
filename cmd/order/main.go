package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	orderconfig "panzucha/internal/order/config"
	orderhandler "panzucha/internal/order/handlers"
	orderrepo "panzucha/internal/order/repositories/postgres"
	orderservice "panzucha/internal/order/services"
	producthandler "panzucha/internal/product/handlers"
	productrepo "panzucha/internal/product/repositories/postgres"
	productservice "panzucha/internal/product/services"
	"panzucha/internal/shared/config"
	"panzucha/internal/shared/db"
	"panzucha/internal/shared/idempotency"
	"panzucha/internal/shared/messaging"
	"panzucha/internal/shared/outbox"
	"panzucha/internal/shared/server"
	"panzucha/internal/shared/telemetry"
	userhandler "panzucha/internal/user/handlers"
	userrepo "panzucha/internal/user/repositories/postgres"
	userservice "panzucha/internal/user/services"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	// 1. Load config
	cfg := config.Load()
	orderCfg := orderconfig.Load()

	// 2. Setup logger (structured)
	stdLogger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(stdLogger)

	broker := messaging.NewRabbitMQBroker(cfg.RabbitMQURL, orderCfg.Exchange)
	if err := broker.Connect(); err != nil {
		slog.Error("failed to connect to RabbitMQ", "error", err)
		os.Exit(1)
	}
	defer broker.Close()

	// 3. Database connection
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	// 4. Telemetry (OpenTelemetry traces via OTLP, metrics via Prometheus)
	telemetryShutdown, err := telemetry.Init(ctx, cfg.ServiceName, cfg.Version, cfg.Environment, cfg.OTLPEndpoint)
	if err != nil {
		slog.Error("failed to init telemetry", "error", err)
		os.Exit(1)
	}
	defer telemetryShutdown(ctx)

	validate := validator.New()
	transactor := db.NewPgxTransactor(pool)

	// 4. Repository -> Service -> Handler
	userRepo := userrepo.NewPostgresUserRepository(pool)
	userService := userservice.NewUserService(userRepo)
	userHandler := userhandler.NewUserHandler(userService, validate)

	productRepo := productrepo.NewPostgresProductRepository(pool)
	productService := productservice.NewProductService(productRepo)
	productHandler := producthandler.NewProductHandler(productService, validate)

	idempotencyRepo := idempotency.NewPostgresIdempotencyKeyRepository(pool)
	outboxRepo := outbox.NewPostgresOutboxRepository(pool)

	orderRepo := orderrepo.NewPostgresOrderRepository(pool)
	orderService := orderservice.NewOrderService(transactor, orderRepo, productRepo, idempotencyRepo, outboxRepo)
	orderHandler := orderhandler.NewOrderHandler(orderService, validate)

	outboxCfg := outbox.Config{}
	outboxRelay := outbox.NewRelay(pool, outboxRepo, broker, outboxCfg)
	go outboxRelay.Start(ctx)

	// 5. Router (chi)
	r := server.NewRouter(cfg, func(r chi.Router) {
		userHandler.RegisterUserRoutes(r)
		productHandler.RegisterProductRoutes(r)
		orderHandler.RegisterOrderRoutes(r)
	})

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
