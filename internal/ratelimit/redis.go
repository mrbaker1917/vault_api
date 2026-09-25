package ratelimit

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisLimiter is a distributed fixed-window limiter backed by Redis INCR + EXPIRE.
type RedisLimiter struct {
	client *redis.Client
	prefix string
	limit  int64
	window time.Duration
}

func NewRedisLimiter(client *redis.Client, prefix string, limit int, window time.Duration) *RedisLimiter {
	return &RedisLimiter{
		client: client,
		prefix: prefix,
		limit:  int64(limit),
		window: window,
	}
}

func (l *RedisLimiter) Allow(ctx context.Context, key string) (bool, error) {
	redisKey := fmt.Sprintf("ratelimit:%s:%s", l.prefix, key)

	count, err := l.client.Incr(ctx, redisKey).Result()
	if err != nil {
		return false, fmt.Errorf("redis incr: %w", err)
	}
	if count == 1 {
		if err := l.client.Expire(ctx, redisKey, l.window).Err(); err != nil {
			return false, fmt.Errorf("redis expire: %w", err)
		}
	}

	return count <= l.limit, nil
}
