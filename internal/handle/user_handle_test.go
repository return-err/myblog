package handle

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockUserService struct {
	UserService

	getProfilefn    func(userID int64) (*model.UserProfile, error)
	updateRolefn    func(operatorID, userID int64, newRole string) error
	updateProfilefn func(userID int64, displayName, bio string) error
	getAllUsersfn   func() ([]model.UserResponse, error)
}

func (m *mockUserService) GetProfile(userID int64) (*model.UserProfile, error) {
	return m.getProfilefn(userID)
}

func (m *mockUserService) UpdateRole(operatorID, userID int64, newRole string) error {
	return m.updateRolefn(operatorID, userID, newRole)
}

func (m *mockUserService) UpdateProfile(userID int64, displayName, bio string) error {
	return m.updateProfilefn(userID, displayName, bio)
}

func (m *mockUserService) GetAllUsers() ([]model.UserResponse, error) {
	return m.getAllUsersfn()
}

func TestMypage(t *testing.T) {
	tests := []struct {
		name         string
		wantCode     int
		middleware   func(c *gin.Context)
		getProfilefn func(userID int64) (*model.UserProfile, error)
	}{
		{
			name:     "success",
			wantCode: http.StatusOK,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				if userID != 100 {
					return nil, fmt.Errorf("unexpected userID: %d", userID)
				}
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "no token",
			wantCode: http.StatusUnauthorized,
			middleware: func(c *gin.Context) {
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "wrong user",
			wantCode: http.StatusNotFound,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(1000))
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				if userID != 100 {
					return nil, apperror.ErrUserNotFound
				}
				return &model.UserProfile{UserID: userID}, nil
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &mockUserService{getProfilefn: tt.getProfilefn}
			h := NewUserHandle(m)

			r := setupTestRouter(http.MethodGet, "/test", h.MyPage, tt.middleware)

			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			r.ServeHTTP(w, req)
			if w.Code != tt.wantCode {
				t.Errorf("expected sataus code %d, got %d, %s", tt.wantCode, w.Code, w.Body.String())
			}
		})
	}
}

func TestGetUserProfile(t *testing.T) {
	tests := []struct {
		name         string
		userID       int64
		paramID      string
		wantCode     int
		wantIsMe     bool
		middleware   func(c *gin.Context)
		getProfilefn func(userID int64) (*model.UserProfile, error)
	}{
		{
			name:     "success self",
			paramID:  "100",
			wantCode: http.StatusOK,
			wantIsMe: true,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "success orter",
			paramID:  "100",
			wantCode: http.StatusOK,
			wantIsMe: false,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(200))
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "invalid URL ID format",
			paramID:  "abc",
			wantCode: http.StatusUnauthorized,
			wantIsMe: false,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "success view",
			paramID:  "100",
			wantCode: http.StatusOK,
			wantIsMe: false,
			middleware: func(c *gin.Context) {
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return &model.UserProfile{UserID: userID}, nil
			},
		},
		{
			name:     "user not found",
			paramID:  "100",
			wantCode: http.StatusNotFound,
			wantIsMe: false,
			middleware: func(c *gin.Context) {
				c.Next()
			},
			getProfilefn: func(userID int64) (*model.UserProfile, error) {
				return nil, apperror.ErrUserNotFound
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{getProfilefn: tt.getProfilefn}
			h := NewUserHandle(mockSvc)

			r := setupTestRouter(http.MethodGet, "/test/:id", h.GetUserProfile, tt.middleware)

			req := httptest.NewRequest(http.MethodGet, "/test/"+tt.paramID, nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)

			require.Equal(t, tt.wantCode, w.Code, w.Body.String())
			if tt.wantCode == http.StatusOK {
				var res model.UserProfileResponse
				err := json.Unmarshal(w.Body.Bytes(), &res)
				require.NoError(t, err)
				assert.Equal(t, tt.wantIsMe, res.IsMe)
			}
		})
	}
}

