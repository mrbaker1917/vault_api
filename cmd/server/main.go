package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/redis/go-redis/v9"
	"vault_api/internal/api"
	"vault_api/internal/config"
	"vault_api/internal/crypto"
	redisclient "vault_api/internal/redis"
	"vault_api/internal/ratelimit"
	"vault_api/internal/repository"
)

type dbConnection interface {
	Close()
}

type connectDBFn func(ctx context.Context, databaseURL string) (dbConnection, error)
type buildDepsFn func(db dbConnection, redisClient *redis.Client) (api.Deps, error)
type listenFn func(server *http.Server) error

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()
	if err := cfg.Validate(); err != nil {
		log.Fatal(err)
	}
	shutdownSignalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := run(
		shutdownSignalCtx,
		cfg,
		func(ctx context.Context, databaseURL string) (dbConnection, error) {
			return repository.NewPostgres(ctx, databaseURL)
		},
		func(db dbConnection, redisClient *redis.Client) (api.Deps, error) {
			pg, ok := db.(*repository.Postgres)
			if !ok {
				return api.Deps{}, fmt.Errorf("failed to cast postgres to *repository.Postgres")
			}

			var authLimiter ratelimit.Limiter
			if redisClient != nil {
				authLimiter = ratelimit.NewRedisLimiter(
					redisClient,
					"auth",
					ratelimit.DefaultAuthLimit,
					ratelimit.DefaultAuthWindow,
				)
			} else {
				authLimiter = ratelimit.NewMemoryLimiter(ratelimit.DefaultAuthLimit, ratelimit.DefaultAuthWindow)
			}

			return api.Deps{
				Users:              repository.NewUserRepository(pg),
				Sessions:           repository.NewSessionRepository(pg),
				RecoveryCodes:      repository.NewRecoveryCodeRepository(pg),
				AuditLogs:          repository.NewAuditLogRepository(pg),
				JWTSecret:          cfg.JWTSecret,
				VaultItems:         repository.NewVaultItemRepository(pg),
				SharedVaultItems:   repository.NewSharedVaultItemRepository(pg),
				DB:                 pg,
				CORSAllowedOrigins: cfg.CORSAllowedOrigins,
				PasswordChecker:    crypto.NewHIBPPasswordBreachChecker(nil),
				AuthRateLimiter:    authLimiter,
				Redis:              redisClient,
			}, nil
		},
		api.NewRouter,
		func(server *http.Server) error {
			return server.ListenAndServe()
		},
	)
	if err != nil {
		log.Fatal(err)
	}

}

func run(ctx context.Context, cfg config.Config, connectDB connectDBFn, buildDeps buildDepsFn, buildRouter func(api.Deps) http.Handler, listen listenFn) error {
	dbInitCtx, dbInitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbInitCancel()

	postgres, err := connectDB(dbInitCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}
	defer postgres.Close()

	var redisClient *redis.Client
	if redisURL := strings.TrimSpace(cfg.RedisURL); redisURL != "" {
		redisInitCtx, redisInitCancel := context.WithTimeout(ctx, 5*time.Second)
		client, err := redisclient.NewClient(redisInitCtx, redisURL)
		redisInitCancel()
		if err != nil {
			slog.Warn("redis unavailable; using in-memory auth rate limiter", "error", err)
		} else {
			redisClient = client
			defer redisClient.Close()
			slog.Info("redis connected for distributed auth rate limiting")
		}
	}

	deps, err := buildDeps(postgres, redisClient)
	if err != nil {
		return fmt.Errorf("failed to build dependencies: %w", err)
	}

	addr := fmt.Sprintf(":%s", cfg.Port)
	server := &http.Server{
		Addr:    addr,
		Handler: buildRouter(deps),
	}

	serverErrCh := make(chan error, 1)
	go func() {
		serverErrCh <- listen(server)
	}()

	log.Printf("server listening on http://localhost%s", addr)

	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			log.Printf("http shutdown error: %v", err)
		}
		return nil
	case err := <-serverErrCh:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http server error: %w", err)
		}
		return nil
	}
}
