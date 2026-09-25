package ratelimit

import (
	"context"
	"testing"
	"time"
)

func TestMemoryLimiterAllowsUpToLimit(t *testing.T) {
	limiter := NewMemoryLimiter(3, time.Minute)
	ctx := context.Background()

	for i := 0; i < 3; i++ {
		allowed, err := limiter.Allow(ctx, "127.0.0.1")
		if err != nil {
			t.Fatalf("allow %d: %v", i, err)
		}
		if !allowed {
			t.Fatalf("expected request %d to be allowed", i+1)
		}
	}

	allowed, err := limiter.Allow(ctx, "127.0.0.1")
	if err != nil {
		t.Fatalf("allow blocked: %v", err)
	}
	if allowed {
		t.Fatal("expected fourth request to be blocked")
	}
}

func TestMemoryLimiterIsolatesKeys(t *testing.T) {
	limiter := NewMemoryLimiter(1, time.Minute)
	ctx := context.Background()

	allowed, err := limiter.Allow(ctx, "127.0.0.1")
	if err != nil || !allowed {
		t.Fatalf("expected first ip request allowed, got allowed=%v err=%v", allowed, err)
	}

	allowed, err = limiter.Allow(ctx, "127.0.0.2")
	if err != nil || !allowed {
		t.Fatalf("expected second ip request allowed, got allowed=%v err=%v", allowed, err)
	}
}
