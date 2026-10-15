package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type Claims struct {
	Role         string `json:"role"`
	TokenVersion int    `json:"token_version"`

	jwt.RegisteredClaims // 匿名フィールド メリット：埋め込み構造体のすべてのフィールドとメソッドを自動的に引き継ぐことができる。
}

type TokenPair struct {
	AccessToken  string
	RefreshToken string
}

type RefreshToken struct {
	ID        int
	UserID    int64
	Token     string
	Revoked   bool
	ExpiresAt time.Time
}
