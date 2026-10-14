package service

import (
	"github.com/return-err/myblog/internal/apperror"
	"github.com/return-err/myblog/internal/model"
)

type UserService struct {
	repo UserRepository
}

type UserRepository interface {
	GetUserProfile(userID int64) (*model.UserProfile, error)
	UpdateRole(userID int64, newRole string) error
	GetAllUsers() ([]model.UserResponse, error)
	UpdateUserProfile(userID int64, displayName, bio string) error
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetProfile(userID int64) (*model.UserProfile, error) {
	return s.repo.GetUserProfile(userID)
}

func (s *UserService) UpdateRole(operatorID, userID int64, newRole string) error {
	if newRole != "admin" && newRole != "user" {
		return apperror.ErrInvalidRole
	}

	if operatorID == userID {
		return apperror.ErrSelfRoleChange
	}

	return s.repo.UpdateRole(userID, newRole)
}

func (s *UserService) GetAllUsers() ([]model.UserResponse, error) {
	return s.repo.GetAllUsers()
}

func (s *UserService) UpdateProfile(userID int64, displayName, bio string) error {
	return s.repo.UpdateUserProfile(userID, displayName, bio)
}
