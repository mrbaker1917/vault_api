package repository

import (
	"context"
	"vault_api/internal/domain"

	"github.com/google/uuid"
)

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) (domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	RevokeByID(ctx context.Context, id, userID uuid.UUID) error
}
