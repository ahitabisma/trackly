package repository

import (
	"context"
	"trackly-backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	q := `SELECT id, email, password, name, created_at FROM users WHERE email = $1`

	var u model.User

	err := r.db.QueryRow(ctx, q, email).Scan(
		&u.ID, &u.Name, &u.Email, &u.Avatar, &u.Role, &u.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return &u, nil
}
