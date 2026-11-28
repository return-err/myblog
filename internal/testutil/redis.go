package testutil

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func SetupRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	c, err := tcredis.Run(ctx, "redis:7")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if termErr := c.Terminate(ctx); termErr != nil {
			t.Errorf("failed to terminate redis container: %v", termErr)
		}
	})

	uri, err := c.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}

	opt, err := redis.ParseURL(uri)
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(opt)
	t.Cleanup(func() {
		if closeErr := rdb.Close(); closeErr != nil {
			t.Errorf("failed to close redis client: %v", closeErr)
		}
	})

	return rdb
}
