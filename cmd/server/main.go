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
	"syscall"
	"time"

	"vault_api/internal/api"
	"vault_api/internal/config"
	"vault_api/internal/repository"
)

type dbConnection interface {
	Close()
}

type connectDBFn func(ctx context.Context, databaseURL string) (dbConnection, error)
type buildDepsFn func(db dbConnection) (api.Deps, error)
type listenFn func(server *http.Server) error

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	cfg := config.Load()
	shutdownSignalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := run(
		shutdownSignalCtx,
		cfg,
		func(ctx context.Context, databaseURL string) (dbConnection, error) {
			return repository.NewPostgres(ctx, databaseURL)
		},
		func(db dbConnection) (api.Deps, error) {
			pg, ok := db.(*repository.Postgres)
			if !ok {
				return api.Deps{}, fmt.Errorf("failed to cast postgres to *repository.Postgres")
			}

			return api.Deps{
				Users:              repository.NewUserRepository(pg),
				Sessions:           repository.NewSessionRepository(pg),
				JWTSecret:          cfg.JWTSecret,
				VaultItems:         repository.NewVaultItemRepository(pg),
				CORSAllowedOrigins: cfg.CORSAllowedOrigins,
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

	deps, err := buildDeps(postgres)
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
