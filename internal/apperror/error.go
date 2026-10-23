package apperror

import "errors"

var (
	// 400 Bad Request
	ErrInvalidInput          = errors.New("invalid input parameters")
	ErrUsernameInvalidLength = errors.New("username must be between 3 and 20 characters")
	ErrUsernameContainsSpace = errors.New("username cannot contain spaces")
	ErrPasswordInvalidLength = errors.New("password must be between 8 and 72 characters")
	ErrPasswordWeak          = errors.New("password must contain both letters and digits")
	ErrInvalidRole           = errors.New("invalid role specified")
	ErrSelfRoleChange        = errors.New("cannot change your own role")
	ErrInvalidTitle          = errors.New("invalid title")

	// 401 Unauthorized
	ErrUnauthorized        = errors.New("unauthorized")
	ErrInvalidCredentials  = errors.New("invalid username or password")
	ErrInvalidToken        = errors.New("invalid token")
	ErrMissingToken        = errors.New("authorization token is missing")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrRefreshTokenExpired = errors.New("refresh token has expired")
	ErrRefreshTokenRevoked = errors.New("refresh token has been revoked")

	// 403 Forbidden
	ErrForbidden = errors.New("permission denied")

	// 404 Not Found
	ErrUserNotFound = errors.New("user not found")
	ErrPostNotFound = errors.New("post not found")

	// 409 Conflict
	ErrAlreadyExists = errors.New("resource already exists")

	// 429 Too Many Requests
	ErrLoginRateLimit = errors.New("too many login attempts")

	// 500 Internal Server Error
	ErrDatabase = errors.New("database error")
)
