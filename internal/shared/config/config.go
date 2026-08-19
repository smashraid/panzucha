package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	// Existing fields
	LogstashURL string `env:"LOGSTASH_URL" envDefault:""`
	Environment string `env:"ENV" envDefault:"development"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://admin:admin@localhost:5432/panzucha_db?sslmode=disable"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"myapp"`

	// OpenTelemetry OTLP endpoint (OpenTelemetry Collector)
	OTLPEndpoint string `env:"OTLP_ENDPOINT" envDefault:"localhost:4317"`

	// Trusted proxies CIDR ranges (comma‑separated). Used by ClientIP middleware.
	TrustedProxies []string `env:"TRUSTED_PROXIES" envDefault:"10.0.0.0/8,172.16.0.0/12,192.168.0.0/16"`

	// Service version – can be set during build (e.g., via -ldflags)
	Version string `env:"SERVICE_VERSION" envDefault:"1.0.0"`

	// RabbitMQ connection (no exchange — that is service-specific config)
	RabbitMQURL string `env:"RABBITMQ_URL" envDefault:"amqp://guest:guest@localhost:5672/"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}
