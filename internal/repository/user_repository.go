package repository

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetUserProfile(userID int64) (*model.UserProfile, error) {
	selectSQL := `SELECT display_name, bio FROM user_profiles WHERE user_id = $1`

	displayName := fmt.Sprintf("ユーザー %d", userID)
	bio := "まだ何もありません"

	row := r.db.QueryRow(selectSQL, userID)
	err := row.Scan(&displayName, &bio)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("failed to scan user profile: %w", err)
	}

	userProfile := model.UserProfile{
		UserID:      userID,
		DisplayName: displayName,
		Bio:         bio,
	}

	return &userProfile, nil
}

func (r *UserRepository) UpdateRole(userID int64, newRole string) error {
	updateSQL := `UPDATE users SET role = $1, token_version = token_version + 1 WHERE id = $2 RETURNING id`

	var returnedID int64
	err := r.db.QueryRow(updateSQL, newRole, userID).Scan(&returnedID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperror.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user role: %w", err)
	}

	return nil
}

func (r *UserRepository) GetAllUsers() ([]model.UserResponse, error) {
	selectSQL := `SELECT id, username, role FROM users ORDER BY id ASC`

	rows, err := r.db.Query(selectSQL)
	if err != nil {
		return nil, fmt.Errorf("failed to query users: %w", err)
	}
	defer rows.Close()

	users := make([]model.UserResponse, 0)

	for rows.Next() {
		var user model.UserResponse
		err := rows.Scan(&user.ID, &user.Username, &user.Role)
		if err != nil {
			return nil, fmt.Errorf("failed to scan user row: %w", err)
		}
		users = append(users, user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error during rows iteration: %w", err)
	}

	return users, nil
}

func (r *UserRepository) UpdateUserProfile(userID int64, displayName, bio string) error {
	updateSQL := `
		INSERT INTO user_profiles (user_id, display_name, bio)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id)
		DO UPDATE SET
			display_name = EXCLUDED.display_name,
			bio = EXCLUDED.bio
	`
	//INSERT で衝突した際、データは一時的に EXCLUDED に移動されます

	_, err := r.db.Exec(updateSQL, userID, displayName, bio)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return apperror.ErrUserNotFound
		}
		return fmt.Errorf("failed to update user profile: %w", err)
	}

	return nil
}

func (r *UserRepository) IncrementTokenVersion(userID int64) error {
	result, err := r.db.Exec(`
		UPDATE users
		SET token_version = token_version + 1
		WHERE id = $1	
	`, userID)
	if err != nil {
		return fmt.Errorf("failed to increment token version: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check rows affected: %w", err)
	}
	if rows == 0 {
		return errors.New("user not found")
	}

	return nil
}
