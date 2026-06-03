package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrEmptyDatabaseURL = errors.New("database url is required")

// Postgres wraps a pgx connection pool used by repository implementations.
type Postgres struct {
	pool *pgxpool.Pool
}

// NewPostgres creates a connection pool and verifies connectivity with a ping.
func NewPostgres(ctx context.Context, databaseURL string) (*Postgres, error) {
	if databaseURL == "" {
		return nil, ErrEmptyDatabaseURL
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &Postgres{pool: pool}, nil
}

// Pool exposes the underlying pool for sqlc/query-layer usage.
func (p *Postgres) Pool() *pgxpool.Pool {
	if p == nil {
		return nil
	}
	return p.pool
}

// Ping checks whether PostgreSQL is reachable.
func (p *Postgres) Ping(ctx context.Context) error {
	if p == nil || p.pool == nil {
		return errors.New("postgres pool is nil")
	}
	return p.pool.Ping(ctx)
}

// BeginTx starts a transaction with the provided options.
func (p *Postgres) BeginTx(ctx context.Context, txOptions pgx.TxOptions) (pgx.Tx, error) {
	if p == nil || p.pool == nil {
		return nil, errors.New("postgres pool is nil")
	}
	return p.pool.BeginTx(ctx, txOptions)
}

// Close releases all open pooled connections.
func (p *Postgres) Close() {
	if p == nil || p.pool == nil {
		return
	}
	p.pool.Close()
}
