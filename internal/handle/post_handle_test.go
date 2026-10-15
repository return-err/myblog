package handle

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
)

type mockPostService struct {
	PostService

	called       bool
	createPostfn func(userID int64, title, content string) error
	postDetailfn func(postID int64) (model.PostDetail, error)
	deletePostfn func(postID, userID int64, userRole string) error
	editPostfn   func(postID, userID int64, title, content, userRole string) error
	getPostfn    func(page, pageSize int) ([]model.ArticleSummary, int64, error)
}

func (m *mockPostService) CreatePost(userID int64, title, content string) error {
	m.called = true
	return m.createPostfn(userID, title, content)
}

func (m *mockPostService) PostDetail(postID int64) (model.PostDetail, error) {
	m.called = true
	return m.postDetailfn(postID)
}

func (m *mockPostService) DeletePost(postID, userID int64, userRole string) error {
	m.called = true
	return m.deletePostfn(postID, userID, userRole)
}

func (m *mockPostService) EditPost(postID, userID int64, title, content, userRole string) error {
	m.called = true
	return m.editPostfn(postID, userID, title, content, userRole)
}

func (m *mockPostService) GetAllPosts(page, pageSize int) ([]model.ArticleSummary, int64, error) {
	m.called = true
	return m.getPostfn(page, pageSize)
}

func (m *mockPostService) SearchPost(keyword string) ([]model.ArticleSummary, error) {
	m.called = true
	if keyword == "error" {
		return nil, errors.New("server error")
	}
	return nil, nil
}

func (m *mockPostService) GetPostTitle(userID int64, page, pageSize int) ([]model.ArticleSummary, int64, error) {
	m.called = true
	return m.getPostfn(page, pageSize)
}

