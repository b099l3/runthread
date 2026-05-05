package config

import (
	"os"
	"strings"
)

const (
	DefaultServerAddress = ":8080"

	EnvServerAddress = "RUNTHREAD_SERVER_ADDR"
	EnvDatabaseURL   = "DATABASE_URL"
)

type Config struct {
	ServerAddress string
	DatabaseURL   string
}

func Load() Config {
	return Config{
		ServerAddress: stringFromEnv(EnvServerAddress, DefaultServerAddress),
		DatabaseURL:   strings.TrimSpace(os.Getenv(EnvDatabaseURL)),
	}
}

func (c Config) DatabaseConfigured() bool {
	return c.DatabaseURL != ""
}

func stringFromEnv(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}
