package service

import (
	"fmt"
	"strings"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type PostRepository interface {
	CreatePost(userID int64, title string, content string) error
	GetTitleByUserID(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error)
	GetPostDetail(postID int64) (model.PostDetail, error)
	DeletePost(postID int64) error
	EditPost(postID int64, title string, content string) error
	GetAllPosts(page, pageSize int) ([]model.ArticleSummary, int64, error)
	SearchPost(keyword string) ([]model.ArticleSummary, error)
}

type PostService struct {
	repo PostRepository
}

func NewPostService(repo PostRepository) *PostService {
	return &PostService{repo: repo}
}

func validateTitle(title string) (string, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return "", apperror.ErrInvalidTitle
	}
	return title, nil
}

func (s *PostService) CreatePost(userID int64, title, content string) error {
	title, err := validateTitle(title)
	if err != nil {
		return err
	}

	err = s.repo.CreatePost(userID, title, content)
	if err != nil {
		return err
	}

	return nil
}

func (s *PostService) GetPostTitle(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error) {
	if userID <= 0 {
		return []model.ArticleSummary{}, 0, apperror.ErrUserNotFound
	}
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}
	return s.repo.GetTitleByUserID(userID, page, pageSize)
}

func (s *PostService) PostDetail(postID int64) (model.PostDetail, error) {
	post, err := s.repo.GetPostDetail(postID)
	if err != nil {
		return model.PostDetail{}, nil
	}

	if post.DisplayName == nil || *post.DisplayName == "" {
		displayname := fmt.Sprintf("ユーザー %d", post.UserID)
		post.DisplayName = &displayname
	}

	return post, nil
}

func (s *PostService) DeletePost(postID, userID int64, userRole string) error {
	post, err := s.repo.GetPostDetail(postID)
	if err != nil {
		return err
	}

	if post.UserID != userID && userRole != "admin" {
		return apperror.ErrForbidden
	}

	return s.repo.DeletePost(postID)
}

func (s *PostService) EditPost(postID, userID int64, title, content, userRole string) error {
	post, err := s.repo.GetPostDetail(postID)
	if err != nil {
		return err
	}

	if post.UserID != userID && userRole != "admin" {
		return apperror.ErrForbidden
	}

	if title, err = validateTitle(title); err != nil {
		return err
	}

	return s.repo.EditPost(postID, title, content)
}

func (s *PostService) GetAllPosts(page, pageSize int) ([]model.ArticleSummary, int64, error) {
	return s.repo.GetAllPosts(page, pageSize)
}

func (s *PostService) SearchPost(keyword string) ([]model.ArticleSummary, error) {
	return s.repo.SearchPost(keyword)
}
