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
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		IsActive: pgBoolToPG(user.IsActive),
		MfaEnabled: pgBoolToPG(user.MfaEnabled),
		MfaSecret: pgTextFromPtr(user.MfaSecret),
	})
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}
	return toDomainUser(userRowFromCreate(row)), nil
}

func (r *userPostgresRepository) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	row, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)	
	}
	return toDomainUser(userRowFromGetByEmail(row)), nil
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

func nullStringFromPtr(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{}
	}
	return sql.NullString{String: *s, Valid: true}
}

func userRowFromCreate(r sqlc.CreateUserRow) userRow {
	return userRow{
		id: r.ID,
		email: r.Email,
		passwordHash: r.PasswordHash,
		createdAt: r.CreatedAt,
		updatedAt: r.UpdatedAt,
		isActive: pgBoolFromPG(r.IsActive),
		mfaEnabled: pgBoolFromPG(r.MfaEnabled),
		mfaSecret: pgTextToPtr(r.MfaSecret),
	}
}

func userRowFromGetByEmail(r sqlc.GetUserByEmailRow) userRow {
	return userRow{
		id: r.ID,
		email: r.Email,
		passwordHash: r.PasswordHash,
		createdAt: r.CreatedAt,
		updatedAt: r.UpdatedAt,
		isActive: pgBoolFromPG(r.IsActive),
		mfaEnabled: pgBoolFromPG(r.MfaEnabled),
		mfaSecret: pgTextToPtr(r.MfaSecret),
	}
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