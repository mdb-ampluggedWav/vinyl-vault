package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"time"
)

type RegistrationKey struct {
	ID        uint64     `json:"id" gorm:"primaryKey;autoIncrement"`
	Key       string     `json:"key" gorm:"uniqueIndex;not null"`
	CreatedBy uint64     `json:"created_by" gorm:"not null"`
	UsedBy    *uint64    `json:"used_by,omitempty" gorm:"index"`
	IsUsed    bool       `json:"is_used" gorm:"default:false;index"`
	ExpiresAt time.Time  `json:"expires_at" gorm:"index"`
	UsedAt    *time.Time `json:"used_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type RegistrationKeyRepository interface {
	FindByKey(ctx context.Context, key string) (*RegistrationKey, error)
	FindByCreator(ctx context.Context, creatorID uint64) ([]*RegistrationKey, error)
	Save(ctx context.Context, key *RegistrationKey) error
	Delete(ctx context.Context, id uint64) error
}

type RegistrationKeyService struct {
	keyRepository  RegistrationKeyRepository
	userRepository UserRepository
}

func NewRegistrationKeyService(keyRepository RegistrationKeyRepository, userRepository UserRepository) *RegistrationKeyService {
	return &RegistrationKeyService{
		keyRepository:  keyRepository,
		userRepository: userRepository,
	}
}

func (r *RegistrationKeyService) GenerateKey(ctx context.Context, creatorID uint64, expirationHours int) (*RegistrationKey, error) {

	creator, err := r.userRepository.FindByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("creator not found: %w", err)
	}

	if !creator.IsAdmin {
		return nil, fmt.Errorf("unauthorized: only admins can generate registration keys")
	}

	keyStr, err := generateSecureKey(32)
	if err != nil {
		return nil, fmt.Errorf("failed to generate key: %w", err)
	}

	expiresAt := time.Now().Add(time.Duration(expirationHours) * time.Hour)

	key := &RegistrationKey{
		Key:       keyStr,
		CreatedBy: creatorID,
		IsUsed:    false,
		ExpiresAt: expiresAt,
	}
	if err = r.keyRepository.Save(ctx, key); err != nil {
		return nil, fmt.Errorf("failed to save registration key: %w", err)
	}

	return key, nil
}

func (r *RegistrationKeyService) ValidateKey(ctx context.Context, keyStr string) (*RegistrationKey, error) {
	key, err := r.keyRepository.FindByKey(ctx, keyStr)
	if err != nil {
		return nil, fmt.Errorf("invalid registration key")
	}

	// Check if already used
	if key.IsUsed {
		return nil, fmt.Errorf("registration key has already been used")
	}

	// Check if expired
	if time.Now().After(key.ExpiresAt) {
		return nil, fmt.Errorf("registration key has expired")
	}

	return key, nil
}

func (r *RegistrationKeyService) MarkKeyAsUsed(ctx context.Context, keyStr string, userID uint64) error {
	key, err := r.keyRepository.FindByKey(ctx, keyStr)
	if err != nil {
		return fmt.Errorf("key not found: %w", err)
	}

	now := time.Now()
	key.IsUsed = true
	key.UsedBy = &userID
	key.UsedAt = &now

	if err = r.keyRepository.Save(ctx, key); err != nil {
		return fmt.Errorf("failed to mark key as used: %w", err)
	}

	return nil
}

func (r *RegistrationKeyService) GetKeysByCreator(ctx context.Context, creatorID uint64) ([]*RegistrationKey, error) {
	// Verify requester is admin
	creator, err := r.userRepository.FindByID(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if !creator.IsAdmin {
		return nil, fmt.Errorf("unauthorized: only admins can view registration keys")
	}

	keys, err := r.keyRepository.FindByCreator(ctx, creatorID)
	if err != nil {
		return nil, fmt.Errorf("failed to get keys: %w", err)
	}

	return keys, nil
}

func (r *RegistrationKeyService) DeleteKey(ctx context.Context, keyID, adminID uint64) error {
	// Verify requester is admin
	admin, err := r.userRepository.FindByID(ctx, adminID)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if !admin.IsAdmin {
		return fmt.Errorf("unauthorized: only admins can delete registration keys")
	}

	if err = r.keyRepository.Delete(ctx, keyID); err != nil {
		return fmt.Errorf("failed to delete key: %w", err)
	}

	return nil
}

func generateSecureKey(length int) (string, error) {

	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
