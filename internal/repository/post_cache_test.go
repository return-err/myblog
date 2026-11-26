package repository

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/return-err/myblog/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupRedis(t *testing.T) *redis.Client {
	t.Helper()
	ctx := context.Background()

	c, err := tcredis.Run(ctx, "redis:7")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { c.Terminate(ctx) })

	uri, err := c.ConnectionString(ctx)
	if err != nil {
		t.Fatal(err)
	}
	opt, err := redis.ParseURL(uri)
	if err != nil {
		t.Fatal(err)
	}

	rdb := redis.NewClient(opt)
	t.Cleanup(func() { rdb.Close() })
	return rdb
}

func TestSetupRedis(t *testing.T) {
	rdb := setupRedis(t)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}
}

func TestCachedGetPostDetail_WritesCache(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := setupRedis(t)
	repo := NewPostRepository(testDB)
	cached := NewCachedPostRepository(repo, rdb)

	err := initAdmin(testDB, config.AdminConfig{Username: "test_user", Password: "test_pass"})
	require.NoError(t, err)

	err = repo.CreatePost(1, "test_post", "this is a test post")
	require.NoError(t, err)

	post, err := cached.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), post.ID)

	n, err := rdb.Exists(context.Background(), "post:1").Result()
	require.NoError(t, err)
	assert.Equal(t, int64(1), n)
}

func TestCachedGetPostDetail_ReturnsCachedValue(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := setupRedis(t)
	repo := NewPostRepository(testDB)
	cached := NewCachedPostRepository(repo, rdb)

	err := initAdmin(testDB, config.AdminConfig{Username: "", Password: ""})
	require.NoError(t, err)
	insertSQL := `
		INSERT INTO posts (user_id, title, content)
		VALUES ($1, $2, $3)
	`

	_, err = testDB.Exec(insertSQL, 1, "test_before", "this is a test")
	require.NoError(t, err)

	post, err := cached.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "test_before", post.Title)
	err = repo.EditPost(1, "test_after", "this is a test")
	require.NoError(t, err)

	fromDB, err := repo.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "test_after", fromDB.Title)

	post, err = cached.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "test_before", post.Title)
}
