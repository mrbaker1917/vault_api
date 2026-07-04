package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
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
type listenFn func(server *http.Server) error

func main() {
	cfg := config.Load()
	shutdownSignalCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err := run(
		shutdownSignalCtx,
		cfg,
		func(ctx context.Context, databaseURL string) (dbConnection, error) {
			return repository.NewPostgres(ctx, databaseURL)
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

func run(ctx context.Context, cfg config.Config, connectDB connectDBFn, buildRouter func(api.Deps) http.Handler, listen listenFn) error {
	dbInitCtx, dbInitCancel := context.WithTimeout(ctx, 10*time.Second)
	defer dbInitCancel()

	postgres, err := connectDB(dbInitCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to initialize postgres: %w", err)
	}
	defer postgres.Close()

	pg, ok := postgres.(*repository.Postgres)
	if !ok {
		return fmt.Errorf("failed to cast postgres to *repository.Postgres")
	}

	users := repository.NewUserRepository(pg)
	sessions := repository.NewSessionRepository(pg)
	vaultItems := repository.NewVaultItemRepository(pg)

	deps := api.Deps{
		Users: users,
		Sessions: sessions,
		JWTSecret: cfg.JWTSecret,
		VaultItems: vaultItems,
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
