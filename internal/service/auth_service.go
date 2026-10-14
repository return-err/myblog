package service

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"

	"github.com/Rashomon-code/myblog/internal/apperror"
	"github.com/Rashomon-code/myblog/internal/model"
	"golang.org/x/crypto/bcrypt"
)

type AuthRepositoryInterface interface {
	CreateUserWithProfile(username, passwordHash string) error
	GetUserByUsername(username string) (*model.User, error)
	SaveRefreshToken(userID int64, refreshToken string, expiresAt time.Time) error
	FindRefreshToken(token string) (*model.RefreshToken, error)
	GetRoleByUserID(userID int64) (string, error)
	RotateRefreshToken(oldTokenID int, userID int64, refreshToken string, expiredAt time.Time) error
}

type AuthService struct {
	repo AuthRepositoryInterface
	jwt  *JWTService
}

func NewAuthService(repo AuthRepositoryInterface, jwt *JWTService) *AuthService {
	return &AuthService{repo: repo, jwt: jwt}
}

func (s *AuthService) Register(username, password string) error {
	if err := validateUsername(username); err != nil {
		return err
	}
	if err := validatePassword(password); err != nil {
		return err
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.MinCost) //学習のため、最低レベルを使用します
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	err = s.repo.CreateUserWithProfile(username, string(passwordHash))
	return err
}

func (s *AuthService) Login(username, password string) (*model.TokenPair, error) {
	user, err := s.repo.GetUserByUsername(username)
	if err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, apperror.ErrInvalidCredentials
	}

	accessToken, err := s.jwt.GenerateToken(user.ID, user.Role)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.repo.SaveRefreshToken(user.ID, refreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) Refresh(refreshToken string) (*model.TokenPair, error) {
	token, err := s.repo.FindRefreshToken(refreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrInvalidRefreshToken
		}
		return nil, err
	}

	if token.ExpiresAt.Before(time.Now()) {
		return nil, apperror.ErrRefreshTokenExpired
	}

	if token.Revoked {
		return nil, apperror.ErrRefreshTokenRevoked
	}

	role, err := s.repo.GetRoleByUserID(token.UserID)
	if err != nil {
		return nil, err
	}

	accessToken, err := s.jwt.GenerateToken(token.UserID, role)
	if err != nil {
		return nil, err
	}

	newRefreshToken, expiresAt, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	err = s.repo.RotateRefreshToken(token.ID, token.UserID, newRefreshToken, expiresAt)
	if err != nil {
		return nil, err
	}

	return &model.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

func generateRefreshToken() (string, time.Time, error) {
	bytes := make([]byte, 32)
	_, err := rand.Read(bytes)
	if err != nil {
		return "", time.Now(), fmt.Errorf("failed to create refresh token: %w", err)
	}

	return hex.EncodeToString(bytes), time.Now().Add(7 * 24 * time.Hour), nil
}

// 正規表現で制限する方法
// var usernameRegex = regexp.MustCompile(`^[a-zA-Z0-9_]{3,20}$`)
// func validateUsernameFormat(username string) error {
// 	if !usernameRegex.MatchString(username) {
// 		return errors.New("Format error")
// 	}
// 	return nil
// }

func validateUsername(username string) error {
	if len(username) < 3 || len(username) > 20 {
		return apperror.ErrUsernameInvalidLength
	}

	if strings.IndexFunc(username, unicode.IsSpace) != -1 {
		return apperror.ErrUsernameContainsSpace
	}

	return nil
}

func validatePassword(password string) error {
	if len(password) < 8 || len(password) > 72 {
		return apperror.ErrPasswordInvalidLength
	}

	hasLetter := false
	hasDigit := false

	for _, ch := range password {
		if unicode.IsLetter(ch) {
			hasLetter = true
		}
		if unicode.IsDigit(ch) {
			hasDigit = true
		}
	}

	if hasLetter && hasDigit {
		return nil
	}
	return apperror.ErrPasswordWeak
}
