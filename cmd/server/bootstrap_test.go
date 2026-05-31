package main

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"vault_api/internal/api"
	"vault_api/internal/config"
)

func TestBootstrapProjectSkeleton(t *testing.T) {
	root := repoRoot(t)

	requiredDirs := []string{
		"cmd",
		"internal",
		"migrations",
		"docker",
		"docs",
	}

	for _, dir := range requiredDirs {
		t.Run(dir, func(t *testing.T) {
			path := filepath.Join(root, dir)

			info, err := os.Stat(path)
			if err != nil {
				t.Fatalf("expected %s to exist: %v", path, err)
			}
			if !info.IsDir() {
				t.Fatalf("expected %s to be a directory", path)
			}
		})
	}
}

func TestHealthEndpointReturnsOK(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	api.NewRouter().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", rec.Body.String())
	}
}

func TestConfigLoadUsesEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://vault:test@db:5432/vault_api?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://cache:6379")
	t.Setenv("JWT_SECRET", "super-secret")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected Port to use environment override, got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://vault:test@db:5432/vault_api?sslmode=disable" {
		t.Fatalf("expected DatabaseURL to use environment override, got %q", cfg.DatabaseURL)
	}
	if cfg.RedisURL != "redis://cache:6379" {
		t.Fatalf("expected RedisURL to use environment override, got %q", cfg.RedisURL)
	}
	if cfg.JWTSecret != "super-secret" {
		t.Fatalf("expected JWTSecret to use environment override, got %q", cfg.JWTSecret)
	}
}

func TestConfigLoadUsesBootstrapDefaults(t *testing.T) {
	t.Setenv("PORT", "")
	t.Setenv("DATABASE_URL", "")
	t.Setenv("REDIS_URL", "")
	t.Setenv("JWT_SECRET", "")

	cfg := config.Load()

	if cfg.Port != "8080" {
		t.Fatalf("expected default Port %q, got %q", "8080", cfg.Port)
	}
	if cfg.DatabaseURL == "" {
		t.Fatal("expected default DatabaseURL to be non-empty")
	}
	if cfg.RedisURL == "" {
		t.Fatal("expected default RedisURL to be non-empty")
	}
	if cfg.JWTSecret == "" {
		t.Fatal("expected default JWTSecret to be non-empty")
	}
}

func TestDockerComposeBootstrapsAppPostgresAndRedis(t *testing.T) {
	root := repoRoot(t)
	composePath := filepath.Join(root, "docker", "docker-compose.yml")

	content, err := os.ReadFile(composePath)
	if err != nil {
		t.Fatalf("expected bootstrap Docker Compose file at %s: %v", composePath, err)
	}

	assertPattern(t, string(content), `(?m)^services:\s*$`, "top-level services block")
	assertPattern(t, string(content), `(?m)^\s{2,}(app|api|vault-api):\s*$`, "application service definition")
	assertPattern(t, string(content), `(?mi)postgres`, "PostgreSQL service or image reference")
	assertPattern(t, string(content), `(?mi)redis`, "Redis service or image reference")
}

func repoRoot(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}

func assertPattern(t *testing.T, content, pattern, description string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(content) {
		t.Fatalf("expected %s in docker-compose.yml", description)
	}
}
