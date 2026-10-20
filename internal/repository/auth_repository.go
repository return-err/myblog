package repository

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateUserWithProfile(username, passwordHash string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	insertSQL := `
		INSERT INTO users (username, password_hash)
		VALUES ($1, $2)
		RETURNING id
	`

	var userID int64
	err = tx.QueryRow(insertSQL, username, passwordHash).Scan(&userID)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return apperror.ErrAlreadyExists
		}

		return fmt.Errorf("failed to insert user: %w", err)
	}

	insertProfileSQL := `
		INSERT INTO user_profiles (user_id, display_name, bio)
		VALUES ($1, $2, $3)
	`
	defaultName := fmt.Sprintf("ユーザー %d", userID)
	_, err = tx.Exec(insertProfileSQL, userID, defaultName, "")
	if err != nil {
		return fmt.Errorf("failed to insert user profile: %w", err)
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (r *AuthRepository) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	selectSQL := `SELECT id, username, password_hash, role, token_version FROM users WHERE username = $1`

	err := r.db.QueryRow(selectSQL, username).Scan(
		&user.ID,
		&user.Username,
		&user.PasswordHash,
		&user.Role,
		&user.TokenVersion,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to query user by username: %w", err)
	}

	return &user, nil
}

func (r *AuthRepository) GetRoleByUserID(userID int64) (string, error) {
	var role string

	err := r.db.QueryRow(`
		SELECT role
		FROM users
		WHERE id = $1
	`, userID).Scan(&role)
	if err != nil {
		return "", fmt.Errorf("failed to query role by user id: %w", err)
	}

	return role, nil
}

func (r *AuthRepository) GetTokenVersionByUserID(userID int64) (int, error) {
	var tokenVersion int

	err := r.db.QueryRow(`
		SELECT token_version
		FROM users
		WHERE id = $1	
	`, userID).Scan(&tokenVersion)
	if err != nil {
		return 0, fmt.Errorf("failed to query token version by user id: %w", err)
	}

	return tokenVersion, nil
}

func (r *AuthRepository) SaveRefreshToken(userID int64, refreshToken string, expiresAt time.Time) error {
	insertSQL := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(insertSQL, userID, refreshToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to insert refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) FindRefreshToken(token string) (*model.RefreshToken, error) {
	var refreshToken model.RefreshToken

	err := r.db.QueryRow(`
		SELECT id, user_id, token, revoked, expires_at
		FROM refresh_tokens
		WHERE token = $1
	`, token).Scan(
		&refreshToken.ID,
		&refreshToken.UserID,
		&refreshToken.Token,
		&refreshToken.Revoked,
		&refreshToken.ExpiresAt,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query token: %w", err)
	}

	return &refreshToken, nil
}

func (r *AuthRepository) RotateRefreshToken(oldTokenID int, userID int64, refreshToken string, expiresAt time.Time) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to created transaction: %w", err)
	}
	defer tx.Rollback()

	result, err := tx.Exec(`
		UPDATE refresh_tokens
		SET revoked = true
		WHERE id = $1 AND user_id = $2 AND revoked = FALSE
	`, oldTokenID, userID)
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}
	if rows != 1 {
		return apperror.ErrRefreshTokenRevoked
	}

	if _, err := tx.Exec(`
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`, userID, refreshToken, expiresAt); err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit refresh token: %w", err)
	}

	return nil
}

func (r *AuthRepository) RevokeAllRefreshTokens(userID int64) error {
	_, err := r.db.Exec(`
		UPDATE refresh_tokens
		SET revoked = TRUE	
		WHERE user_id = $1 AND revoked = FALSE
	`, userID)
	return err
}
