package repository

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/repository/sqlc"
)

type recoveryCodePostgresRepository struct {
	q *sqlc.Queries
}

func NewRecoveryCodeRepository(pg *Postgres) RecoveryCodeRepository {
	return &recoveryCodePostgresRepository{
		q: sqlc.New(pg.Pool()),
	}
}

func (r *recoveryCodePostgresRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	if err := r.q.DeleteRecoveryCodesByUserID(ctx, pgUUIDToPG(userID)); err != nil {
		return fmt.Errorf("delete recovery codes: %w", err)
	}
	return nil
}

func (r *recoveryCodePostgresRepository) Create(ctx context.Context, userID uuid.UUID, codeHash string) (domain.RecoveryCode, error) {
	row, err := r.q.CreateRecoveryCode(ctx, sqlc.CreateRecoveryCodeParams{
		UserID:   pgUUIDToPG(userID),
		CodeHash: codeHash,
	})
	if err != nil {
		return domain.RecoveryCode{}, fmt.Errorf("create recovery code: %w", err)
	}
	return toDomainRecoveryCode(row)
}

func (r *recoveryCodePostgresRepository) ListUnusedByUserID(ctx context.Context, userID uuid.UUID) ([]domain.RecoveryCode, error) {
	rows, err := r.q.ListUnusedRecoveryCodesByUserID(ctx, pgUUIDToPG(userID))
	if err != nil {
		return nil, fmt.Errorf("list unused recovery codes: %w", err)
	}

	codes := make([]domain.RecoveryCode, 0, len(rows))
	for _, row := range rows {
		code, err := toDomainRecoveryCode(row)
		if err != nil {
			return nil, fmt.Errorf("list unused recovery codes: %w", err)
		}
		codes = append(codes, code)
	}
	return codes, nil
}

func (r *recoveryCodePostgresRepository) MarkUsed(ctx context.Context, id uuid.UUID) error {
	rowsAffected, err := r.q.MarkRecoveryCodeUsed(ctx, pgUUIDToPG(id))
	if err != nil {
		return fmt.Errorf("mark recovery code used: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}

func toDomainRecoveryCode(row sqlc.RecoveryCode) (domain.RecoveryCode, error) {
	id, err := uuidFromPG(row.ID)
	if err != nil {
		return domain.RecoveryCode{}, err
	}
	userID, err := uuidFromPG(row.UserID)
	if err != nil {
		return domain.RecoveryCode{}, err
	}
	return domain.RecoveryCode{
		ID:        id,
		UserID:    userID,
		CodeHash:  row.CodeHash,
		UsedAt:    pgTimestampToPtr(row.UsedAt),
		CreatedAt: pgTimestampFromPG(row.CreatedAt),
	}, nil
}
