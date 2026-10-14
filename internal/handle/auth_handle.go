package handle

import (
	"net/http"

	"github.com/Rashomon-code/myblog/internal/apperror"
	"github.com/Rashomon-code/myblog/internal/model"
	"github.com/gin-gonic/gin"
)

type AuthHandle struct {
	authService AuthService
}

func NewAuthHandle(s AuthService) *AuthHandle {
	return &AuthHandle{authService: s}
}

type AuthService interface {
	Register(username, password string) error
	Login(username, password string) (*model.TokenPair, error)
	Refresh(refreshToken string) (*model.TokenPair, error)
}

func (a *AuthHandle) Register(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	err := a.authService.Register(req.Username, req.Password)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "user registered successfully"})
}

func (a *AuthHandle) Login(c *gin.Context) {
	var req model.LoginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	token, err := a.authService.Login(req.Username, req.Password)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.SetCookie("refresh_token", token.RefreshToken, 7*24*3600, "/api/auth/refresh", "", false, true)
	c.JSON(http.StatusOK, gin.H{"message": "login successfully", "token": token.AccessToken})
}

func (a *AuthHandle) Refresh(c *gin.Context) {
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		RespondWithError(c, err)
		return
	}

	token, err := a.authService.Refresh(refreshToken)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.SetCookie("refresh_token", token.RefreshToken, 7*24*3600, "/api/auth/refresh", "", false, true)
	c.JSON(http.StatusOK, gin.H{"token": token.AccessToken})
}
