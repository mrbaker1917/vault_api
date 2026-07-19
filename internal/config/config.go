package config

import (
	"os"
	"strings"
)

// Config holds environment-driven runtime settings.
type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	CORSAllowedOrigins []string
}

// Load reads config from environment with sensible local defaults.
func Load() Config {
	return Config{
		Port:               envOrDefault("PORT", "8081"),
		DatabaseURL:        envOrDefault("DATABASE_URL", "postgres://vault:vault@localhost:5433/vault_api?sslmode=disable"),
		RedisURL:           envOrDefault("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          envOrDefault("JWT_SECRET", "change-me"),
		CORSAllowedOrigins: corsAllowedOrigins(),
	}
}

func corsAllowedOrigins() []string {
	if value := strings.TrimSpace(os.Getenv("CORS_ALLOWED_ORIGINS")); value != "" {
		return splitCSV(value)
	}
	return []string{
		"http://localhost:5173",
		"http://127.0.0.1:5173",
	}
}

func splitCSV(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	origins := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			origins = append(origins, trimmed)
		}
	}
	return origins
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
