package telemetry

import (
	"context"
	"errors"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	otelprometheus "go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Shutdown flushes telemetry and releases its resources.
type Shutdown func(ctx context.Context) error

// Init configures OpenTelemetry: traces export via OTLP to a collector,
// metrics register on the default Prometheus registry so the existing
// /metrics endpoint (promhttp.Handler()) serves them alongside any
// promauto counters. It sets the global tracer and meter providers.
func Init(ctx context.Context, serviceName, version, environment, otlpEndpoint string) (Shutdown, error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(version),
			semconv.DeploymentEnvironmentKey.String(environment),
		),
	)
	if err != nil {
		return nil, err
	}

	// --- Traces (OTLP to OpenTelemetry Collector) ---
	traceExporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(otlpEndpoint),
	)
	if err != nil {
		return nil, err
	}
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)

	// --- Metrics (Prometheus) ---
	// Register on the default registry so /metrics serves both the OTel
	// (otelhttp) metrics and the promauto counters defined in shared/metrics.
	metricExporter, err := otelprometheus.New(otelprometheus.WithRegisterer(prometheus.DefaultRegisterer))
	if err != nil {
		return nil, err
	}
	mp := metric.NewMeterProvider(
		metric.WithReader(metricExporter),
		metric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return func(ctx context.Context) error {
		return errors.Join(tp.Shutdown(ctx), mp.Shutdown(ctx))
	}, nil
}
