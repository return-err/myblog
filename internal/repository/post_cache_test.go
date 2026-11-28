package repository

import (
	"context"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/config"
	"github.com/return-err/myblog/internal/testutil"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSetupRedis(t *testing.T) {
	rdb := testutil.SetupRedis(t)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		t.Fatal(err)
	}
}

func TestCachedGetPostDetail_WritesCache(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := testutil.SetupRedis(t)
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

	rdb := testutil.SetupRedis(t)
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

func TestCachedEditPost_InvalidatesCache(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := testutil.SetupRedis(t)
	repo := NewPostRepository(testDB)
	cached := NewCachedPostRepository(repo, rdb)

	if err := seedUser(t); err != nil {
		t.Fatal(err)
	}
	if err := seedPost(t, "title"); err != nil {
		t.Fatal(err)
	}
	post, err := cached.GetPostDetail(1)
	require.NoError(t, err)
	n, err := rdb.Exists(context.Background(), "post:1").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	err = cached.EditPost(1, "edited_title", "edited_context")
	require.NoError(t, err)

	n, err = rdb.Exists(context.Background(), "post:1").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), n)

	post, err = cached.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "edited_title", post.Title)
}

func TestCachedDeletePost_InvalidatesCache(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := testutil.SetupRedis(t)
	repo := NewPostRepository(testDB)
	cached := NewCachedPostRepository(repo, rdb)

	if err := seedUser(t); err != nil {
		t.Fatal(err)
	}
	if err := seedPost(t, "title"); err != nil {
		t.Fatal(err)
	}
	_, err := cached.GetPostDetail(1)
	require.NoError(t, err)
	n, err := rdb.Exists(context.Background(), "post:1").Result()
	require.NoError(t, err)
	require.Equal(t, int64(1), n)

	err = cached.DeletePost(1)
	require.NoError(t, err)

	n, err = rdb.Exists(context.Background(), "post:1").Result()
	require.NoError(t, err)
	require.Equal(t, int64(0), n)

	_, err = cached.GetPostDetail(1)
	assert.ErrorIs(t, err, apperror.ErrPostNotFound)
}

func TestCachedGetPostDetail_FallsBackWhenRedisDown(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	rdb := redis.NewClient(&redis.Options{Addr: "localhost:1"})
	t.Cleanup(func() { rdb.Close() })
	repo := NewPostRepository(testDB)
	cached := NewCachedPostRepository(repo, rdb)

	if err := seedUser(t); err != nil {
		t.Fatal(err)
	}
	if err := seedPost(t, "title"); err != nil {
		t.Fatal(err)
	}

	post, err := cached.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "title", post.Title)

	err = cached.EditPost(1, "edited_title", "")
	require.NoError(t, err)

	fromDB, err := repo.GetPostDetail(1)
	require.NoError(t, err)
	assert.Equal(t, "edited_title", fromDB.Title)
}
