package middleware

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/return-err/myblog/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRedisLimiter_BlockRequestOverLimit(t *testing.T) {
	rdb := testutil.SetupRedis(t)
	limiter := NewRedisLimiter(rdb, "test:", 3, time.Minute)

	for i := range 3 {
		assert.True(t, limiter.Allow("0.0.0.1"), "request %d should be allowed", i)
	}
	assert.False(t, limiter.Allow("0.0.0.1"), "request 4 should be blocked")
}

func TestRedisLimiter_KeysAreIndependent(t *testing.T) {
	rdb := testutil.SetupRedis(t)
	limiter := NewRedisLimiter(rdb, "test:", 3, time.Minute)

	for i := range 3 {
		assert.True(t, limiter.Allow("0.0.0.1"), "request %d should be allowed", i)
	}
	assert.False(t, limiter.Allow("0.0.0.1"), "request 4 should be blocked")
	assert.True(t, limiter.Allow("0.0.0.2"), "request 1 should be allowed")
}

func TestRedisLimiter_SetsExpiry(t *testing.T) {
	ctx := context.Background()
	rdb := testutil.SetupRedis(t)
	limiter := NewRedisLimiter(rdb, "test:", 3, time.Minute)

	assert.True(t, limiter.Allow("0.0.0.1"), "request should be allowed")
	ttl, err := rdb.TTL(ctx, "test:0.0.0.1").Result()
	require.NoError(t, err)
	assert.Greater(t, ttl, time.Duration(0), "TTL should be greater than zero")
	assert.InDelta(t, 60, ttl.Seconds(), 1, "TTL should be approximately one minute")
}

func TestRedisLimiter_ResetsAfterWindow(t *testing.T) {
	rdb := testutil.SetupRedis(t)
	limiter := NewRedisLimiter(rdb, "test:", 1, time.Second)

	assert.True(t, limiter.Allow("0.0.0.1"), "request should be allowed")
	assert.False(t, limiter.Allow("0.0.0.1"), "second request should be blocked")

	time.Sleep(1100 * time.Millisecond)
	assert.True(t, limiter.Allow("0.0.0.1"), "request should be allowed")
}

func TestRedisLimiter_FailsOpenWhenRedisDown(t *testing.T) {
	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})
	limiter := NewRedisLimiter(rdb, "test:", 1, time.Second)

	assert.True(t, limiter.Allow("0.0.0.1"), "should fail open")
	assert.True(t, limiter.Allow("0.0.0.1"), "should still fail open past the limit")
}
