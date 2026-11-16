package service

import (
	"strconv"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/return-err/myblog/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testAccessTTL = 15 * time.Minute

func TestGenerateToken_Success(t *testing.T) {
	jwtService := NewJWTService("test", testAccessTTL)

	tokenString, err := jwtService.GenerateToken(int64(100), "admin", 0)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	if tokenString == "" {
		t.Error("expected non-empty, got empty")
	}
}

func TestGenerateToken_UsesAccessTTL(t *testing.T) {
	jwtService := NewJWTService("test", 30*time.Second)

	tokenString, err := jwtService.GenerateToken(int64(100), "admin", 0)
	require.NoError(t, err)

	claims, err := jwtService.ParseToken(tokenString)
	require.NoError(t, err)

	// 誤差を許容する
	assert.WithinDuration(t, time.Now().Add(30*time.Second), claims.ExpiresAt.Time, 2*time.Second)
}

func TestParseToken_Success(t *testing.T) {
	jwtService := NewJWTService("test", testAccessTTL)
	tokenString, err := jwtService.GenerateToken(int64(100), "admin", 100)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	parsedClaims, err := jwtService.ParseToken(tokenString)
	if err != nil {
		t.Fatalf("expected no err, got %v", err)
	}

	userID, err := strconv.ParseInt(parsedClaims.Subject, 10, 64)
	if userID != 100 {
		t.Errorf("expected UserID %d, got %d", 100, userID)
	}
	if parsedClaims.Role != "admin" {
		t.Errorf("expected Username %q, got %q", "admin", parsedClaims.Role)
	}
	if parsedClaims.TokenVersion != 100 {
		t.Errorf("expected token version %d, got %d", 100, parsedClaims.TokenVersion)
	}
}

func TestParseToken_InvalidSecret(t *testing.T) {
	jwtService := NewJWTService("correct", testAccessTTL)
	tokenString, _ := jwtService.GenerateToken(int64(100), "admin", 0)

	wrongService := NewJWTService("wrong", testAccessTTL)
	_, err := wrongService.ParseToken(tokenString)
	if err == nil {
		t.Error("expected err, got no err")
	}
}

func TestParseToken_Expired(t *testing.T) {
	expiredClaims := model.Claims{
		Role: "admin",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatInt(100, 10),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(-1 * time.Hour)),
		},
	}
	jwtService := NewJWTService("test", testAccessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, expiredClaims)
	tokenString, _ := token.SignedString([]byte("test"))

	_, err := jwtService.ParseToken(tokenString)
	if err == nil {
		t.Error("expected err, got no err")
	}
}
