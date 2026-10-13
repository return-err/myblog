package repository

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/moby/moby/api/types/network"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
)

var testDB *sql.DB

func setupTestDB(ctx context.Context) (func(), error) {
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		testcontainers.WithWaitStrategy(
			//初期化して一回再起動されます、一回しかチャックしない場合、 ready の同時に stop されます
			// 13:03:52 🔔 Container is ready: ae49e7e12d95
			// 13:03:52 🐳 Stopping container: ae49e7e12d95
			// wait.ForLog("database system is ready to accept connections").
			// 	WithOccurrence(2).
			// 	WithStartupTimeout(30*time.Second),
			wait.ForSQL("5432/tcp", "pgx", func(host string, port network.Port) string {
				return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", "testuser", "testpass", host, port.Port(), "testdb")
			}).WithStartupTimeout(30*time.Second).WithPollInterval(500*time.Millisecond),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("Postgres コンテナー起動できませんでした: %w", err)
	}

	host, _ := pgContainer.Host(ctx)
	port, _ := pgContainer.MappedPort(ctx, "5432")

	os.Setenv("DB_HOST", host)
	os.Setenv("DB_PORT", port.Port())
	os.Setenv("DB_USER", "testuser")
	os.Setenv("DB_PASSWORD", "testpass")
	os.Setenv("DB_NAME", "testdb")
	os.Setenv("INITIAL_ADMIN_USER", "admin")
	os.Setenv("INITIAL_ADMIN_PASS", "admin1234")

	return func() {
		if err := pgContainer.Terminate(ctx); err != nil {
			log.Fatalf("クリアできませんでした: %v", err)
		}
	}, nil
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	teardown, err := setupTestDB(ctx)
	if err != nil {
		log.Fatalf("Test container 起動できませんでした: %v", err)
	}

	testDB, err = InitAPP()
	if err != nil {
		teardown()
		log.Fatalf("テストデータベース初期化できませんでした: %v", err)
	}

	code := m.Run()

	if testDB != nil {
		testDB.Close()
	}
	teardown()

	os.Exit(code) //os.Exit は defer を執行しない！
}

func resetTable(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec("TRUNCATE TABLE user_profiles, users, posts, refresh_tokens RESTART IDENTITY CASCADE;")
	require.NoError(t, err, "リセットできませんでした")
}
