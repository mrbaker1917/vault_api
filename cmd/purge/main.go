package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"strconv"
	"time"

	"vault_api/internal/config"
	"vault_api/internal/repository"
)

const defaultRetentionDays = 30

func main() {
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})))

	retentionDays := defaultRetentionDays
	if raw := os.Getenv("PURGE_RETENTION_DAYS"); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			log.Fatalf("PURGE_RETENTION_DAYS must be a positive integer, got %q", raw)
		}
		retentionDays = parsed
	}

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := repository.NewPostgres(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect postgres: %v", err)
	}
	defer pg.Close()

	vaultItems := repository.NewVaultItemRepository(pg)
	deleted, err := vaultItems.PurgeSoftDeleted(ctx, int32(retentionDays))
	if err != nil {
		log.Fatalf("purge soft-deleted vault items: %v", err)
	}

	fmt.Printf("purged %d soft-deleted vault items older than %d days\n", deleted, retentionDays)
}
