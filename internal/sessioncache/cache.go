package sessioncache

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	keyPrefix  = "session:active:"
	DefaultTTL = 15 * time.Minute
)

// Cache stores active session IDs mapped to user IDs in Redis.
type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func New(client *redis.Client, ttl time.Duration) *Cache {
	if ttl <= 0 {
		ttl = DefaultTTL
	}
	return &Cache{
		client: client,
		ttl:    ttl,
	}
}

func (c *Cache) Get(ctx context.Context, sessionID uuid.UUID) (uuid.UUID, bool, error) {
	value, err := c.client.Get(ctx, keyPrefix+sessionID.String()).Result()
	if err == redis.Nil {
		return uuid.Nil, false, nil
	}
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("redis get session cache: %w", err)
	}

	userID, err := uuid.Parse(value)
	if err != nil {
		return uuid.Nil, false, fmt.Errorf("parse cached user id: %w", err)
	}
	return userID, true, nil
}

func (c *Cache) Set(ctx context.Context, sessionID, userID uuid.UUID, expiresAt time.Time) error {
	ttl := c.ttl
	if remaining := time.Until(expiresAt); remaining > 0 && remaining < ttl {
		ttl = remaining
	}
	if ttl <= 0 {
		return nil
	}

	if err := c.client.Set(ctx, keyPrefix+sessionID.String(), userID.String(), ttl).Err(); err != nil {
		return fmt.Errorf("redis set session cache: %w", err)
	}
	return nil
}

func (c *Cache) Delete(ctx context.Context, sessionID uuid.UUID) error {
	if err := c.client.Del(ctx, keyPrefix+sessionID.String()).Err(); err != nil {
		return fmt.Errorf("redis delete session cache: %w", err)
	}
	return nil
}
