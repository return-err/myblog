package handle

import (
	"errors"
	"log"
	"net/http"

	"github.com/Rashomon-code/myblog/internal/apperror"
	"github.com/gin-gonic/gin"
)

func RespondWithError(c *gin.Context, err error) {
	// 400
	if errors.Is(err, apperror.ErrUsernameInvalidLength) ||
		errors.Is(err, apperror.ErrUsernameContainsSpace) ||
		errors.Is(err, apperror.ErrInvalidRole) ||
		errors.Is(err, apperror.ErrSelfRoleChange) ||
		errors.Is(err, apperror.ErrInvalidInput) ||
		errors.Is(err, apperror.ErrInvalidTitle) {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// 401
	if errors.Is(err, apperror.ErrInvalidCredentials) ||
		errors.Is(err, apperror.ErrUnauthorized) ||
		errors.Is(err, apperror.ErrInvalidRefreshToken) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// 403
	if errors.Is(err, apperror.ErrForbidden) {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	// 404
	if errors.Is(err, apperror.ErrUserNotFound) ||
		errors.Is(err, apperror.ErrPostNotFound) {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// 409
	if errors.Is(err, apperror.ErrAlreadyExists) {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	//500
	log.Printf("[Internal Server Error] %v", err)
	c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
}
