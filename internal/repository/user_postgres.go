package repository

import (
	"context"
	"fmt"
	"time"
	"errors"

	"github.com/jackc/pgx/v5"
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
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)	
	}
	userRow, err := userRowFromGetByEmail(row)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	return toDomainUser(userRow), nil
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

func (r *userPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.User, error) {
	row, err := r.q.GetUserByID(ctx, pgUUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	userRow, err := userRowFromGetByID(row)
	if err != nil {
		return domain.User{}, fmt.Errorf("get user by id: %w", err)
	}
	return toDomainUser(userRow), nil
}

func userRowFromGetByID(r sqlc.GetUserByIDRow) (userRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return userRow{}, fmt.Errorf("get user by id: %w", err)
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

func (r *userPostgresRepository) EnableMFASecret(ctx context.Context, id uuid.UUID, secret string) error {
	err := r.q.EnableMFASecret(ctx, sqlc.EnableMFASecretParams{
		ID:        pgUUIDToPG(id),
		MfaSecret: pgTextFromString(secret),
	})
	if err != nil {
		return fmt.Errorf("enable mfa secret: %w", err)
	}
	return nil
}

func (r *userPostgresRepository) ConfirmMFA(ctx context.Context, id uuid.UUID) error {
	err := r.q.ConfirmMFA(ctx, pgUUIDToPG(id))
	if err != nil {
		return fmt.Errorf("confirm mfa: %w", err)
	}
	return nil
}

func (r *userPostgresRepository) DisableMFA(ctx context.Context, id uuid.UUID) error {
	err := r.q.DisableMFA(ctx, pgUUIDToPG(id))
	if err != nil {
		return fmt.Errorf("disable mfa: %w", err)
	}
	return nil
}

func (r *userPostgresRepository) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	rowsAffected, err := r.q.UpdateUserPassword(ctx, sqlc.UpdateUserPasswordParams{
		ID:           pgUUIDToPG(id),
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("update password: %w", err)
	}
	if rowsAffected == 0 {
		return ErrNotFound
	}
	return nil
}