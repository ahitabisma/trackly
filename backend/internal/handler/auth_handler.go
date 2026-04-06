package handler

import (
	"encoding/json"
	"net/http"

	"trackly-backend/internal/dto"
	"trackly-backend/internal/service"
	"trackly-backend/pkg/httpx"
)

type AuthHandler struct {
	service service.UserService
}

func NewAuthHandler(s service.UserService) *AuthHandler {
	return &AuthHandler{service: s}
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) error {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		return httpx.NewBadRequestError("Invalid request body")
	}

	// Basic validation
	var fieldErrors []httpx.FieldError
	if req.Email == "" {
		fieldErrors = append(fieldErrors, httpx.NewFieldError("email", "Email is required"))
	}
	if req.Password == "" {
		fieldErrors = append(fieldErrors, httpx.NewFieldError("password", "Password is required"))
	}

	if len(fieldErrors) > 0 {
		return httpx.NewValidationError("Validation failed", fieldErrors)
	}

	resp, err := h.service.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		return err
	}

	httpx.Success(w, resp, "Login successful")
	return nil
}
