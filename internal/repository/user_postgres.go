package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/repository/sqlc"
)

type userPostgresRepository struct {
	q *sqlc.Queries
}

type userRow struct {
	id           uuid.UUID
	email        string
	passwordHash string
	createdAt    time.Time
	updatedAt    time.Time
	isActive     bool
	mfaEnabled   bool
	mfaSecret    *string
}

func NewUserRepository(pg *Postgres) UserRepository {
	return &userPostgresRepository{
		q: sqlc.New(pg.Pool()),
	}
}

func (r *userPostgresRepository) Create(ctx context.Context, user domain.User) (domain.User, error) {
	row, err := r.q.CreateUser(ctx, sqlc.CreateUserParams{
		Email: user.Email,
		PasswordHash: user.PasswordHash,
		CreatedAt: pgTimestampToPG(user.CreatedAt),
		UpdatedAt: pgTimestampToPG(user.UpdatedAt),
		IsActive: pgBoolToPG(user.IsActive),
		MfaEnabled: pgBoolToPG(user.MfaEnabled),
		MfaSecret: pgTextFromPtr(user.MfaSecret),
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	userRow, err := userRowFromCreate(row)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(userRow), nil
}

func (r *userPostgresRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)	
	}
	userRow, err := userRowFromGetByEmail(row)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toDomainUser(userRow), nil
}

func pgBoolToPG(v bool) pgtype.Bool {
	return pgtype.Bool{Bool: v, Valid: true}
}

func pgBoolFromPG(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

func pgTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgTimestampToPG(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

func pgTimestampFromPG(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func uuidFromPG(u pgtype.UUID) (uuid.UUID, error) {
	if !u.Valid {
		return uuid.Nil, fmt.Errorf("null uuid")
	}
	return uuid.FromBytes(u.Bytes[:])
}

func userRowFromCreate(r sqlc.CreateUserRow) (userRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return userRow{}, err
	}
	return userRow{
		id: id,
		email: r.Email,
		passwordHash: r.PasswordHash,
		createdAt: pgTimestampFromPG(r.CreatedAt),
		updatedAt: pgTimestampFromPG(r.UpdatedAt),
		isActive: pgBoolFromPG(r.IsActive),
		mfaEnabled: pgBoolFromPG(r.MfaEnabled),
		mfaSecret: pgTextToPtr(r.MfaSecret),
	}, nil
}

func userRowFromGetByEmail(r sqlc.GetUserByEmailRow) (userRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return userRow{}, fmt.Errorf("get user by email: %w", err)
	}
	userRow := userRow{
		id: id,
		email: r.Email,
		passwordHash: r.PasswordHash,
		createdAt: pgTimestampFromPG(r.CreatedAt),
		updatedAt: pgTimestampFromPG(r.UpdatedAt),
		isActive: pgBoolFromPG(r.IsActive),
		mfaEnabled: pgBoolFromPG(r.MfaEnabled),
		mfaSecret: pgTextToPtr(r.MfaSecret),
	}
	return userRow, nil
}

func toDomainUser(row userRow) domain.User {
	return domain.User{
		ID: row.id,
		Email: row.email,
		PasswordHash: row.passwordHash,
		CreatedAt: row.createdAt,
		UpdatedAt: row.updatedAt,
		IsActive: row.isActive,
		MfaEnabled: row.mfaEnabled,
		MfaSecret: row.mfaSecret,
	}
}