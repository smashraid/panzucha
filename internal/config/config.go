package config

type Config struct {
	LogstashURL string `env:"LOGSTASH_URL" envDefault:"http://localhost:8080"`
	Environment string `env:"ENV" envDefault:"development"`
}
