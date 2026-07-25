package user

import (
	"context"
	"trackly-backend/pkg/filter"
	"trackly-backend/pkg/httpx"
)

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) GetUserProfile(ctx context.Context, id uint) (*UserResponse, *httpx.AppError) {
	u, err := s.repo.FindByID(ctx, id)
	if err != nil {
		return nil, &httpx.AppError{
			Code:   httpx.ErrInvalidCredentials,
			Detail: "user not found",
		}
	}

	return &UserResponse{
		ID:        u.ID,
		Name:      u.Name,
		Email:     u.Email,
		Avatar:    u.Avatar,
		Role:      u.Role,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}, nil
}

func (s *UserService) GetAllUsers(ctx context.Context, fq filter.FilteringQuery) (*filter.PaginatedResult[*UserResponse], *httpx.AppError) {
	data, total, err := s.repo.GetAllUsers(ctx, fq)
	if err != nil {
		return nil, &httpx.AppError{
			Code:   httpx.ErrInternal,
			Detail: "Failed to retrieve users",
		}
	}

	res := ToUserResponseList(data)
	result := filter.WrapPaginated(res, total, fq.Page, fq.Limit)

	return &result, nil
}
