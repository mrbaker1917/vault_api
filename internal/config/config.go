package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds environment-driven runtime settings.
type Config struct {
	Port               string
	DatabaseURL        string
	RedisURL           string
	JWTSecret          string
	AppEnv             string
	CORSAllowedOrigins []string
}

// Load reads config from environment with sensible local defaults.
func Load() Config {
	return Config{
		Port:               envOrDefault("PORT", "8081"),
		DatabaseURL:        envOrDefault("DATABASE_URL", "postgres://vault:vault@localhost:5433/vault_api?sslmode=disable"),
		RedisURL:           envOrDefault("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:          envOrDefault("JWT_SECRET", "change-me"),
		AppEnv:             envOrDefault("APP_ENV", "production"),
		CORSAllowedOrigins: corsAllowedOrigins(),
	}
}

// Validate rejects insecure JWT secrets outside development.
func (c Config) Validate() error {
	if isDevelopment(c.AppEnv) {
		return nil
	}
	secret := strings.TrimSpace(c.JWTSecret)
	if secret == "" || isInsecureJWTSecret(secret) {
		return fmt.Errorf("JWT_SECRET must be set to a strong value (not empty, change-me, or dev-secret); set APP_ENV=development only for local use")
	}
	if len(secret) < 32 {
		return fmt.Errorf("JWT_SECRET must be at least 32 characters in %s", c.AppEnv)
	}
	return nil
}

func isDevelopment(appEnv string) bool {
	switch strings.ToLower(strings.TrimSpace(appEnv)) {
	case "development", "dev", "local", "test":
		return true
	default:
		return false
	}
}

func isInsecureJWTSecret(secret string) bool {
	switch strings.ToLower(strings.TrimSpace(secret)) {
	case "change-me", "dev-secret", "secret", "jwt-secret", "password":
		return true
	default:
		return false
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
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

func envOrDefault(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
