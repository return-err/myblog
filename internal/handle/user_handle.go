package handle

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type UserService interface {
	GetProfile(userID int64) (*model.UserProfile, error)
	UpdateRole(operatorID, userID int64, newRole string) error
	GetAllUsers() ([]model.UserResponse, error)
	UpdateProfile(userID int64, displayName, bio string) error
}

type UserHandle struct {
	userService UserService
}

func NewUserHandle(userService UserService) *UserHandle {
	return &UserHandle{
		userService: userService,
	}
}

func (h *UserHandle) MyPage(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, apperror.ErrUnauthorized)
		return
	}

	userID := userIDVal.(int64)

	user, err := h.userService.GetProfile(userID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, user)
}

func (h *UserHandle) GetUserProfile(c *gin.Context) {
	ID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		RespondWithError(c, apperror.ErrUnauthorized)
		return
	}

	var userID int64
	userIDVal, exists := c.Get("userID")

	if exists {
		switch v := userIDVal.(type) {
		case int64:
			userID = v
		case float64:
			userID = int64(v)
		case int:
			userID = int64(v)
		}
	}

	isMe := userID != 0 && userID == ID

	user, err := h.userService.GetProfile(ID)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	var response model.UserProfileResponse

	response.IsMe = isMe
	response.Data.UserProfile = *user

	c.JSON(http.StatusOK, response)
}

func (h *UserHandle) UpdateRole(c *gin.Context) {
	targetUserID, err := strconv.ParseInt(c.Param("id"), 10, 64) // URL から id を引き出す
	if err != nil || targetUserID <= int64(0) {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}
	currentUserID := c.GetInt64("userID") // JWT から id を引き出す
	if currentUserID == 0 {
		RespondWithError(c, apperror.ErrUnauthorized)
		return
	}

	var req model.UpdateRoleRequest
	err = c.ShouldBindJSON(&req)
	if err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	err = h.userService.UpdateRole(currentUserID, targetUserID, req.Role)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "role updated successfully"})
}

func (h *UserHandle) UpdateProfile(c *gin.Context) {
	userIDVal, exists := c.Get("userID")
	if !exists {
		RespondWithError(c, apperror.ErrUnauthorized)
		return
	}
	userID := userIDVal.(int64)

	var profile model.ProfileRequest
	if err := c.ShouldBindJSON(&profile); err != nil {
		RespondWithError(c, apperror.ErrInvalidInput)
		return
	}

	err := h.userService.UpdateProfile(userID, profile.DisplayName, profile.Bio)
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "profile updated successfully"})
}

func (h *UserHandle) GetAllUsers(c *gin.Context) {
	users, err := h.userService.GetAllUsers()
	if err != nil {
		RespondWithError(c, err)
		return
	}

	c.JSON(http.StatusOK, users)
}
