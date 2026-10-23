package handle

import (
	"errors"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
)

func RespondWithError(c *gin.Context, err error) {
	if err == nil {
		return
	}

	switch {
	// 400 Bad Request
	case errors.Is(err, apperror.ErrInvalidInput),
		errors.Is(err, apperror.ErrUsernameInvalidLength),
		errors.Is(err, apperror.ErrUsernameContainsSpace),
		errors.Is(err, apperror.ErrPasswordInvalidLength),
		errors.Is(err, apperror.ErrPasswordWeak),
		errors.Is(err, apperror.ErrInvalidRole),
		errors.Is(err, apperror.ErrSelfRoleChange),
		errors.Is(err, apperror.ErrInvalidTitle):
		c.JSON(http.StatusBadRequest, gin.H{
			"error": err.Error(),
		})

	// 401 Unauthorized
	case errors.Is(err, apperror.ErrUnauthorized),
		errors.Is(err, apperror.ErrInvalidCredentials),
		errors.Is(err, apperror.ErrInvalidToken),
		errors.Is(err, apperror.ErrMissingToken),
		errors.Is(err, apperror.ErrInvalidRefreshToken),
		errors.Is(err, apperror.ErrRefreshTokenExpired),
		errors.Is(err, apperror.ErrRefreshTokenRevoked):
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})

	// 403 Forbidden
	case errors.Is(err, apperror.ErrForbidden):
		c.JSON(http.StatusForbidden, gin.H{
			"error": err.Error(),
		})

	// 404 Not Found
	case errors.Is(err, apperror.ErrUserNotFound),
		errors.Is(err, apperror.ErrPostNotFound):
		c.JSON(http.StatusNotFound, gin.H{
			"error": err.Error(),
		})

	// 409 Conflict
	case errors.Is(err, apperror.ErrAlreadyExists):
		c.JSON(http.StatusConflict, gin.H{
			"error": err.Error(),
		})

	// 429 Too Many Requests
	case errors.Is(err, apperror.ErrLoginRateLimit):
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error": err.Error(),
		})

	// 500 Internal Server Error
	default:
		log.Printf("[Internal Server Error] %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "internal server error",
		})
	}
}
