package dto

import "trackly-backend/internal/model"

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type LoginResponse struct {
	User        *model.User `json:"user"`
	AccessToken string      `json:"access_token"`
	TokenType   string      `json:"token_type"`
}
