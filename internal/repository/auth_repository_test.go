package repository

import (
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/return-err/myblog/internal/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUserWithProfile(t *testing.T) {
	repo := NewAuthRepository(testDB)

	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := repo.CreateUserWithProfile("testuser", "hashed_pw")
		require.NoError(t, err)

		var userID int64
		err = testDB.QueryRow(`
			SELECT id FROM users WHERE username = 'testuser'	
		`).Scan(&userID)
		require.NoError(t, err)

		var displayName, bio string
		err = testDB.QueryRow(`
			SELECT display_name, bio FROM user_profiles WHERE user_id = $1	
		`, userID).Scan(&displayName, &bio)
		require.NoError(t, err)

		assert.Equal(t, fmt.Sprintf("ユーザー %d", userID), displayName)
		assert.Equal(t, "", bio)
	})

	t.Run("false", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := repo.CreateUserWithProfile("testuser", "hashed1")
		require.NoError(t, err)

		err = repo.CreateUserWithProfile("testuser", "hashed2")
		require.Error(t, err)

		var count int
		err = testDB.QueryRow("SELECT COUNT(*) FROM users").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)

		err = testDB.QueryRow("SELECT COUNT(*) FROM user_profiles").Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 1, count)
	})
}

func TestGetUserByUsername(t *testing.T) {
	repo := NewAuthRepository(testDB)

	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := initAdmin(testDB)
		require.NoError(t, err)

		user, err := repo.GetUserByUsername("admin")
		require.NoError(t, err)

		assert.Equal(t, int64(1), user.ID)
		assert.Equal(t, "admin", user.Role)

		err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte("admin1234"))
		assert.NoError(t, err)
	})

	t.Run("no user", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		_, err := repo.GetUserByUsername("admin")
		assert.Error(t, err)
	})
}

func TestGetRoleUserID(t *testing.T) {
	repo := NewAuthRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := initAdmin(testDB)
		require.NoError(t, err)
		role, err := repo.GetRoleByUserID(1)
		require.NoError(t, err)
		assert.Equal(t, "admin", role)

		err = repo.CreateUserWithProfile("test", "pass")
		require.NoError(t, err)
		role, err = repo.GetRoleByUserID(2)
		require.NoError(t, err)
		assert.Equal(t, "user", role)
	})

	t.Run("user not found", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		role, err := repo.GetRoleByUserID(1)
		assert.Error(t, err)
		assert.Equal(t, "", role)
	})
}

func TestSaveRefreshToken(t *testing.T) {
	repo := NewAuthRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := createTestUser(int64(100), testDB)
		require.NoError(t, err)

		expiresAT := time.Now().Add(24 * time.Hour)
		err = repo.SaveRefreshToken(100, "token", expiresAT)
		require.NoError(t, err)

		var userID int64
		var token string
		selectSQL := `SELECT user_id, token FROM refresh_tokens WHERE id = 1`
		err = testDB.QueryRow(selectSQL).Scan(&userID, &token)
		require.NoError(t, err)

		assert.Equal(t, int64(100), userID)
		assert.Equal(t, "token", token)
	})
}

func TestFindRefreshToken(t *testing.T) {
	repo := NewAuthRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := createTestUser(int64(100), testDB)
		require.NoError(t, err)

		expiresAT := time.Now().Add(24 * time.Hour)
		err = repo.SaveRefreshToken(100, "token", expiresAT)
		require.NoError(t, err)
		err = repo.UseRefreshToken(1)
		require.NoError(t, err)
		refreshToken, err := repo.FindRefreshToken("token")
		require.NoError(t, err)

		assert.Equal(t, 1, refreshToken.ID)
		assert.Equal(t, int64(100), refreshToken.UserID)
		assert.Equal(t, true, refreshToken.Revoked)
	})

	t.Run("token not found", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		refreshToken, err := repo.FindRefreshToken("non existent token")
		assert.Error(t, err)
		assert.Nil(t, refreshToken)
	})
}

