package model

type MyPageResponse struct {
	UserProfile UserProfile      `json:"userprofile"`
	Posts       []ArticleSummary `json:"posts"`
}

type UserProfile struct {
	UserID      int64  `json:"user_id"`
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

type User struct {
	ID           int64  `json:"id" db:"id"`
	Username     string `json:"username" db:"username"`
	PasswordHash string `json:"-" db:"password_hash"`
	Role         string `json:"role" db:"role"`
	TokenVersion int    `json:"-" db:"token_version"`
}

type UpdateRoleRequest struct {
	Role string `json:"role" binding:"required"`
}

type UserResponse struct {
	ID       int64  `json:"id" db:"id"`
	Username string `json:"username" db:"username"`
	Role     string `json:"role" db:"role"`
}

type ProfileRequest struct {
	DisplayName string `json:"display_name"`
	Bio         string `json:"bio"`
}

type UserProfileResponse struct {
	IsMe bool `json:"is_me"`
	Data struct {
		UserProfile UserProfile `json:"userprofile"`
	} `json:"data"`
}
