package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

// Config holds order-service-specific settings.
// Shared/infra settings (DB, RabbitMQ connection, telemetry) live in
// internal/shared/config — this package owns only order business config.
type Config struct {
	Exchange    string `env:"ORDER_EXCHANGE" envDefault:"order.events"`
	QueuePrefix string `env:"ORDER_QUEUE_PREFIX" envDefault:"order"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse order config: %v", err)
	}
	return cfg
}
