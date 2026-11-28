package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLimiter struct {
	rdb    *redis.Client
	prefix string
	limit  int64
	window time.Duration
}

func NewRedisLimiter(rdb *redis.Client, prefix string, limit int64, window time.Duration) *RedisLimiter {
	return &RedisLimiter{rdb: rdb, prefix: prefix, limit: limit, window: window}
}

func (l *RedisLimiter) Allow(key string) bool {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		200*time.Millisecond,
	)
	defer cancel()

	redisKey := l.prefix + key

	n, err := l.rdb.Incr(ctx, redisKey).Result()
	if err != nil {
		slog.Warn("rate limiter: redis INCR failed",
			"key", redisKey,
			"error", err,
		)
		return true
	}

	if n == 1 {
		if err := l.rdb.Expire(ctx, redisKey, l.window).Err(); err != nil {
			slog.Warn("rate limiter: redis EXPIRE failed",
				"key", redisKey,
				"error", err,
			)
		}
	}

	return n <= l.limit
}
