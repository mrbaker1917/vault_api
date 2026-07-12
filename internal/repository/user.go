package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (domain.User, error)
	GetByEmail(ctx context.Context, email string) (domain.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.User, error)
	EnableMFASecret(ctx context.Context, id uuid.UUID, secret string) error
	ConfirmMFA(ctx context.Context, id uuid.UUID) error
	DisableMFA(ctx context.Context, id uuid.UUID) error
}

var ErrNotFound = errors.New("not found")