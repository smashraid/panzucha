package config

import (
	"log"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	LogstashURL string `env:"LOGSTASH_URL" envDefault:"http://localhost:8080"`
	Environment string `env:"ENV" envDefault:"development"`
	DatabaseURL string `env:"DATABASE_URL" envDefault:"postgres://admin:admin@postgres:5432/postgres?sslmode=disable"`
	ServiceName string `env:"SERVICE_NAME" envDefault:"myapp"`
}

func Load() *Config {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		log.Fatalf("failed to parse config: %v", err)
	}
	return cfg
}
