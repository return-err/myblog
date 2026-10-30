package repository

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
	migrator "github.com/return-err/myblog/db"
	"golang.org/x/crypto/bcrypt"
)

func InitAPP() (*sql.DB, error) {
	db, err := NewDB()
	if err != nil {
		return nil, err
	}

	err = migrator.RunMigrations(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	err = initAdmin(db)
	if err != nil {
		db.Close()
		return nil, err
	}

	return db, nil
}

func NewDB() (*sql.DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", host, port, user, password, dbname)

	db, err := sql.Open("pgx", connStr)
	if err != nil {
		return nil, fmt.Errorf("データベースを開く際にエラーが発生しました: %w", err)
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(10)
	db.SetConnMaxLifetime(5 * time.Minute)

	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("database connection failed: %w", err)
	}
	fmt.Println("PostgreSQL に接続済み")

	return db, nil
}

func initAdmin(db *sql.DB) error {
	var count int
	selectSQL := `SELECT COUNT(*) FROM users WHERE role = 'admin'`

	row := db.QueryRow(selectSQL)
	err := row.Scan(&count)
	if err != nil {
		return err
	}

	if count == 0 {
		defaultUsername := os.Getenv("INITIAL_ADMIN_USER")
		defaultPassword := os.Getenv("INITIAL_ADMIN_PASS")

		passwordHash, err := bcrypt.GenerateFromPassword([]byte(defaultPassword), 4)
		if err != nil {
			return fmt.Errorf("管理者初期化失敗しました: %w", err)
		}

		initAdminSQL := `INSERT INTO users (username, password_hash, role) VALUES ($1, $2, $3)`
		_, err = db.Exec(initAdminSQL, defaultUsername, passwordHash, "admin")
		if err != nil {
			return fmt.Errorf("アカウント初期化失敗しました: %w", err)
		}

		log.Println("管理者アカウントが生成されました")
	}

	return nil
}
