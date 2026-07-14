package repository

import (
	"context"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.AuditLog, error)
}
