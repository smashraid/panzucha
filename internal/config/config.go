package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Existing fields
	LogstashURL string `env:"LOGSTASH_URL" envDefault:""`
	Environment string `env:"ENV" envDefault:"development"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://admin:admin@localhost:5432/postgres?sslmode=disable"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"myapp"`

	// OpenTelemetry OTLP endpoint (e.g., Jaeger or OpenTelemetry Collector)
	OTLPEndpoint string `env:"OTLP_ENDPOINT" envDefault:"localhost:4317"`

	// Trusted proxies CIDR ranges (comma‑separated). Used by ClientIP middleware.
	TrustedProxies []string `env:"TRUSTED_PROXIES" envDefault:"10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"`

	// Service version – can be set during build (e.g., via -ldflags)
	Version string `env:"SERVICE_VERSION" envDefault:"1.0.0"`

	// RabbitMQ configuration
	RabbitMQURL      string `env:"RABBITMQ_URL" envDefault:"amqp://admin:admin123@localhost:5672/"`
	OrderExchange    string `env:"ORDER_EXCHANGE" envDefault:"order.events"`
	OrderQueuePrefix string `env:"ORDER_QUEUE_PREFIX" envDefault:"order"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}

//export OTLP_ENDPOINT="otel-collector.monitoring.svc.cluster.local:4317"
//export TRUSTED_PROXIES="10.0.0.0/8,172.16.0.0/12,192.168.0.0/16,100.64.0.0/10"
