//go:build integration

package integration

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"vault_api/internal/api"
	"vault_api/internal/api/middleware"
	"vault_api/internal/repository"
)

const (
	testJWTSecret = "integration-test-secret"
	postgresImage = "postgres:16-alpine"
	postgresDB    = "vault_api"
	postgresUser  = "vault"
	postgresPass  = "vault"
)

func NewTestRouter(t *testing.T) (http.Handler, func()) {
	t.Helper()

	if testing.Short() {
		t.Skip("skipping integration test in short mode")
	}

	middleware.ResetAuthRateLimiterForTests()

	connStr, cleanupDB, ok := resolveDatabaseURL(t)
	if !ok {
		t.Skip("integration tests require TEST_DATABASE_URL or a working Docker environment")
	}
	if cleanupDB == nil {
		cleanupDB = func() {}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := repository.NewPostgres(ctx, connStr)
	if err != nil {
		t.Fatalf("connect postgres: %v", err)
	}

	migrateCtx, migrateCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer migrateCancel()
	if err := runMigrations(migrateCtx, pg.Pool(), migrationsDir(t)); err != nil {
		t.Fatalf("run migrations: %v", err)
	}

	deps := api.Deps{
		Users:         repository.NewUserRepository(pg),
		Sessions:      repository.NewSessionRepository(pg),
		RecoveryCodes: repository.NewRecoveryCodeRepository(pg),
		AuditLogs:     repository.NewAuditLogRepository(pg),
		JWTSecret:     testJWTSecret,
		VaultItems:    repository.NewVaultItemRepository(pg),
	}

	cleanup := func() {
		pg.Close()
		cleanupDB()
	}

	return api.NewRouter(deps), cleanup
}

func resolveDatabaseURL(t *testing.T) (connStr string, cleanup func(), ok bool) {
	t.Helper()

	if url := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL")); url != "" {
		return url, func() {}, true
	}

	connStr, cleanup, err := startPostgresContainer()
	if err != nil {
		t.Logf("testcontainers unavailable: %v", err)
		return "", nil, false
	}
	return connStr, cleanup, true
}

func startPostgresContainer() (connStr string, cleanup func(), err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("docker unavailable: %v", r)
		}
	}()

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	pgContainer, runErr := postgres.Run(ctx,
		postgresImage,
		postgres.WithDatabase(postgresDB),
		postgres.WithUsername(postgresUser),
		postgres.WithPassword(postgresPass),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(60*time.Second),
		),
	)
	if runErr != nil {
		return "", nil, runErr
	}

	connStr, err = pgContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		_ = pgContainer.Terminate(termCtx)
		return "", nil, err
	}

	cleanup = func() {
		termCtx, termCancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer termCancel()
		_ = pgContainer.Terminate(termCtx)
	}

	return connStr, cleanup, nil
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool, dir string) error {
	if err := resetSchema(ctx, pool); err != nil {
		return err
	}

	files := []string{
		"004_enable_pgcrypto.sql",
		"001_create_users.sql",
		"002_create_sessions.sql",
		"003_create_vault_items.sql",
	}

	for _, name := range files {
		content, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		sql := gooseUpSQL(string(content))
		if sql == "" {
			continue
		}
		if _, err := pool.Exec(ctx, sql); err != nil {
			return fmt.Errorf("migration %s: %w", name, err)
		}
	}
	return nil
}

func resetSchema(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, `
		DROP TABLE IF EXISTS shared_vault_items CASCADE;
		DROP TABLE IF EXISTS recovery_codes CASCADE;
		DROP TABLE IF EXISTS audit_log CASCADE;
		DROP TABLE IF EXISTS vault_items CASCADE;
		DROP TABLE IF EXISTS sessions CASCADE;
		DROP TABLE IF EXISTS users CASCADE;
	`)
	return err
}

func gooseUpSQL(content string) string {
	const upMarker = "-- +goose Up"
	const downMarker = "-- +goose Down"

	upIdx := strings.Index(content, upMarker)
	if upIdx == -1 {
		return strings.TrimSpace(content)
	}

	sql := content[upIdx+len(upMarker):]
	if downIdx := strings.Index(sql, downMarker); downIdx != -1 {
		sql = sql[:downIdx]
	}
	return strings.TrimSpace(sql)
}

func migrationsDir(t *testing.T) string {
	t.Helper()

	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("failed to determine test file location")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", "..", "migrations"))
}
