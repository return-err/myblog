package repository

import (
	"testing"

	migrator "github.com/return-err/myblog/db"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestInitAPP_success(t *testing.T) {
	if testDB == nil {
		t.Fatalf("testDB is nil")
	}

	err := testDB.Ping()
	if err != nil {
		t.Fatalf("接続に失敗しました: %v", err)
	}
}

func TestInitAPP_CreatesTables(t *testing.T) {
	err := migrator.RunMigrations(testDB)
	require.NoError(t, err)

	tables := []string{"users", "posts", "user_profiles"}
	query := `SELECT EXISTS (
		SELECT FROM information_schema.tables
		WHERE table_schema = 'public'
		AND table_name = $1
	)`
	for _, table := range tables {
		t.Run(table, func(t *testing.T) {
			var exists bool
			row := testDB.QueryRow(query, table)
			err := row.Scan(&exists)
			require.NoError(t, err)
			assert.True(t, exists)
		})
	}
}

func TestInitAdmin(t *testing.T) {
	resetTable(t)
	defer resetTable(t)

	err := initAdmin(testDB)
	if err != nil {
		require.NoError(t, err, "admin 初期化できませんでした")
	}

	var username, role, passwordHash string
	var count int

	err = testDB.QueryRow(`
		SELECT username, role, password_hash
		FROM users
		WHERE role = 'admin'	
	`).Scan(&username, &role, &passwordHash)
	require.NoError(t, err, "検索できませんでした")

	assert.Equal(t, "admin", username)
	assert.Equal(t, "admin", role)
	err = bcrypt.CompareHashAndPassword([]byte(passwordHash), []byte("admin1234"))
	assert.NoError(t, err, "パスワードが正しくありません")

	err = initAdmin(testDB)
	require.NoError(t, err, "二回目の初期化に問題が起きました")

	err = testDB.QueryRow(`SELECT COUNT(*) FROM users WHERE role = 'admin'`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
