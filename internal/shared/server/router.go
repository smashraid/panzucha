package server

import (
	"log/slog"
	"net"
	"panzucha/internal/shared/config"
	sharedmiddleware "panzucha/internal/shared/middleware"

	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// RouterRegistrar wires domain routes onto a chi router.
// Defined here so the shared server stays domain-agnostic —
// each cmd entrypoint composes its own registrations.
type RouterRegistrar func(r chi.Router)

func NewRouter(
	cfg *config.Config,
	register RouterRegistrar,
) *chi.Mux {
	r := chi.NewRouter()

	// Define Trusted Proxies for Client IP Extraction
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

	r.Use(middleware.Recoverer)
	r.Use(sharedmiddleware.ClientIP(trustedProxies))
	r.Use(middleware.RequestID)

	// OpenTelemetry tracing & metrics (automatically records HTTP metrics)
	r.Use(otelhttp.NewMiddleware(cfg.ServiceName))

	// Logging & metrics
	if cfg.Environment == "development" {
		r.Use(middleware.Logger)
	} else {
		r.Use(sharedmiddleware.StructuredLogger(slog.Default()))
	}

	r.Use(middleware.Timeout(60 * time.Second))

	// Expose Prometheus metrics endpoint
	r.Handle("/metrics", promhttp.Handler())

	// Register domain routes
	r.Route("/api/v1", func(r chi.Router) {
		register(r)
	})

	return r
}