func TestUpdateRole(t *testing.T) {
	tests := []struct {
		name         string
		paramID      string
		wantCode     int
		reqBody      string
		middleware   func(*gin.Context)
		updateRolefn func(operatorID, userID int64, newRole string) error
	}{
		{
			name:     "success",
			paramID:  "100",
			wantCode: http.StatusOK,
			reqBody:  `{ "role": "user" }`,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(200))
				c.Next()
			},
			updateRolefn: func(operatorID, userID int64, newRole string) error {
				assert.Equal(t, int64(200), operatorID)
				assert.Equal(t, int64(100), userID)
				assert.Equal(t, "user", newRole)
				return nil
			},
		},
		{
			name:     "invalid user ID",
			paramID:  "abc",
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "missing user ID",
			paramID:  "100",
			wantCode: http.StatusUnauthorized,
		},
		{
			name:    "invalid JSON",
			paramID: "100",
			reqBody: `{ "": "" }`,
			middleware: func(ctx *gin.Context) {
				ctx.Set("userID", int64(200))
				ctx.Next()
			},
			wantCode: http.StatusBadRequest,
		},
		{
			name:     "service error",
			paramID:  "100",
			wantCode: http.StatusInternalServerError,
			reqBody:  `{ "role": "user" }`,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(200))
				c.Next()
			},
			updateRolefn: func(operatorID, userID int64, newRole string) error {
				return errors.New("permission denied")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{
				updateRolefn: tt.updateRolefn,
			}
			h := NewUserHandle(mockSvc)
			r := setupTestRouter(http.MethodPut, "/test/:id", h.UpdateRole, tt.middleware)
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodPut, "/test/"+tt.paramID, strings.NewReader(tt.reqBody))
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestUpdateProfile(t *testing.T) {
	tests := []struct {
		name            string
		wantCode        int
		reqBody         string
		middleware      func(*gin.Context)
		updateProfilefn func(userID int64, displayName, bio string) error
	}{
		{
			name:     "success",
			wantCode: http.StatusOK,
			reqBody:  `{ "display_name": "test", "bio": "test" }`,
			middleware: func(ctx *gin.Context) {
				ctx.Set("userID", int64(100))
				ctx.Next()
			},
			updateProfilefn: func(userID int64, displayName, bio string) error {
				assert.Equal(t, int64(100), userID)
				assert.Equal(t, "test", displayName)
				assert.Equal(t, "test", bio)
				return nil
			},
		},
		{
			name:     "missing user id",
			wantCode: http.StatusUnauthorized,
			middleware: func(ctx *gin.Context) {
				ctx.Next()
			},
		},
		{
			name:     "missing request",
			wantCode: http.StatusBadRequest,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
		},
		{
			name:     "service error",
			wantCode: http.StatusInternalServerError,
			reqBody:  `{ "display_name": "test", "bio": "test" }`,
			middleware: func(c *gin.Context) {
				c.Set("userID", int64(100))
				c.Next()
			},
			updateProfilefn: func(userID int64, displayName, bio string) error {
				return errors.New("error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{updateProfilefn: tt.updateProfilefn}
			h := NewUserHandle(mockSvc)
			r := setupTestRouter(http.MethodPut, "/test", h.UpdateProfile, tt.middleware)
			req := httptest.NewRequest(http.MethodPut, "/test", strings.NewReader(tt.reqBody))
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)
		})
	}
}

func TestGetAllUser(t *testing.T) {
	tests := []struct {
		name          string
		wantCode      int
		wantID        int64
		getAllUsersfn func() ([]model.UserResponse, error)
	}{
		{
			name:     "success",
			wantCode: http.StatusOK,
			wantID:   int64(100),
			getAllUsersfn: func() ([]model.UserResponse, error) {
				response := []model.UserResponse{
					{ID: int64(100), Username: "test", Role: ""},
				}
				return response, nil
			},
		},
		{
			name:     "fail",
			wantCode: http.StatusInternalServerError,
			getAllUsersfn: func() ([]model.UserResponse, error) {
				return nil, errors.New("database error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockSvc := &mockUserService{getAllUsersfn: tt.getAllUsersfn}
			h := NewUserHandle(mockSvc)
			r := setupTestRouter(http.MethodGet, "/test", h.GetAllUsers)
			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.wantCode, w.Code)

			if w.Code == http.StatusOK {
				var got []model.UserResponse
				err := json.Unmarshal(w.Body.Bytes(), &got)

				require.NoError(t, err)
				require.Len(t, got, 1)
				assert.Equal(t, tt.wantID, got[0].ID)
			}
		})
	}
}
