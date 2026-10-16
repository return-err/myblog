package apperror

import "errors"

var (
	ErrAlreadyExists = errors.New("resource already exists")
	ErrUserNotFound  = errors.New("user not found")

	ErrUsernameInvalidLength = errors.New("username must be between 3 and 20 characters")
	ErrUsernameContainsSpace = errors.New("username cannot contain spaces")

	ErrPasswordInvalidLength = errors.New("password must be between 8 and 72 characters")
	ErrPasswordWeak          = errors.New("password must contain both letters and digits")

	ErrInvalidCredentials = errors.New("invalid username or password")
	ErrDatabase           = errors.New("database error")

	ErrInvalidRole    = errors.New("invalid role specified")
	ErrSelfRoleChange = errors.New("cannot change your own role")
	ErrForbidden      = errors.New("permission denied")

	ErrInvalidTitle = errors.New("invalid title")
	ErrPostNotFound = errors.New("post not found")

	ErrInvalidInput = errors.New("invalid input parameters")
	ErrUnauthorized = errors.New("unauthorized")

	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")

	ErrLoginRateLimit = errors.New("too many login attempts")
)