func TestCreatePost(t *testing.T) {
	tests := []struct {
		name         string
		wantCode     int
		reqBody      string
		wantCalled   bool
		middleware   func(*gin.Context)
		createPostfn func(userID int64, title, content string) error
	}{
		{
			name:       "success",
			wantCode:   http.StatusCreated,
			reqBody:    `{"title": "test post", "content": "test content"}`,
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			createPostfn: func(userID int64, title, content string) error {
				assert.Equal(t, int64(100), userID)
				assert.Equal(t, "test post", title)
				assert.Equal(t, "test content", content)
				return nil
			},
		},
		{
			name:       "missing user",
			wantCode:   http.StatusUnauthorized,
			wantCalled: false,
		},
		{
			name:     "missing request",
			wantCode: http.StatusBadRequest,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			wantCalled: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{
				createPostfn: tt.createPostfn,
			}
			postHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodPost, "/test", postHandle.CreatePost, tt.middleware)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.reqBody))
			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestPostDetail(t *testing.T) {
	tests := []struct {
		name         string
		wantCode     int
		paramID      string
		wantCalled   bool
		postDetailfn func(postID int64) (model.PostDetail, error)
	}{
		{
			name:       "success",
			wantCode:   http.StatusOK,
			paramID:    "100",
			wantCalled: true,
			postDetailfn: func(postID int64) (model.PostDetail, error) {
				return model.PostDetail{}, nil
			},
		},
		{
			name:     "bad id",
			wantCode: http.StatusBadRequest,
			paramID:  "a1",
		},
		{
			name:       "service error",
			wantCode:   http.StatusInternalServerError,
			paramID:    "100",
			wantCalled: true,
			postDetailfn: func(postID int64) (model.PostDetail, error) {
				return model.PostDetail{}, errors.New("service err")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{
				postDetailfn: tt.postDetailfn,
			}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodGet, "/test/:id", mockHandle.PostDetail)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test/"+tt.paramID, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestDeletePost(t *testing.T) {
	tests := []struct {
		name         string
		wantCode     int
		paramID      string
		wantCalled   bool
		middleware   func(*gin.Context)
		deletePostfn func(postID, userID int64, userRole string) error
	}{
		{
			name:       "success",
			wantCode:   http.StatusOK,
			paramID:    "100",
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			deletePostfn: func(postID, userID int64, userRole string) error {
				return nil
			},
		},
		{
			name:     "bad post id",
			wantCode: http.StatusBadRequest,
			paramID:  "a1",
		},
		{
			name:       "service error",
			wantCode:   http.StatusInternalServerError,
			paramID:    "100",
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			deletePostfn: func(postID, userID int64, userRole string) error {
				return errors.New("service error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{
				deletePostfn: tt.deletePostfn,
			}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodDelete, "/test/:id", mockHandle.DeletePost, tt.middleware)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodDelete, "/test/"+tt.paramID, nil)

			r.ServeHTTP(w, req)

			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestEditPost(t *testing.T) {
	tests := []struct {
		name       string
		wantCode   int
		paramID    string
		reqBody    string
		wantCalled bool
		middleware func(*gin.Context)
		editPostfn func(postID, userID int64, title, content, userRole string) error
	}{
		{
			name:       "success",
			wantCode:   http.StatusOK,
			paramID:    "100",
			reqBody:    `{"title": "title", "content": "content"}`,
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Set("role", "user")
				c.Next()
			},
			editPostfn: func(postID, userID int64, title, content, userRole string) error {
				return nil
			},
		},
		{
			name:     "bad post id",
			wantCode: http.StatusBadRequest,
			paramID:  "abc",
		},
		{
			name:     "bad requsert",
			wantCode: http.StatusBadRequest,
			paramID:  "100",
		},
		{
			name:       "server error",
			wantCode:   http.StatusInternalServerError,
			paramID:    "100",
			reqBody:    `{"title": "title", "content": "content"}`,
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Set("role", "user")
				c.Next()
			},
			editPostfn: func(postID, userID int64, title, content, userRole string) error {
				return errors.New("server error")
			},
		},
		{
			name:       "not fount",
			wantCode:   http.StatusNotFound,
			paramID:    "100",
			reqBody:    `{"title": "title", "content": "content"}`,
			wantCalled: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Set("role", "user")
				c.Next()
			},
			editPostfn: func(postID, userID int64, title, content, userRole string) error {
				return apperror.ErrUserNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{editPostfn: tt.editPostfn}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodPut, "/test/:id", mockHandle.EditPost)
			req := httptest.NewRequest(http.MethodPut, "/test/"+tt.paramID, strings.NewReader(tt.reqBody))
			req.Header.Set("Content-Type", "application/json") // 指定しなければ、なんでも解析に問題が発生しません
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestPostsList(t *testing.T) {
	tests := []struct {
		name       string
		wantCode   int
		page       string
		pageSize   string
		wantCalled bool
		getPostfn  func(page, pagesize int) ([]model.ArticleSummary, int64, error)
	}{
		{
			name:       "success",
			wantCode:   http.StatusOK,
			page:       "page=2",
			pageSize:   "page_size=20",
			wantCalled: true,
			getPostfn: func(page, pageSize int) ([]model.ArticleSummary, int64, error) {
				if page == 2 && pageSize == 20 {
					return nil, 0, nil
				}
				return nil, 0, errors.New("expected no error")
			},
		},
		{
			name:       "success no page detail",
			wantCode:   http.StatusOK,
			wantCalled: true,
			getPostfn: func(page, pageSize int) ([]model.ArticleSummary, int64, error) {
				if page == 1 && pageSize == 10 {
					return nil, 0, nil
				}
				return nil, 0, errors.New("expected no error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{getPostfn: tt.getPostfn}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodGet, "/test", mockHandle.PostsList)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test?"+tt.page+"&"+tt.pageSize, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestSearchPost(t *testing.T) {
	tests := []struct {
		name       string
		wantCalled bool
		wantCode   int
		keyword    string
	}{
		{
			name:       "success empty",
			keyword:    "keyword=",
			wantCalled: false,
			wantCode:   http.StatusOK,
		},
		{
			name:       "success",
			keyword:    "keyword=keyword",
			wantCalled: true,
			wantCode:   http.StatusOK,
		},
		{
			name:       "server error",
			keyword:    "keyword=error",
			wantCalled: true,
			wantCode:   http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodGet, "/test", mockHandle.SearchPost)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test?"+tt.keyword, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestGetUserPosts(t *testing.T) {
	tests := []struct {
		name       string
		wantCode   int
		wantCalled bool
		id         string
		page       string
		pageSize   string
		getPostfn  func(page, pageSize int) ([]model.ArticleSummary, int64, error)
	}{
		{
			name:       "success",
			wantCode:   http.StatusOK,
			wantCalled: true,
			id:         "100",
			page:       "page=2",
			pageSize:   "page_size=20",
			getPostfn: func(page, pageSize int) ([]model.ArticleSummary, int64, error) {
				if page != 2 || pageSize != 20 {
					return nil, 0, errors.New("expected no error")
				}
				return nil, 0, nil
			},
		},
		{
			name:       "success default",
			wantCode:   http.StatusOK,
			wantCalled: true,
			id:         "100",
			getPostfn: func(page, pageSize int) ([]model.ArticleSummary, int64, error) {
				if page != 1 || pageSize != 10 {
					return nil, 0, errors.New("expected no error")
				}
				return nil, 0, nil
			},
		},
		{
			name:     "bad user",
			id:       "abc",
			wantCode: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockPostService{getPostfn: tt.getPostfn}
			mockHandle := NewPostHandle(mockSvc)
			r := setupTestRouter(http.MethodGet, "test/:id", mockHandle.GetUserPosts)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test/"+tt.id+"?"+tt.page+"&"+tt.pageSize, nil)
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCalled, mockSvc.called)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}
