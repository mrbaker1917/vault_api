package repository

import (
	"context"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type RecoveryCodeRepository interface {
	DeleteByUserID(ctx context.Context, userID uuid.UUID) error
	Create(ctx context.Context, userID uuid.UUID, codeHash string) (domain.RecoveryCode, error)
	ListUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]domain.RecoveryCode, error)
	MarkUsed(ctx context.Context, id uuid.UUID) error
}
