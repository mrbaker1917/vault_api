package main

import (
	"context"
	"errors"
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

type stubDBConnection struct {
	closed bool
}

func (s *stubDBConnection) Close() {
	s.closed = true
}

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

	api.NewRouter(api.Deps{
		Users: nil,
		Sessions: nil,
		VaultItems: nil,
		JWTSecret: "test-secret",
	}).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}
	if rec.Body.String() != "ok" {
		t.Fatalf("expected body %q, got %q", "ok", rec.Body.String())
	}
}

func TestConfigLoadUsesEnvironmentVariables(t *testing.T) {
	t.Setenv("PORT", "9090")
	t.Setenv("DATABASE_URL", "postgres://vault:test@db:5433/vault_api?sslmode=disable")
	t.Setenv("REDIS_URL", "redis://cache:6379")
	t.Setenv("JWT_SECRET", "super-secret")

	cfg := config.Load()

	if cfg.Port != "9090" {
		t.Fatalf("expected Port to use environment override, got %q", cfg.Port)
	}
	if cfg.DatabaseURL != "postgres://vault:test@db:5433/vault_api?sslmode=disable" {
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

func TestRunBootstrapsDBAndRouterWithoutLiveDatabase(t *testing.T) {
	t.Parallel()

	ctx := context.Background()
	cfg := config.Config{
		Port:        "8080",
		DatabaseURL: "postgres://example",
	}

	stubDB := &stubDBConnection{}
	connected := false
	healthHandled := false

	err := run(
		ctx,
		cfg,
		func(_ context.Context, databaseURL string) (dbConnection, error) {
			if databaseURL != cfg.DatabaseURL {
				t.Fatalf("expected database URL %q, got %q", cfg.DatabaseURL, databaseURL)
			}
			connected = true
			return stubDB, nil
		},
		api.NewRouter,
		func(server *http.Server) error {
			req := httptest.NewRequest(http.MethodGet, "/health", nil)
			rec := httptest.NewRecorder()
			server.Handler.ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("expected health status %d, got %d", http.StatusOK, rec.Code)
			}
			if rec.Body.String() != "ok" {
				t.Fatalf("expected health body %q, got %q", "ok", rec.Body.String())
			}

			healthHandled = true
			return nil
		},
	)

	if err != nil {
		t.Fatalf("expected run to succeed, got error: %v", err)
	}
	if !connected {
		t.Fatal("expected database connector to be called")
	}
	if !healthHandled {
		t.Fatal("expected router health endpoint to be exercised")
	}
	if !stubDB.closed {
		t.Fatal("expected database connection to be closed on exit")
	}
}

func TestRunReturnsWrappedErrorWhenDBInitializationFails(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		Port:        "8080",
		DatabaseURL: "postgres://example",
	}

	expectedErr := errors.New("db unavailable")
	listenCalled := false

	err := run(
		context.Background(),
		cfg,
		func(_ context.Context, _ string) (dbConnection, error) {
			return nil, expectedErr
		},
		api.NewRouter,
		func(_ *http.Server) error {
			listenCalled = true
			return nil
		},
	)

	if err == nil {
		t.Fatal("expected run to return an error when DB initialization fails")
	}
	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected error to wrap %v, got %v", expectedErr, err)
	}
	if err.Error() != "failed to initialize postgres: db unavailable" {
		t.Fatalf("expected wrapped startup error message, got %q", err.Error())
	}
	if listenCalled {
		t.Fatal("expected HTTP listen not to start when DB initialization fails")
	}
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
