package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type PostRepository struct {
	db *sql.DB
}

func NewPostRepository(db *sql.DB) *PostRepository {
	return &PostRepository{db: db}
}

func (r *PostRepository) CreatePost(userID int64, title string, content string) error {
	insertSQL := `
		INSERT INTO posts (user_id, title, content)
		VALUES ($1, $2, $3)	
	`
	_, err := r.db.Exec(insertSQL, userID, title, content)
	if err != nil {
		return fmt.Errorf("failed to insert post for user %d: %w", userID, err)
	}

	return nil
}

func (r *PostRepository) GetTitleByUserID(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error) {
	var totalCount int64
	countSQL := `SELECT COUNT(*) FROM posts WHERE user_id = $1`
	err := r.db.QueryRow(countSQL, userID).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count posts by user %d: %w", userID, err)
	}

	offset := (page - 1) * pageSize

	selectSQL := `
		SELECT id, title, created_at
		FROM posts
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(selectSQL, userID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query posts by user %d: %w", userID, err)
	}

	posts, err := scanArticleSummaries(rows)
	if err != nil {
		return nil, 0, err
	}
	return posts, totalCount, nil
}

func (r *PostRepository) GetPostDetail(postID int64) (model.PostDetail, error) {
	selectSQL := `
		SELECT p.id, p.title, p.content, p.user_id, p.created_at, up.display_name
		FROM posts p
		LEFT JOIN user_profiles up ON p.user_id = up.user_id
		WHERE p.id = $1
	`

	var p model.PostDetail
	row := r.db.QueryRow(selectSQL, postID)
	err := row.Scan(&p.ID, &p.Title, &p.Content, &p.UserID, &p.CreatedAt, &p.DisplayName)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return model.PostDetail{}, apperror.ErrPostNotFound
		}
		return model.PostDetail{}, fmt.Errorf("failed to scan post detail for post %d: %w", postID, err)
	}

	return p, nil
}

func (r *PostRepository) DeletePost(postID int64) error {
	deleteSQL := `DELETE FROM posts WHERE id = $1`

	result, err := r.db.Exec(deleteSQL, postID)
	if err != nil {
		return fmt.Errorf("failed to execute delete post %d: %w", postID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected on delete post %d: %w", postID, err)
	}
	if rows == 0 {
		return apperror.ErrPostNotFound
	}

	return nil
}

func (r *PostRepository) EditPost(postID int64, title string, content string) error {
	updateSQL := `UPDATE posts SET title = $1, content = $2 WHERE id = $3`

	result, err := r.db.Exec(updateSQL, title, content, postID)
	if err != nil {
		return fmt.Errorf("failed to execute update post %d: %w", postID, err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected on update post %d: %w", postID, err)
	}
	if rows == 0 {
		return apperror.ErrPostNotFound
	}

	return nil
}

func (r *PostRepository) GetAllPosts(page, pageSize int) ([]model.ArticleSummary, int64, error) {
	offset := (page - 1) * pageSize

	var totalCount int64
	countSQL := `SELECT COUNT(*) FROM posts`
	err := r.db.QueryRow(countSQL).Scan(&totalCount)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count all posts %w", err)
	}

	selectSQL := `SELECT id, title, created_at FROM posts ORDER BY created_at DESC LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(selectSQL, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query all posts: %w", err)
	}

	posts, err := scanArticleSummaries(rows)
	if err != nil {
		return nil, 0, err
	}

	return posts, totalCount, nil
}

func (r *PostRepository) SearchPost(keyword string) ([]model.ArticleSummary, error) {
	selectSQL := `SELECT id, title, created_at FROM posts WHERE title LIKE '%' || $1 || '%' ORDER BY created_at DESC`

	rows, err := r.db.Query(selectSQL, keyword)
	if err != nil {
		return nil, fmt.Errorf("failed to search posts with keyword '%s': %w", keyword, err)
	}
	posts, err := scanArticleSummaries(rows)
	if err != nil {
		return nil, err
	}

	return posts, nil
}

func scanArticleSummaries(rows *sql.Rows) ([]model.ArticleSummary, error) {
	defer rows.Close()

	posts := make([]model.ArticleSummary, 0)
	for rows.Next() {
		var a model.ArticleSummary
		if err := rows.Scan(&a.ID, &a.Title, &a.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan article summary row: %w", err)
		}
		posts = append(posts, a)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	return posts, nil
}
