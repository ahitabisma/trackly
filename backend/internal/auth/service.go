package auth

import (
	"trackly-backend/internal/user"
	"trackly-backend/pkg/httpx"
)

type AuthService struct {
	userRepo user.UserRepository
	jwtSvc   *JwtService
}

func NewAuthService(userRepo user.UserRepository, jwtSvc *JwtService) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSvc:   jwtSvc,
	}
}

func (s *AuthService) Register(req *RegisterRequest) *httpx.AppError {
	if req.Email == "" {
		return &httpx.AppError{
			Code: httpx.ErrValidation,
			Fields: map[string]string{
				"email": "email is required",
			},
		}
	}

	existingUser, _ := s.userRepo.FindByEmail(req.Email)
	if existingUser != nil {
		return &httpx.AppError{
			Code: httpx.ErrValidation,
			Fields: map[string]string{
				"email": "email already in use",
			},
		}
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "failed to hash password",
		}
	}

	user := &user.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: hash,
		Role:     "user",
	}

	if err := s.userRepo.Create(user); err != nil {
		return &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: err.Error(),
		}
	}

	return nil
}

func (s *AuthService) Login(req *LoginRequest) (*LoginResponse, *httpx.AppError) {
	u, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, &httpx.AppError{
			Code: httpx.ErrInvalidCredentials,
		}
	}

	if !CheckPassword(req.Password, u.Password) {
		return nil, &httpx.AppError{
			Code: httpx.ErrInvalidCredentials,
		}
	}

	token, err := s.jwtSvc.GenerateToken(u.ID, u.Email, u.Role)
	if err != nil {
		return nil, &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "failed to generate token",
		}
	}

	return &LoginResponse{
		AccessToken: token,
		User: &user.UserResponse{
			ID:        u.ID,
			Name:      u.Name,
			Email:     u.Email,
			Avatar:    u.Avatar,
			Role:      u.Role,
			CreatedAt: u.CreatedAt,
			UpdatedAt: u.UpdatedAt,
		},
	}, nil
}
