package main

import (
	"fmt"
	"log"
	"math/rand/v2"
	"time"

	"github.com/joho/godotenv"
	"github.com/return-err/myblog/internal/repository"
)

func main() {
	godotenv.Load()
	db, err := repository.InitAPP()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	userID := 2
	fmt.Println("テスト用データ作成中 ...")

	tx, err := db.Begin()
	if err != nil {
		log.Fatal(err)
	}

	stmt, err := tx.Prepare(`
		INSERT INTO posts (title, content, user_id, created_at)
		VALUES ($1, $2, $3, $4)
	`) //SQL を用意する、 Prepare は分析するが、実行はしない
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

	topics := []string{"Go", "Gin", "PostgreSQL", "Docker", "Neovim", "Linux", "REST API"}

	for i := 0; i < 100; i++ {
		topic := topics[rand.N(len(topics))]
		title := fmt.Sprintf("「テスト用 %03d」%s勉強中", i+1, topic)
		content := fmt.Sprintf("テスト　%d", i+1)
		createdAt := time.Now()

		_, err := stmt.Exec(title, content, userID, createdAt)
		if err != nil {
			tx.Rollback()
			log.Fatalf("テスト　%d　取り込み失敗: %v", i, err)
		}
	}

	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}

	fmt.Println("テスト内容準備完了")
}
