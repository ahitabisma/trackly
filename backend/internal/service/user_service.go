package service

import (
	"context"
	"errors"
	"trackly-backend/internal/repository"

	"github.com/sirupsen/logrus"
)

type UserService interface {
	EnsureUser(ctx context.Context, userID string, email string) error
	GetRole(ctx context.Context, userID string) (string, error)
}

type userService struct {
	repo repository.UserRepository
	log  *logrus.Logger
}

func NewUserService(repo repository.UserRepository, log *logrus.Logger) UserService {
	return &userService{
		repo: repo,
		log:  log,
	}
}

func (s *userService) EnsureUser(ctx context.Context, userID string, email string) error {

	_, _, err := s.repo.FindByID(ctx, userID)
	if err == nil {
		return nil // user already exists
	}

	// user tidak ditemukan, buat baru
	s.log.Warn("user not found, creating new user")

	err = s.repo.Create(ctx, userID, email)
	if err != nil {
		s.log.WithError(err).Error("failed to create user")
		return err
	}

	return nil
}

func (s *userService) GetRole(ctx context.Context, userID string) (string, error) {
	role, err := s.repo.GetRole(ctx, userID)
	if err != nil {
		s.log.WithError(err).Error("failed to get role")
		return "", errors.New("role not found")
	}

	return role, nil
}
