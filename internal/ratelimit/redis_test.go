package ratelimit

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisLimiterAllowsUpToLimit(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	limiter := NewRedisLimiter(client, "auth", 3, time.Minute)
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

func TestRedisLimiterSharesStateAcrossInstances(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	limiterA := NewRedisLimiter(client, "auth", 2, time.Minute)
	limiterB := NewRedisLimiter(client, "auth", 2, time.Minute)
	ctx := context.Background()

	if allowed, err := limiterA.Allow(ctx, "10.0.0.1"); err != nil || !allowed {
		t.Fatalf("first allow via A: allowed=%v err=%v", allowed, err)
	}
	if allowed, err := limiterB.Allow(ctx, "10.0.0.1"); err != nil || !allowed {
		t.Fatalf("second allow via B: allowed=%v err=%v", allowed, err)
	}
	if allowed, err := limiterA.Allow(ctx, "10.0.0.1"); err != nil || allowed {
		t.Fatalf("third allow should be blocked: allowed=%v err=%v", allowed, err)
	}
}
