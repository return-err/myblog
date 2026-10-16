package middleware

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/service"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(mw gin.HandlerFunc, handler func(c *gin.Context)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(mw)

	r.GET("/test", handler)

	return r
}

type mockAuthRepo struct {
	service.AuthRepositoryInterface
}

func (m *mockAuthRepo) GetTokenVersionByUserID(userID int64) (int, error) {
	if userID != int64(100) {
		return 0, errors.New("user not found")
	}
	return 1, nil
}

func newTestMiddleware(secret string) *Middleware {
	jwtService := service.NewJWTService(secret)
	var mockRepo mockAuthRepo
	mw := NewMiddleware(jwtService, &mockRepo)
	return mw
}

func TestAuthMiddleware_Success(t *testing.T) {
	mw := newTestMiddleware("test")

	handlerReached := false

	handler := func(c *gin.Context) {
		handlerReached = true
		userID, exists := c.Get("userID")
		//userID.(int64) != 100 ok, userID != int64(100) ok, but userID != 100 fail
		if !exists || userID != int64(100) {
			c.Status(http.StatusInternalServerError)
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "user" {
			c.Status(http.StatusInternalServerError)
			return
		}

		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.AuthMiddleware(), handler)

	tokenString, _ := mw.jwtService.GenerateToken(int64(100), "user", 1)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerReached)
}

func TestAuthMiddleware_MissingHeader(t *testing.T) {
	handlerReached := false
	mw := newTestMiddleware("test")
	handler := func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.AuthMiddleware(), handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil) // "/test" 定義されていない Path にアクセスしようとしても、Middleware に経由します
	// req.Header.Set("", "")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerReached)
}

func TestAuthMiddleware_InvalidFormat(t *testing.T) {
	mw := newTestMiddleware("test")
	handlerReached := false
	handler := func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	}
	r := setupTestRouter(mw.AuthMiddleware(), handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer this is an expected 401 test header")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerReached)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	mw := newTestMiddleware("test")
	wrongJWTService := service.NewJWTService("wrong")
	tokenString, _ := wrongJWTService.GenerateToken(int64(100), "user", 1)
	handlerReached := false
	handler := func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.AuthMiddleware(), handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerReached) //assert.Equal(t, handlerReached, false)と同じ

	// if handlerReached {
	// 	t.Errorf("expected handlerReached false, got %v", handlerReached)
	// }
	// if w.Code != http.StatusUnauthorized {
	// 	t.Errorf("expected status %v, got %d", http.StatusUnauthorized, w.Code)
	// }
	//どちらも内容を比較し、結果を t に記録します
}

func TestRequireRole_Success(t *testing.T) {
	handlerReached := false
	mw := newTestMiddleware("test")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "admin")
		c.Next()
	})
	r.Use(mw.RequireRole("admin"))
	r.GET("/role", func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/role", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, handlerReached)
}

func TestRequireRole_Forbidden(t *testing.T) {
	handlerReached := false
	mw := newTestMiddleware("test")
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set("role", "user")
		c.Next()
	})
	r.Use(mw.RequireRole("admin"))
	r.GET("/role", func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/role", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.False(t, handlerReached)
}

func TestRequireRole_MissingRoleContext(t *testing.T) {
	handlerReached := false
	mw := newTestMiddleware("test")

	handler := func(c *gin.Context) {
		handlerReached = true
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.RequireRole("admin"), handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.False(t, handlerReached)
}

func TestOptionalAuthMiddleware_ValidToken(t *testing.T) {
	handlerReached := false
	mw := newTestMiddleware("test")
	handler := func(c *gin.Context) {
		handlerReached = true
		userID, exists := c.Get("userID")
		if !exists || userID != int64(100) {
			c.Status(http.StatusInternalServerError)
			return
		}

		role, exists := c.Get("role")
		if !exists || role != "user" {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.OptionalAuthMiddleware(), handler)

	tokenString, _ := mw.jwtService.GenerateToken(int64(100), "user", 1)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, handlerReached)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestOptionalAuthMiddleware_NoHeader(t *testing.T) {
	handlerReached := false
	var userID int64
	mw := newTestMiddleware("test")
	handler := func(c *gin.Context) {
		handlerReached = true
		userIDAny, exists := c.Get("userID")
		if exists {
			userID = userIDAny.(int64)
		}
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.OptionalAuthMiddleware(), handler)
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, handlerReached)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(0), userID)
}

func TestOptionalAuthMiddleware_InvalidToken(t *testing.T) {
	handlerReached := false
	var userID int64
	mw := newTestMiddleware("test")
	handler := func(c *gin.Context) {
		handlerReached = true
		userIDAny, exists := c.Get("userID")
		if exists {
			userID = userIDAny.(int64)
		}
		c.Status(http.StatusOK)
	}

	r := setupTestRouter(mw.OptionalAuthMiddleware(), handler)

	tokenString := "this.is.a.test.token" //期限切れtokenを作成してもいい、かりのclaimsから書く必要があります
	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("Authorization", "Bearer "+tokenString)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	assert.True(t, handlerReached)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, int64(0), userID)
}
