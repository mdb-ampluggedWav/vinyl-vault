package repositories

import (
	"context"
	"errors"
	"fmt"
	"vinyl-vault/internal/services"

	"gorm.io/gorm"
)

type GormUserRepository struct {
	db *gorm.DB
}

func NewGormUserRepository(db *gorm.DB) services.UserRepository {
	return &GormUserRepository{
		db: db,
	}
}

func (r *GormUserRepository) FindByID(ctx context.Context, id uint64) (*services.User, error) {

	var user services.User

	result := r.db.WithContext(ctx).First(&user, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user with id %d not found", id)
		}
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

func (r *GormUserRepository) FindByUsername(ctx context.Context, username string) (*services.User, error) {

	var user services.User
	if result := r.db.WithContext(ctx).Where("username = ?", username).First(&user); result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

func (r *GormUserRepository) FindByEmail(ctx context.Context, email string) (*services.User, error) {

	var user services.User
	if result := r.db.WithContext(ctx).Where("email = ?", email).First(&user); result.Error != nil {
		return nil, fmt.Errorf("failed to find user: %w", result.Error)
	}
	return &user, nil
}

func (r *GormUserRepository) Save(ctx context.Context, user *services.User) error {
	result := r.db.WithContext(ctx).Save(user)
	if result.Error != nil {
		return fmt.Errorf("failed to save user: %w", result.Error)
	}
	return nil
}

func (r *GormUserRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&services.User{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return fmt.Errorf("user with id %d not found", id)
	}
	return nil
}
