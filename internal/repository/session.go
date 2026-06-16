package repository

import (
	"context"
	"vault_api/internal/domain"
)

type SessionRepository interface {
	Create(ctx context.Context, session domain.Session) (domain.Session, error)
	GetByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error)
}