func TestUseRefreshToken(t *testing.T) {
	repo := NewAuthRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := createTestUser(int64(100), testDB)
		require.NoError(t, err)

		err = repo.SaveRefreshToken(100, "token", time.Now())
		require.NoError(t, err)
		err = repo.UseRefreshToken(1)
		require.NoError(t, err)
		refreshToken, err := repo.FindRefreshToken("token")
		require.NoError(t, err)
		assert.True(t, refreshToken.Revoked)
	})
}

func TestRotateRefreshToken(t *testing.T) {
	repo := NewAuthRepository(testDB)
	tests := []struct {
		name           string
		oldTokenID     int
		userID         int64
		refreshToken   string
		revoked        bool
		wantErr        error
		wantOldRevoked bool
		wantNewToken   bool
	}{
		{
			name:           "success",
			oldTokenID:     1,
			userID:         1,
			refreshToken:   "new token",
			wantErr:        nil,
			wantOldRevoked: true,
			wantNewToken:   true,
		},
		{
			name:         "failed token already revoked",
			oldTokenID:   1,
			userID:       1,
			refreshToken: "new token",
			revoked:      true,
			wantErr:      apperror.ErrRefreshTokenRevoked,
			wantNewToken: false,
		},
		{
			name:         "failed token not found",
			oldTokenID:   100,
			userID:       1,
			refreshToken: "new token",
			wantErr:      apperror.ErrRefreshTokenRevoked,
			wantNewToken: false,
		},
		{
			name:         "failed wrong user",
			oldTokenID:   1,
			userID:       100,
			refreshToken: "new token",
			wantErr:      apperror.ErrRefreshTokenRevoked,
			wantNewToken: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTable(t)
			defer resetTable(t)

			err := createTestUser(int64(1), testDB)

			err = repo.SaveRefreshToken(1, "old token", time.Now().Add(10*time.Minute))
			require.NoError(t, err)

			if tt.revoked {
				err := repo.UseRefreshToken(1)
				require.NoError(t, err)
			}

			err = repo.RotateRefreshToken(tt.oldTokenID, tt.userID, tt.refreshToken, time.Now().Add(10*time.Minute))
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}

			_, err = repo.FindRefreshToken(tt.refreshToken)
			if tt.wantNewToken {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}

			if tt.wantOldRevoked {
				var revoked bool
				err := testDB.QueryRow(`
					SELECT revoked
					FROM refresh_tokens	
					WHERE id = $1
				`, tt.oldTokenID).Scan(&revoked)

				require.NoError(t, err)
				require.True(t, revoked)
			}
		})
	}
}

func TestRevokeAllRefreshTokens(t *testing.T) {
	repo := NewAuthRepository(testDB)
	t.Run("success", func(t *testing.T) {
		resetTable(t)
		defer resetTable(t)

		err := createTestUser(int64(100), testDB)
		require.NoError(t, err)

		for i := range 5 {
			_, err := testDB.Exec(`
				INSERT INTO refresh_tokens (user_id, token, expires_at)	
				VALUES ($1, $2, $3)
			`, 100, fmt.Sprintf("token%d", i), time.Now().Add(10*time.Minute))
			require.NoError(t, err)
		}

		err = repo.RevokeAllRefreshTokens(int64(100))
		require.NoError(t, err)

		var count int
		err = testDB.QueryRow(`
			SELECT COUNT (*) FROM refresh_tokens
			WHERE revoked = FALSE
		`).Scan(&count)
		require.NoError(t, err)
		assert.Equal(t, 0, count)
	})
}

func createTestUser(userID int64, db *sql.DB) error {
	insertSQL := `
		INSERT INTO users (id, username, password_hash, role)
		VALUES ($1, $2, $3, $4)
	`

	_, err := db.Exec(insertSQL, userID, "test user", "pass", "user")
	if err != nil {
		return fmt.Errorf("failed to create test user: %w", err)
	}
	return nil
}
