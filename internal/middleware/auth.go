package middleware

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/handle"
	"github.com/return-err/myblog/internal/service"
)

type Middleware struct {
	jwtService *service.JWTService
	authRepo   service.AuthRepositoryInterface
}

func NewMiddleware(jwtService *service.JWTService, authRepo service.AuthRepositoryInterface) *Middleware {
	return &Middleware{
		jwtService: jwtService,
		authRepo:   authRepo,
	}
}

func (m *Middleware) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization が見つかりませんでした。"})
			c.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "求められる Token ではありません。"})
			c.Abort()
			return
		}

		tokenString := parts[1]

		claims, err := m.jwtService.ParseToken(tokenString)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "求められる Token ではありません。"})
			c.Abort()
			return
		}

		userID, err := strconv.ParseInt(claims.Subject, 10, 64)
		if err != nil {
			handle.RespondWithError(c, err)
			c.Abort()
			return
		}

		dbTokenVersion, err := m.authRepo.GetTokenVersionByUserID(userID)
		if err != nil || claims.TokenVersion != dbTokenVersion {
			handle.RespondWithError(c, apperror.ErrUnauthorized)
			c.Abort()
			return
		}

		c.Set("userID", userID)
		c.Set("role", claims.Role)
		c.Next()
	}
}

func (m *Middleware) RequireRole(requiredRole string) gin.HandlerFunc {
	return func(c *gin.Context) {
		roleVal, exists := c.Get("role")
		if !exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "認証情報が見つかりません。"})
			c.Abort()
			return
		}

		role, ok := roleVal.(string)
		if !ok || role != requiredRole {
			c.JSON(http.StatusForbidden, gin.H{"error": "アクセス権限がありません。"})
			c.Abort()
			return
		}

		c.Next()
	}
}

func (m *Middleware) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) == 2 && parts[0] == "Bearer" {
			tokenString := parts[1]
			claims, err := m.jwtService.ParseToken(tokenString)
			if err != nil {
				c.Next()
				return
			}

			userID, err := strconv.ParseInt(claims.Subject, 10, 64)
			if err != nil {
				c.Next()
				return
			}

			tokenVersion, err := m.authRepo.GetTokenVersionByUserID(userID)
			if err != nil || claims.TokenVersion != tokenVersion {
				c.Next()
				return
			}

			c.Set("userID", userID)
			c.Set("role", claims.Role)
		}
		c.Next()
	}
}
