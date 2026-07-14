package repository

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"vault_api/internal/domain"
	"vault_api/internal/repository/sqlc"
)

type auditLogPostgresRepository struct {
	q *sqlc.Queries
}

func NewAuditLogRepository(pg *Postgres) AuditLogRepository {
	return &auditLogPostgresRepository{
		q: sqlc.New(pg.Pool()),
	}
}

func (r *auditLogPostgresRepository) Create(ctx context.Context, log domain.AuditLog) (domain.AuditLog, error) {
	ip, err := netipAddrFromString(log.IPAddress)
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("parse ip address: %w", err)
	}

	var metadata []byte
	if len(log.Metadata) > 0 {
		metadata = log.Metadata
	}

	row, err := r.q.CreateAuditLog(ctx, sqlc.CreateAuditLogParams{
		UserID:       pgUUIDFromPtr(log.UserID),
		Action:       log.Action,
		ResourceType: pgTextFromString(log.ResourceType),
		ResourceID:   pgUUIDFromPtr(log.ResourceID),
		IpAddress:    ip,
		UserAgent:    pgTextFromString(log.UserAgent),
		Metadata:     metadata,
	})
	if err != nil {
		return domain.AuditLog{}, fmt.Errorf("create audit log: %w", err)
	}
	return toDomainAuditLog(row)
}

func (r *auditLogPostgresRepository) ListByUserID(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.AuditLog, error) {
	rows, err := r.q.ListAuditLogsByUserID(ctx, sqlc.ListAuditLogsByUserIDParams{
		UserID: pgUUIDToPG(userID),
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}

	logs := make([]domain.AuditLog, 0, len(rows))
	for _, row := range rows {
		entry, err := toDomainAuditLog(row)
		if err != nil {
			return nil, fmt.Errorf("list audit logs: %w", err)
		}
		logs = append(logs, entry)
	}
	return logs, nil
}

func toDomainAuditLog(row sqlc.AuditLog) (domain.AuditLog, error) {
	id, err := uuidFromPG(row.ID)
	if err != nil {
		return domain.AuditLog{}, err
	}

	var userID *uuid.UUID
	if row.UserID.Valid {
		parsed, err := uuidFromPG(row.UserID)
		if err != nil {
			return domain.AuditLog{}, err
		}
		userID = &parsed
	}

	var resourceID *uuid.UUID
	if row.ResourceID.Valid {
		parsed, err := uuidFromPG(row.ResourceID)
		if err != nil {
			return domain.AuditLog{}, err
		}
		resourceID = &parsed
	}

	var metadata json.RawMessage
	if len(row.Metadata) > 0 {
		metadata = json.RawMessage(row.Metadata)
	}

	return domain.AuditLog{
		ID:           id,
		UserID:       userID,
		Action:       row.Action,
		ResourceType: pgTextToString(row.ResourceType),
		ResourceID:   resourceID,
		IPAddress:    netipAddrToString(row.IpAddress),
		UserAgent:    pgTextToString(row.UserAgent),
		Metadata:     metadata,
		CreatedAt:    pgTimestampFromPG(row.CreatedAt),
	}, nil
}

func pgUUIDFromPtr(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgUUIDToPG(*id)
}
