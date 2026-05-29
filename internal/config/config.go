package config

import "os"

// Config holds environment-driven runtime settings.
type Config struct {
	Port        string
	DatabaseURL string
	RedisURL    string
	JWTSecret   string
}

// Load reads config from environment with sensible local defaults.
func Load() Config {
	return Config{
		Port:        envOrDefault("PORT", "8081"),
		DatabaseURL: envOrDefault("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/vault_api?sslmode=disable"),
		RedisURL:    envOrDefault("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:   envOrDefault("JWT_SECRET", "change-me"),
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
