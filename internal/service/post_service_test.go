package service

import (
	"errors"
	"testing"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockPostRepository struct {
	PostRepository
	//関数を変数として struct に保存して仕様する前に実作します（関数も変数？関数ポインター？）
	GetPostDetailFunc func(postID int64) (model.PostDetail, error)
	DeletePostFunc    func(postID int64) error
	CreatePostFunc    func(userID int64, title string, content string) error
	EditPostFunc      func(postID int64, title string, content string) error

	called bool
}

func (m *mockPostRepository) DeletePost(postID int64) error {
	return m.DeletePostFunc(postID)
}

func (m *mockPostRepository) GetPostDetail(postID int64) (model.PostDetail, error) {
	return m.GetPostDetailFunc(postID)
}

func (m *mockPostRepository) CreatePost(userID int64, title string, content string) error {
	return m.CreatePostFunc(userID, title, content)
}

func (m *mockPostRepository) EditPost(postID int64, title string, content string) error {
	return m.EditPostFunc(postID, title, content)
}

func (m *mockPostRepository) GetTitleByUserID(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error) {
	m.called = true
	return []model.ArticleSummary{}, 0, nil
}

func TestValidateTitle(t *testing.T) {
	testCases := []string{"", " ", "\n", "  \n"}
	for _, tc := range testCases {
		_, err := validateTitle(tc)
		if !errors.Is(err, apperror.ErrInvalidTitle) {
			t.Errorf("expected %v, got %v", apperror.ErrInvalidTitle, err)
		}
	}
}

func TestCreatePost_InvalidTitle(t *testing.T) {

	service := NewPostService(nil)
	err := service.CreatePost(1, "", "content")
	if !errors.Is(err, apperror.ErrInvalidTitle) {
		t.Errorf("expected %v, got %v", apperror.ErrInvalidTitle, err)
	}
}

func TestCreatePost_RepoError(t *testing.T) {
	mockRepo := &mockPostRepository{
		CreatePostFunc: func(userID int64, title, content string) error {
			return errors.New("db error")
		},
	}

	service := NewPostService(mockRepo)
	err := service.CreatePost(1, "title", "content")
	if err == nil {
		t.Fatalf("expected error %q, got nil", "db error")
	}
	if err.Error() != "db error" {
		t.Errorf("expected error: %q, got %v", "db error", err)
	}
}

func TestCreatePost_Success(t *testing.T) {
	var (
		capturedUserID  int64
		capturedTitle   string
		capturedContent string
	)

	mockRepo := &mockPostRepository{
		CreatePostFunc: func(userID int64, title, content string) error {
			capturedUserID = userID
			capturedTitle = title
			capturedContent = content
			return nil
		},
	}

	service := NewPostService(mockRepo)
	err := service.CreatePost(100, " Test Post ", "Hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if capturedUserID != 100 {
		t.Errorf("expected userID 100, got '%d'", capturedUserID)
	}

	if capturedTitle != "Test Post" {
		t.Errorf("expected title %q, got '%s'", "Test Post", capturedTitle)
	}

	if capturedContent != "Hello" {
		t.Errorf("expected content %q, got '%s'", "Hello", capturedContent)
	}
}

func TestDeletePost_Success(t *testing.T) {
	mockRepo := &mockPostRepository{
		GetPostDetailFunc: func(postID int64) (model.PostDetail, error) {
			return model.PostDetail{
				ID:     postID,
				UserID: 100,
			}, nil
		},
		DeletePostFunc: func(postID int64) error {
			return nil
		},
	}

	service := NewPostService(mockRepo)
	err := service.DeletePost(1, 100, "user")
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestDeletePost_Forbidden(t *testing.T) {
	mockRepo := &mockPostRepository{
		GetPostDetailFunc: func(postID int64) (model.PostDetail, error) {
			return model.PostDetail{
				ID:     postID,
				UserID: 100,
			}, nil
		},
	}

	service := NewPostService(mockRepo)

	err := service.DeletePost(1, 200, "user")
	if !errors.Is(err, apperror.ErrForbidden) {
		t.Errorf("expected ErrForbidden, got %v", err)
	}
}

func TestEditPost_Success(t *testing.T) {
	var post = model.PostDetail{
		ID:      100,
		UserID:  100,
		Title:   "編集前タイトル",
		Content: "編集前",
	}

	mockRepo := &mockPostRepository{
		GetPostDetailFunc: func(postID int64) (model.PostDetail, error) {
			return post, nil
		},
		EditPostFunc: func(postID int64, title string, content string) error {
			post.Title = title
			post.Content = content
			return nil
		},
	}

	service := NewPostService(mockRepo)

	err := service.EditPost(100, 100, "編集後タイトル", "編集後", "user")
	if err != nil {
		t.Fatalf("unexpected err %v", err)
	}

	if post.Title != "編集後タイトル" {
		t.Errorf("expected %q, got %q", "編集後タイトル", post.Title)
	}

	if post.Content != "編集後" {
		t.Errorf("expected %q, got %q", "編集後", post.Content)
	}
}

func TestGetPostTitle(t *testing.T) {
	tests := []struct {
		name       string
		userID     int64
		page       int
		pageSize   int
		wantErr    bool
		wantCalled bool
	}{
		{
			name:       "success",
			userID:     100,
			page:       1,
			pageSize:   20,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "invalid userID",
			userID:     -100,
			page:       1,
			pageSize:   20,
			wantErr:    true,
			wantCalled: false,
		},
		{
			name:       "invalid page",
			userID:     100,
			page:       -1,
			pageSize:   20,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "invalid page size",
			userID:     100,
			page:       1,
			pageSize:   0,
			wantErr:    false,
			wantCalled: true,
		},
		{
			name:       "invalid page size (too big)",
			userID:     100,
			page:       1,
			pageSize:   101,
			wantErr:    false,
			wantCalled: true,
		},
	}
	for _, tt := range tests {
		mockService := NewPostService(&mockPostRepository{called: false})
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := mockService.GetPostTitle(tt.userID, tt.page, tt.pageSize)
			if tt.wantErr {
				require.Error(t, err)
				assert.False(t, tt.wantCalled)
			} else {
				require.NoError(t, err)
				assert.True(t, tt.wantCalled)
			}
		})
	}
}
