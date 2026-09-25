package sessioncache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

func TestCacheSetGetDelete(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	cache := New(client, DefaultTTL)
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(30 * time.Minute)

	if err := cache.Set(ctx, sessionID, userID, expiresAt); err != nil {
		t.Fatalf("set: %v", err)
	}

	gotUserID, ok, err := cache.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !ok {
		t.Fatal("expected cache hit")
	}
	if gotUserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, gotUserID)
	}

	if err := cache.Delete(ctx, sessionID); err != nil {
		t.Fatalf("delete: %v", err)
	}

	_, ok, err = cache.Get(ctx, sessionID)
	if err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if ok {
		t.Fatal("expected cache miss after delete")
	}
}

func TestCacheSetUsesRemainingSessionLifetime(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	cache := New(client, DefaultTTL)
	ctx := context.Background()
	sessionID := uuid.New()
	userID := uuid.New()
	expiresAt := time.Now().Add(2 * time.Minute)

	if err := cache.Set(ctx, sessionID, userID, expiresAt); err != nil {
		t.Fatalf("set: %v", err)
	}

	ttl := mr.TTL("session:active:" + sessionID.String())
	if ttl <= 0 || ttl > 2*time.Minute+time.Second {
		t.Fatalf("expected ttl near 2m, got %v", ttl)
	}
}
