package repository

import (
	"testing"

	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserProfile(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	err := initAdmin(testDB)
	require.NoError(t, err)

	userRepo := NewUserRepository(testDB)

	t.Run("no displayname", func(t *testing.T) {
		userProfile, err := userRepo.GetUserProfile(1)
		require.NoError(t, err)
		assert.Equal(t, "ユーザー 1", userProfile.DisplayName)
	})

	t.Run("success", func(t *testing.T) {
		userRepo.UpdateUserProfile(1, "Admin", "")
		userProfile, err := userRepo.GetUserProfile(1)
		require.NoError(t, err)
		assert.Equal(t, "Admin", userProfile.DisplayName)
		assert.Equal(t, "", userProfile.Bio)
	})

	t.Run("no profile", func(t *testing.T) {
		userProfile, err := userRepo.GetUserProfile(2)
		require.NoError(t, err)
		assert.Equal(t, "ユーザー 2", userProfile.DisplayName)
		assert.Equal(t, "まだ何もありません", userProfile.Bio)
	})
}

func TestUpdateRole(t *testing.T) {
	authRepo := NewAuthRepository(testDB)
	userRepo := NewUserRepository(testDB)

	tests := []struct {
		name        string
		userID      int64
		wantErr     bool
		wantVersion int
	}{
		{
			name:        "success",
			userID:      1,
			wantVersion: 1,
		},
		{
			name:    "invalid user",
			userID:  10,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resetTable(t)
			defer resetTable(t)
			err := authRepo.CreateUserWithProfile("testuser", "hashedpass")
			require.NoError(t, err)

			err = userRepo.UpdateRole(tt.userID, "admin")
			if tt.wantErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)

				var tokenVersion int
				err := testDB.QueryRow(`
					SELECT token_version FROM users WHERE id = $1
				`, tt.userID).Scan(&tokenVersion)
				require.NoError(t, err)
				assert.Equal(t, tt.wantVersion, tokenVersion)
			}

		})
	}
}

func TestGetAllUsers(t *testing.T) {
	resetTable(t)
	defer resetTable(t)
	userRepo := NewUserRepository(testDB)

	t.Run("empty table", func(t *testing.T) {
		users, err := userRepo.GetAllUsers()
		require.NoError(t, err)
		assert.NotNil(t, users)
		assert.Len(t, users, 0)
		assert.Equal(t, []model.UserResponse{}, users)
	})

	t.Run("success", func(t *testing.T) {
		defer resetTable(t)

		_, err := testDB.Exec(`
			INSERT INTO users (id, username, password_hash, role) VALUES
			(2, 'user2', 'pass2', 'user'),
			(1, 'user1', 'pass1', 'admin'),
			(3, 'user3', 'pass3', 'user')
		`)
		require.NoError(t, err)

		users, err := userRepo.GetAllUsers()
		require.NoError(t, err)
		assert.Len(t, users, 3)

		assert.Equal(t, int64(1), users[0].ID)
		assert.Equal(t, "user1", users[0].Username)
		assert.Equal(t, "admin", users[0].Role)

		assert.Equal(t, int64(2), users[1].ID)
		assert.Equal(t, "user2", users[1].Username)

		assert.Equal(t, int64(3), users[2].ID)
		assert.Equal(t, "user3", users[2].Username)
	})
}

func TestUpdateUserProfile(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	userRepo := NewUserRepository(testDB)
	err := initAdmin(testDB)
	require.NoError(t, err)

	t.Run("create profile", func(t *testing.T) {
		err := userRepo.UpdateUserProfile(1, "Admin", "nothing")
		require.NoError(t, err)
		profile, err := userRepo.GetUserProfile(1)
		require.NoError(t, err)

		require.Equal(t, "Admin", profile.DisplayName)
		require.Equal(t, "nothing", profile.Bio)
	})

	t.Run("upadte profile", func(t *testing.T) {
		err := userRepo.UpdateUserProfile(1, "Super", "This is a admin account")
		require.NoError(t, err)
		profile, err := userRepo.GetUserProfile(1)
		require.NoError(t, err)

		assert.Equal(t, "Super", profile.DisplayName)
		assert.Equal(t, "This is a admin account", profile.Bio)
	})

	t.Run("invalid user", func(t *testing.T) {
		err := userRepo.UpdateUserProfile(2, "test", "test")
		require.Error(t, err)
	})
}
