package service

import (
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

type mockAuthRepository struct {
	AuthRepositoryInterface

	createUserFn         func(username, passwordHash string) error
	getUserFn            func(username string) (*model.User, error)
	saveTokenFn          func(userID int64, refreshToken string, expiresAt time.Time) error
	findTokenFn          func(token string) (*model.RefreshToken, error)
	getRoleFn            func(userID int64) (string, error)
	rotateRefreshTokenFn func(oldTokenID int, userID int64, refreshToken string, expiredAt time.Time) error
}

func (r *mockAuthRepository) CreateUserWithProfile(username, passwordHash string) error {
	return r.createUserFn(username, passwordHash)
}

func (r *mockAuthRepository) GetUserByUsername(username string) (*model.User, error) {
	return r.getUserFn(username)
}

func (r *mockAuthRepository) SaveRefreshToken(userID int64, refreshToken string, expiresAt time.Time) error {
	return r.saveTokenFn(userID, refreshToken, expiresAt)
}

func (r *mockAuthRepository) FindRefreshToken(token string) (*model.RefreshToken, error) {
	return r.findTokenFn(token)
}

func (r *mockAuthRepository) GetRoleByUserID(userID int64) (string, error) {
	return r.getRoleFn(userID)
}

func (r *mockAuthRepository) RotateRefreshToken(oldTokenID int, userID int64, refreshToken string, expiredAt time.Time) error {
	return r.rotateRefreshTokenFn(oldTokenID, userID, refreshToken, expiredAt)
}

func TestRegister(t *testing.T) {
	tests := []struct {
		name             string
		username         string
		password         string
		mockCreateUserFn func(username, passwordHash string) error
		wantErr          error
	}{
		{
			name:     "success",
			username: "testuser",
			password: "test123456",
			mockCreateUserFn: func(username, passwordHash string) error {
				return nil
			},
			wantErr: nil,
		},
		{
			name:             "invalid username space",
			username:         " wrong ",
			password:         "test123456",
			mockCreateUserFn: nil,
			wantErr:          apperror.ErrUsernameContainsSpace,
		},
		{
			name:             "invalid username",
			username:         "wr",
			password:         "test123456",
			mockCreateUserFn: nil,
			wantErr:          apperror.ErrUsernameInvalidLength,
		},
		{
			name:     "repository error",
			username: "testuser",
			password: "test123456",
			mockCreateUserFn: func(username, passwordHash string) error {
				return apperror.ErrDatabase
			},
			wantErr: apperror.ErrDatabase,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockAuthRepository{
				createUserFn: tt.mockCreateUserFn,
			}
			jwtService := NewJWTService("test")
			mockS := NewAuthService(r, jwtService)

			err := mockS.Register(tt.username, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestLogin(t *testing.T) {
	tests := []struct {
		name            string
		username        string
		password        string
		mockGetUserfn   func(username string) (*model.User, error)
		mockSaveTokenFn func(userID int64, refreshToken string, expiresAt time.Time) error
		wantErr         error
	}{
		{
			name:     "success",
			username: "testuser",
			password: "123456",
			mockGetUserfn: func(username string) (*model.User, error) {
				hash, err := bcrypt.GenerateFromPassword([]byte("123456"), 4)
				if err != nil {
					t.Fatal(err)
				}
				return &model.User{
					ID:           100,
					Username:     "testuser",
					PasswordHash: string(hash),
					Role:         "admin",
				}, nil
			},
			mockSaveTokenFn: func(userID int64, refreshToken string, expiresAt time.Time) error {
				if userID != 100 {
					return errors.New("failed to save token")
				}
				return nil
			},
			wantErr: nil,
		},
		{
			name: "wrong user",
			mockGetUserfn: func(username string) (*model.User, error) {
				return nil, apperror.ErrInvalidCredentials
			},
			wantErr: apperror.ErrInvalidCredentials,
		},
		{
			name:     "wrong password",
			username: "testuser",
			password: "123456",
			mockGetUserfn: func(username string) (*model.User, error) {
				hash, err := bcrypt.GenerateFromPassword([]byte("654321"), 4)
				if err != nil {
					t.Fatal(err)
				}
				return &model.User{
					ID:           100,
					Username:     "testuser",
					PasswordHash: string(hash),
					Role:         "admin",
				}, apperror.ErrInvalidCredentials
			},
			wantErr: apperror.ErrInvalidCredentials,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockAuthRepository{
				getUserFn:   tt.mockGetUserfn,
				saveTokenFn: tt.mockSaveTokenFn,
			}
			jwtService := NewJWTService("test")
			mockS := NewAuthService(r, jwtService)

			_, err := mockS.Login(tt.username, tt.password)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("expected error %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestRefresh(t *testing.T) {
	tests := []struct {
		name                 string
		oldRefreshToken      string
		findTokenFn          func(token string) (*model.RefreshToken, error)
		getRoleFn            func(userID int64) (string, error)
		rotateRefreshTokenFn func(oldTokenID int, userID int64, refreshToken string, expiredAt time.Time) error
		wantErr              error
	}{
		{
			name:            "success",
			oldRefreshToken: "old refresh token",
			findTokenFn: func(token string) (*model.RefreshToken, error) {
				if token != "old refresh token" {
					return nil, errors.New("failed to find refresh token")
				}
				return &model.RefreshToken{
					ID:        1,
					UserID:    100,
					Token:     "test",
					Revoked:   false,
					ExpiresAt: time.Now().Add(1 * time.Minute),
				}, nil
			},
			getRoleFn: func(userID int64) (string, error) {
				return "user", nil
			},
			rotateRefreshTokenFn: func(oldTokenID int, userID int64, refreshToken string, expiredAt time.Time) error {
				require.Equal(t, 1, oldTokenID)
				require.Equal(t, int64(100), userID)
				require.NotEmpty(t, refreshToken)
				require.True(t, expiredAt.After(time.Now()))
				return nil
			},
			wantErr: nil,
		},
		{
			name:            "error: token not found",
			oldRefreshToken: "non token",
			findTokenFn: func(token string) (*model.RefreshToken, error) {
				return nil, sql.ErrNoRows
			},
			wantErr: apperror.ErrInvalidRefreshToken,
		},
		{
			name:            "error: token expired",
			oldRefreshToken: "expired token",
			findTokenFn: func(token string) (*model.RefreshToken, error) {
				return &model.RefreshToken{
					ID:        1,
					UserID:    100,
					Token:     "test",
					Revoked:   false,
					ExpiresAt: time.Now().Add(-1 * time.Minute),
				}, nil
			},
			wantErr: apperror.ErrRefreshTokenExpired,
		},
		{
			name:            "error: token revoked",
			oldRefreshToken: "revoked token",
			findTokenFn: func(token string) (*model.RefreshToken, error) {
				return &model.RefreshToken{
					ID:        1,
					UserID:    100,
					Token:     "test",
					Revoked:   true,
					ExpiresAt: time.Now().Add(1 * time.Minute),
				}, nil
			},
			wantErr: apperror.ErrRefreshTokenRevoked,
		},
		{
			name:            "error: get role failed",
			oldRefreshToken: "valid token",
			findTokenFn: func(token string) (*model.RefreshToken, error) {
				return &model.RefreshToken{
					ID:        1,
					UserID:    100,
					Token:     "test",
					Revoked:   false,
					ExpiresAt: time.Now().Add(1 * time.Minute),
				}, nil
			},
			getRoleFn: func(userID int64) (string, error) {
				return "", errors.New("failed to get role")
			},
			wantErr: errors.New("failed to get role"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := &mockAuthRepository{
				findTokenFn:          tt.findTokenFn,
				getRoleFn:            tt.getRoleFn,
				rotateRefreshTokenFn: tt.rotateRefreshTokenFn,
			}
			jwtService := NewJWTService("test")
			mockS := NewAuthService(r, jwtService)

			tokenPair, err := mockS.Refresh(tt.oldRefreshToken)

			if tt.wantErr != nil {
				assert.Nil(t, tokenPair)
				assert.True(t, errors.Is(err, tt.wantErr) || err.Error() == tt.wantErr.Error())
			} else {
				claims, err := mockS.jwt.ParseToken(tokenPair.AccessToken)
				require.NoError(t, err)
				assert.Equal(t, int64(100), claims.UserID)
				assert.Equal(t, "user", claims.Role)
			}
		})
	}
}

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  error
	}{
		{
			name:     "success",
			username: "test_user",
			wantErr:  nil,
		},
		{
			name:     "username has space",
			username: "test user",
			wantErr:  apperror.ErrUsernameContainsSpace,
		},
		{
			name:     "invalid length",
			username: "this_is_a_invalid_username",
			wantErr:  apperror.ErrUsernameInvalidLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUsername(tt.username)
			if err != tt.wantErr {
				t.Errorf("expected %q, got %q", tt.wantErr, err)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "success",
			password: "testpass1",
			wantErr:  nil,
		},
		{
			name:     "missing digit",
			password: "testpass",
			wantErr:  apperror.ErrPasswordWeak,
		},
		{
			name:     "missing letter",
			password: "12345678",
			wantErr:  apperror.ErrPasswordWeak,
		},
		{
			name:     "invalid length",
			password: "test123",
			wantErr:  apperror.ErrPasswordInvalidLength,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("expected %q, got %q", tt.wantErr, err)
			}
		})
	}
}
