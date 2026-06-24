package repository

import (
	"context"
	"errors"
	"vault_api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
}

var ErrNotFound = errors.New("not found")