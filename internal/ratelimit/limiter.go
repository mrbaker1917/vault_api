package ratelimit

import (
	"context"
	"time"
)

const (
	DefaultAuthLimit  = 10
	DefaultAuthWindow = time.Minute
)

// Limiter tracks request counts for a key within a time window.
type Limiter interface {
	Allow(ctx context.Context, key string) (bool, error)
}
