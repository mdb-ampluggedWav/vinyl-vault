package services

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const passwordMinLength = 8

type User struct {
	ID           uint64    `json:"id" gorm:"primaryKey;autoIncrement"`
	Username     string    `json:"username" gorm:"uniqueIndex;not null"`
	Email        string    `json:"email" gorm:"uniqueIndex;not null"`
	PasswordHash string    `json:"-" gorm:"not null"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type UserRepository interface {
	FindByID(ctx context.Context, id uint64) (*User, error)
	FindByUsername(ctx context.Context, username string) (*User, error)
	FindByEmail(ctx context.Context, email string) (*User, error)
	Save(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uint64) error
}

type UserService struct {
	userRepository UserRepository
}

func NewUserService(userRepository UserRepository) *UserService {
	return &UserService{
		userRepository: userRepository,
	}
}

func (u *UserService) Register(ctx context.Context, username, email, password string) (*User, error) {

	if len(password) < passwordMinLength {
		return nil, fmt.Errorf("password needs at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	if err = u.userRepository.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	return user, nil
}

func (u *UserService) Login(ctx context.Context, username, password string) (*User, error) {
	user, err := u.userRepository.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	return user, nil
}

func (u *UserService) GetUser(ctx context.Context, id uint64) (*User, error) {
	user, err := u.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (u *UserService) UpdateUsername(ctx context.Context, id uint64, username string) (*User, error) {
	user, err := u.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	user.Username = username

	if err = u.userRepository.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update username: %w", err)
	}
	return user, nil
}

func (u *UserService) UpdateEmail(ctx context.Context, id uint64, email string) (*User, error) {
	user, err := u.userRepository.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	user.Email = email

	if err = u.userRepository.Save(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update email: %w", err)
	}
	return user, nil
}

func (u *UserService) ChangePassword(ctx context.Context, id uint64, oldPassword, newPassword string) error {
	user, err := u.userRepository.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("user not found: %w", err)
	}

	if err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return fmt.Errorf("incorrect current password")
	}

	if len(newPassword) < passwordMinLength {
		return fmt.Errorf("password needs at least 8 characters")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	user.PasswordHash = string(hashedPassword)

	if err = u.userRepository.Save(ctx, user); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}
	return nil
}

func (u *UserService) DeleteUser(ctx context.Context, id uint64) error {

	if _, err := u.userRepository.FindByID(ctx, id); err != nil {
		return fmt.Errorf("user not found: %w", err)
	}
	if err := u.userRepository.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}
	return nil
}
