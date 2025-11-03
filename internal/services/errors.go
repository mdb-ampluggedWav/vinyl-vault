package services

import (
	"errors"
	"fmt"
)

var (
	ErrUnauthorized       = errors.New("unauthorized")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrNotOwner           = errors.New("you don't own this resource")

	ErrInvalidInput     = errors.New("invalid input")
	ErrInvalidEmail     = errors.New("invalid email")
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrRequiredField    = errors.New("required field missing")

	ErrNotFound      = errors.New("resource not found")
	ErrUserNotFound  = errors.New("user not found")
	ErrAlbumNotFound = errors.New("album not found")
	ErrTrackNotFound = errors.New("track not found")
	ErrKeyNotFound   = errors.New("registration key not found")

	ErrFileNotFound      = errors.New("file not found")
	ErrFileTooLarge      = errors.New("file too large")
	ErrUnsupportedFormat = errors.New("unsupported file format")
	ErrFileOutsideDir    = errors.New("file path is outside allowed directories")

	ErrKeyAlreadyUsed = errors.New("registration key has already been used")
	ErrKeyExpired     = errors.New("registration key has expired")
	ErrInvalidKey     = errors.New("invalid registration key")

	ErrAdminOnly = errors.New("only admins can perform this action")
)

type ServiceError struct {
	Op  string // Operation that failed. ex: ("UserService.Register")
	Err error  // The actual error
	Msg string // optional User-friendly message
}

func (e *ServiceError) Error() string {
	if e.Msg != "" {
		return fmt.Sprintf("%s: %s", e.Op, e.Msg)
	}
	return fmt.Sprintf("%s: %v", e.Op, e.Err)
}

func (e *ServiceError) Unwrap() error {
	return e.Err
}

func NewServiceError(op string, err error, msg string) *ServiceError {
	return &ServiceError{
		Op:  op,
		Err: err,
		Msg: msg,
	}
}

type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

func NewValidationError(field, message string) *ValidationError {
	return &ValidationError{
		Field:   field,
		Message: message,
	}
}

func IsNotFound(err error) bool {
	return errors.Is(err, ErrNotFound) ||
		errors.Is(err, ErrUserNotFound) ||
		errors.Is(err, ErrAlbumNotFound) ||
		errors.Is(err, ErrTrackNotFound) ||
		errors.Is(err, ErrKeyNotFound) ||
		errors.Is(err, ErrFileNotFound)
}

func IsUnauthorized(err error) bool {
	return errors.Is(err, ErrUnauthorized) ||
		errors.Is(err, ErrInvalidCredentials) ||
		errors.Is(err, ErrNotOwner) ||
		errors.Is(err, ErrAdminOnly)
}

func IsValidation(err error) bool {
	var valErr *ValidationError
	return errors.As(err, &valErr) ||
		errors.Is(err, ErrInvalidInput) ||
		errors.Is(err, ErrInvalidEmail) ||
		errors.Is(err, ErrPasswordTooShort) ||
		errors.Is(err, ErrRequiredField)
}
