package dto

import "time"

type UserResponse struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Avatar    *string   `json:"avatar,omitempty"`
	Role      string    `json:"role"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type CreateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Avatar   *string `json:"avatar"`
	Role     string  `json:"role"`
}

type UpdateUserRequest struct {
	Name     string  `json:"name" binding:"required"`
	Email    string  `json:"email" binding:"required"`
	Password string  `json:"password" binding:"required"`
	Avatar   *string `json:"avatar"`
	Role     string  `json:"role"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	AccessToken string `json:"access_token"`
	User        *UserResponse
}
