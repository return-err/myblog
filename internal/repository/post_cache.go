package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/return-err/myblog/internal/model"
)

type CachedPostRepository struct {
	*PostRepository
	rdb *redis.Client
}

func postKey(postID int64) string {
	return fmt.Sprintf("post:%d", postID)
}

func NewCachedPostRepository(repo *PostRepository, rdb *redis.Client) *CachedPostRepository {
	return &CachedPostRepository{PostRepository: repo, rdb: rdb}
}

func (c *CachedPostRepository) GetPostDetail(postID int64) (model.PostDetail, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()

	data, err := c.rdb.Get(ctx, postKey(postID)).Bytes()
	if err == nil {
		var cached model.PostDetail
		if json.Unmarshal(data, &cached) == nil {
			return cached, nil
		}
	} else if err != redis.Nil {
		slog.Warn("redis get failed", "key", postKey(postID), "err", err)
	}

	p, err := c.PostRepository.GetPostDetail(postID)
	if err != nil {
		return model.PostDetail{}, err
	}

	if b, err := json.Marshal(p); err != nil {
		if err := c.rdb.Set(ctx, postKey(postID), b, 5*time.Minute).Err(); err != nil {
			slog.Warn("redis set failed", "key", postKey(postID), "err", err)
		}
	}

	return p, nil
}

func (c *CachedPostRepository) EditPost(postID int64, title, content string) error {
	if err := c.PostRepository.EditPost(postID, title, content); err != nil {
		return err
	}
	c.invalidate(postID)
	return nil
}

func (c *CachedPostRepository) DeletePost(postID int64) error {
	if err := c.PostRepository.DeletePost(postID); err != nil {
		return err
	}
	c.invalidate(postID)
	return nil
}

func (c *CachedPostRepository) invalidate(postID int64) {
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	err := c.rdb.Del(ctx, postKey(postID)).Err()
	if err != nil {
		slog.Warn("redis del failed", "key", postKey(postID), "err", err)
	}
}
