package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient(ctx context.Context, addr string) (*redis.Client, error) {
	rdb := redis.NewClient(&redis.Options{Addr: addr})

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()
	err := rdb.Ping(pingCtx).Err()
	if err != nil {
		rdb.Close()
		return nil, fmt.Errorf("redis ping: %w", err)
	}
	return rdb, nil
}
