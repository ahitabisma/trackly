package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (string, string, error)
	Create(ctx context.Context, id string, email string) error
	GetRole(ctx context.Context, id string) (string, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (string, string, error) {
	var userID string
	var email string

	err := r.db.QueryRow(ctx,
		`SELECT id, email FROM users WHERE id = $1`,
		id,
	).Scan(&userID, &email)

	return userID, email, err
}

func (r *userRepository) Create(ctx context.Context, id string, email string) error {
	_, err := r.db.Exec(ctx,
		`INSERT INTO users (id, email, role) VALUES ($1, $2, 'user')`,
		id, email,
	)
	return err
}

func (r *userRepository) GetRole(ctx context.Context, id string) (string, error) {
	var role string

	err := r.db.QueryRow(ctx,
		`SELECT role FROM users WHERE id = $1`,
		id,
	).Scan(&role)

	return role, err
}
