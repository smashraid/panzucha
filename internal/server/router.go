package server

import (
	"context"
	"log"
	"log/slog"
	"net"
	"panzucha/internal/config"
	"panzucha/internal/handlers"
	mymiddleware "panzucha/internal/middleware"
	"panzucha/internal/routes"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
	"go.uber.org/zap"
)

func NewRouter(
	cfg *config.Config,
	productHandler *handlers.ProductHandler,
	userHandler *handlers.UserHandler,
	orderHandler *handlers.OrderHandler,
) (*chi.Mux, func()) {
	ctx := context.Background()

	// --- 1. Setup OpenTelemetry Resource ---
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.Version),
			attribute.String("environment", cfg.Environment),
		),
	)
	if err != nil {
		log.Fatal("failed to create resource", zap.Error(err))
	}

	// --- 2. Setup OTLP Trace Exporter (to Jaeger / Collector) ---
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	)
	if err != nil {
		log.Fatal("failed to create trace exporter", zap.Error(err))
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// --- 3. Setup Prometheus Metric Exporter ---
	metricExporter, err := prometheus.New()
	if err != nil {
		log.Fatal("failed to create metric exporter", zap.Error(err))
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	shutdown := func() {
		_ = tp.Shutdown(ctx)
		_ = mp.Shutdown(ctx)
	}

	r := chi.NewRouter()

	// --- 4. Define Trusted Proxies for Client IP Extraction ---
	var trustedProxies []*net.IPNet
	for _, cidr := range cfg.TrustedProxies {
		_, ipNet, err := net.ParseCIDR(cidr)
		if err == nil {
			trustedProxies = append(trustedProxies, ipNet)
		}
	}
	// Always trust loopback
	_, loopback, _ := net.ParseCIDR("127.0.0.0/8")
	trustedProxies = append(trustedProxies, loopback)

	// Security & Request Identification
	r.Use(mymiddleware.ClientIP(trustedProxies))
	r.Use(mymiddleware.RequestID)

	// OpenTelemetry tracing & metrics (automatically records HTTP metrics)
	r.Use(otelhttp.NewMiddleware("panzucha-api"))

	// Standard Recovery and Timeout
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Logging & metrics
	if cfg.Environment == "development" {
		r.Use(middleware.Logger)
	} else {
		r.Use(mymiddleware.StructuredLogger(slog.Default()))
	}

	// Expose Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Register domain routes
	r.Route("/api/v1", func(r chi.Router) {
		routes.RegisterProductRoutes(r, productHandler)
		routes.RegisterUserRoutes(r, userHandler)
		routes.RegisterOrderRoutes(r, orderHandler)
	})

	return r, shutdown
}
