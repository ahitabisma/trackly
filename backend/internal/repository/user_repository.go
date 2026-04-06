package repository

import (
	"context"
	"trackly-backend/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository interface {
	FindByID(ctx context.Context, id string) (*model.User, error)
	Create(ctx context.Context, id string, email string) error
	GetRole(ctx context.Context, id string) (string, error)
	FindByEmail(ctx context.Context, email string) (*model.User, error)
}

type userRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(db *pgxpool.Pool) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByID(ctx context.Context, id string) (*model.User, error) {
	query := `SELECT id, email, password, name, created_at FROM users WHERE id = $1`

	var user model.User

	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt,
	)

	if err != nil {
		return nil, err
	}
	
	return &user, nil
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

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*model.User, error) {
	query := `SELECT id, email, password, name, created_at FROM users WHERE email = $1`

	var user model.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Email, &user.Password, &user.Name, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}
