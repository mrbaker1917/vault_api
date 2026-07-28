package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type AuditLogRepository interface {
	Create(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.AuditLog, error)
	ListFiltered(ctx context.Context, userID uuid.UUID, filter ListAuditLogsFilter) ([]domain.AuditLog, error)
}

type ListAuditLogsFilter struct {
	ActionPrefix string
	Action       string
	Since        *time.Time
	Until        *time.Time
	Limit        int32
	Offset       int32
}
