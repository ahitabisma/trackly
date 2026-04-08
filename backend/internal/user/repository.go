package user

import (
	"context"
	"trackly-backend/pkg/filter"

	"gorm.io/gorm"
)

type UserRepository interface {
	FindByEmail(ctx context.Context, email string) (*User, error)
	Create(ctx context.Context, user *User) (*User, error)
	FindByID(ctx context.Context, id uint) (*User, error)
	GetAllUsers(ctx context.Context, fq filter.FilteringQuery) ([]*User, int64, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Create(ctx context.Context, user *User) (*User, error) {
	if err := r.db.WithContext(ctx).Create(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*User, error) {
	var user User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetAllUsers(ctx context.Context, fq filter.FilteringQuery) ([]*User, int64, error) {
	allowed := []string{"name", "email", "role"}

	db := r.db.WithContext(ctx).Model(&User{})

	// apply filters
	db = filter.ApplyGormFilter(db, fq, allowed)

	// count
	var total int64
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// order
	order := "name ASC"
	if fq.OrderKey != "" {
		rule := "DESC"
		if fq.OrderRule == "asc" {
			rule = "ASC"
		}
		order = fq.OrderKey + " " + rule
	}

	db = db.Order(order)

	// pagination
	db, _, _ = filter.ApplyPagination(db, fq)

	var result []*User
	if err := db.Find(&result).Error; err != nil {
		return nil, 0, err
	}

	return result, total, nil
}
