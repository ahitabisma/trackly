package service

import (
	"context"
	"errors"
	"time"
	"trackly-backend/internal/dto"
	"trackly-backend/internal/repository"
	"trackly-backend/pkg/httpx"

	"github.com/golang-jwt/jwt/v5"
	"github.com/sirupsen/logrus"
	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetRole(ctx context.Context, userID string) (string, error)
	Login(ctx context.Context, email string, password string) (*dto.LoginResponse, error)
}

type userService struct {
	repo   repository.UserRepository
	log    *logrus.Logger
	config string
}

func NewUserService(repo repository.UserRepository, log *logrus.Logger, secretKey string) UserService {
	return &userService{
		repo:   repo,
		log:    log,
		config: secretKey,
	}
}

func (s *userService) GetRole(ctx context.Context, userID string) (string, error) {
	role, err := s.repo.GetRole(ctx, userID)
	if err != nil {
		s.log.WithError(err).Error("failed to get role")
		return "", errors.New("role not found")
	}

	return role, nil
}

func (s *userService) Login(ctx context.Context, email string, password string) (*dto.LoginResponse, error) {
	// Find User
	user, err := s.repo.FindByEmail(ctx, email)

	if err != nil {
		return nil, httpx.NewCustomError(401, "INVALID_CREDENTIALS", "Email atau password salah")
	}

	// Check Password
	if user.Password == nil {
		return nil, httpx.NewCustomError(401, "INVALID_CREDENTIALS", "Email atau password salah")
	}

	err = bcrypt.CompareHashAndPassword([]byte(*user.Password), []byte(password))
	if err != nil {
		return nil, httpx.NewCustomError(401, "INVALID_CREDENTIALS", "Email atau password salah")
	}

	// 3. Generate JWT
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"exp":     time.Now().Add(time.Hour * 24).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.config))
	if err != nil {
		return nil, httpx.NewCustomError(500, "INTERNAL_ERROR", "Gagal generate token")
	}

	return &dto.LoginResponse{
		AccessToken: tokenString,
		User:        user,
		TokenType:   "Bearer",
	}, nil
}
