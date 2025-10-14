package repositories

import (
	"context"
	"errors"
	"fmt"
	"vinyl-vault/internal/services"

	"gorm.io/gorm"
)

type GormRegistrationKeyRepository struct {
	db *gorm.DB
}

func NewGormRegistrationKeyRepository(db *gorm.DB) services.RegistrationKeyRepository {
	return &GormRegistrationKeyRepository{
		db: db,
	}
}

func (r *GormRegistrationKeyRepository) FindByKey(ctx context.Context, key string) (*services.RegistrationKey, error) {
	var regKey services.RegistrationKey

	result := r.db.WithContext(ctx).Where("key = ?", key).First(&regKey)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("registration key not found")
		}
		return nil, fmt.Errorf("failed to find registration key: %w", result.Error)
	}
	return &regKey, nil
}

func (r *GormRegistrationKeyRepository) FindByCreator(ctx context.Context, creatorID uint64) ([]*services.RegistrationKey, error) {
	var keys []*services.RegistrationKey

	result := r.db.WithContext(ctx).Where("created_by = ?", creatorID).Order("created_at DESC").Find(&keys)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to find registration keys: %w", result.Error)
	}

	return keys, nil
}

func (r *GormRegistrationKeyRepository) Save(ctx context.Context, key *services.RegistrationKey) error {
	result := r.db.WithContext(ctx).Save(key)
	if result.Error != nil {
		return fmt.Errorf("failed to save registration key: %w", result.Error)
	}
	return nil
}

func (r *GormRegistrationKeyRepository) Delete(ctx context.Context, id uint64) error {
	result := r.db.WithContext(ctx).Delete(&services.RegistrationKey{}, id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete registration key: %w", result.Error)
	}

	if result.RowsAffected == 0 {
		return fmt.Errorf("registration key with id %d not found", id)
	}

	return nil
}
