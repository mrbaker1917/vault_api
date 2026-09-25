package ratelimit

import (
	"context"
	"sync"
	"time"
)

// MemoryLimiter is an in-process fixed-window limiter for single-instance deployments.
type MemoryLimiter struct {
	mu      sync.Mutex
	limit   int
	window  time.Duration
	buckets map[string][]time.Time
}

func NewMemoryLimiter(limit int, window time.Duration) *MemoryLimiter {
	return &MemoryLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string][]time.Time),
	}
}

func (l *MemoryLimiter) Allow(_ context.Context, key string) (bool, error) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()
	cutoff := now.Add(-l.window)

	hits := l.buckets[key]
	active := hits[:0]
	for _, hit := range hits {
		if hit.After(cutoff) {
			active = append(active, hit)
		}
	}

	if len(active) >= l.limit {
		l.buckets[key] = active
		return false, nil
	}

	active = append(active, now)
	l.buckets[key] = active
	return true, nil
}

// Reset clears all tracked keys. Useful in tests.
func (l *MemoryLimiter) Reset() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.buckets = make(map[string][]time.Time)
}
