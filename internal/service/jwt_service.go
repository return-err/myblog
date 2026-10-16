package service

import (
	"fmt"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/return-err/myblog/internal/model"
)

type JWTService struct {
	Secret string
}

func NewJWTService(secret string) *JWTService {
	return &JWTService{Secret: secret}
}

func (j *JWTService) GenerateToken(userID int64, role string, tokenVersion int) (string, error) {
	claims := model.Claims{
		Role:         role,
		TokenVersion: tokenVersion,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(userID, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(15 * time.Minute)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	tokenString, err := token.SignedString(
		[]byte(j.Secret),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create token: %w", err)
	}

	return tokenString, nil
}

func (j *JWTService) ParseToken(tokenString string) (model.Claims, error) {
	var claims model.Claims
	token, err := jwt.ParseWithClaims(tokenString, &claims, func(t *jwt.Token) (any, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			// jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}) 同じ効果？
			return nil, fmt.Errorf("HMAC が間違っています: %v", t.Header["alg"])
		}
		return []byte(j.Secret), nil
	})
	if err != nil {
		return model.Claims{}, fmt.Errorf("Token 解析できませんでした: %w", err)
	}
	if !token.Valid {
		return model.Claims{}, fmt.Errorf("無効な JWT: %w", err)
	}

	return claims, nil

	//戻す際数字はデフォルトのfloat (jwt.MapClaims{} を使用する場合)
	// userid, ok := claims["id"].(float64)
	// if !ok {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"エラー": "ログイン中にエラーが起きました。"})
	// 	c.Abort()
	// 	return
	// }
	// username, ok := claims["username"].(string)
	// if !ok {
	// 	c.JSON(http.StatusUnauthorized, gin.H{"エラー": "ユーザーが見つかりませんでした。"})
	// 	c.Abort()
	// 	return
	// }
}
