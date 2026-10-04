package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) (domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error)
	GetByPreviousTokenHash(ctx context.Context, tokenHash string) (domain.Session, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error)
	RotateToken(ctx context.Context, sessionID uuid.UUID, currentTokenHash, newTokenHash string) (bool, error)
	RotateFromPreviousToken(ctx context.Context, sessionID uuid.UUID, previousTokenHash, newTokenHash string, grace time.Duration) (bool, error)
	Revoke(ctx context.Context, id uuid.UUID) error
	ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error)
	RevokeByID(ctx context.Context, id, userID uuid.UUID) error
	RevokeAllExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) error
}
