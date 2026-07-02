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

type sessionPostgresRepository struct {
	q *sqlc.Queries
}

type sessionRow struct {
	id uuid.UUID
	userID uuid.UUID
	tokenHash string
	createdAt time.Time
	expiresAt time.Time
	revokedAt *time.Time
	deviceName string
	ipAddress string
	userAgent string
}

func NewSessionRepository(pg *Postgres) SessionRepository {
    return &sessionPostgresRepository{
        q: sqlc.New(pg.Pool()),
    }
}

func (r *sessionPostgresRepository) Create(ctx context.Context, session domain.Session) (domain.Session, error) {
	ip, err := netipAddrFromString(session.IPAddress)
	if err != nil {
    	return domain.Session{}, fmt.Errorf("parse ip address: %w", err)
	}

	row, err := r.q.CreateSession(ctx, sqlc.CreateSessionParams{
		UserID: pgUUIDToPG(session.UserID),
		TokenHash: session.TokenHash,
		CreatedAt: pgTimestampToPG(session.CreatedAt),
		ExpiresAt: pgTimestampToPG(session.ExpiresAt),
		DeviceName: pgTextFromString(session.DeviceName),
		IpAddress: ip,
		UserAgent: pgTextFromString(session.UserAgent),
	})
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}
	sessionRow, err := sessionRowFromCreate(row)
	if err != nil {
		return domain.Session{}, fmt.Errorf("create session: %w", err)
	}
	return toDomainSession(sessionRow), nil
}

func (r *sessionPostgresRepository) GetByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	row, err := r.q.GetSessionByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.Session{}, ErrNotFound
		}
		return domain.Session{}, fmt.Errorf("get session by token hash: %w", err)
	}
	sessionRow, err := sessionRowFromGetByTokenHash(row)
	if err != nil {
		return domain.Session{}, fmt.Errorf("get session by token hash: %w", err)
	}
	return toDomainSession(sessionRow), nil
}

func (r *sessionPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	row, err := r.q.GetSessionByID(ctx, pgUUIDToPG(id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
				return domain.Session{}, ErrNotFound
			}
		return domain.Session{}, fmt.Errorf("get session by id: %w", err)
	}
	sessionRow, err := sessionRowFromGetByID(row)
	if err != nil {
		return domain.Session{}, fmt.Errorf("get session by id: %w", err)
	}
	return toDomainSession(sessionRow), nil
}

func sessionRowFromGetByID(r sqlc.Session) (sessionRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return sessionRow{}, err
	}
	userID, err := uuidFromPG(r.UserID)
	if err != nil {
		return sessionRow{}, err
	}
	return sessionRow{
		id: id,
		userID: userID,
		tokenHash: r.TokenHash,
		createdAt: pgTimestampFromPG(r.CreatedAt),
		expiresAt: pgTimestampFromPG(r.ExpiresAt),
		revokedAt: pgTimestampToPtr(r.RevokedAt),
		deviceName: pgTextToString(r.DeviceName),
		ipAddress: netipAddrToString(r.IpAddress),
		userAgent: pgTextToString(r.UserAgent),
	}, nil
}

func sessionRowFromCreate(r sqlc.Session) (sessionRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return sessionRow{}, err
	}
	userID, err := uuidFromPG(r.UserID)
	if err != nil {
		return sessionRow{}, err
	}
	return sessionRow{
		id: id,
		userID: userID,
		tokenHash: r.TokenHash,
		createdAt: pgTimestampFromPG(r.CreatedAt),
		expiresAt: pgTimestampFromPG(r.ExpiresAt),
		revokedAt: pgTimestampToPtr(r.RevokedAt),
		deviceName: pgTextToString(r.DeviceName),
		ipAddress: netipAddrToString(r.IpAddress),
		userAgent: pgTextToString(r.UserAgent),
	}, nil
}

func sessionRowFromGetByTokenHash(r sqlc.Session) (sessionRow, error) {
	id, err := uuidFromPG(r.ID)
	if err != nil {
		return sessionRow{}, err
	}
	userID, err := uuidFromPG(r.UserID)
	if err != nil {
		return sessionRow{}, err
	}
	return sessionRow{
		id: id,
		userID: userID,
		tokenHash: r.TokenHash,
		createdAt: pgTimestampFromPG(r.CreatedAt),
		expiresAt: pgTimestampFromPG(r.ExpiresAt),
		revokedAt: pgTimestampToPtr(r.RevokedAt),
		deviceName: pgTextToString(r.DeviceName),
		ipAddress: netipAddrToString(r.IpAddress),
		userAgent: pgTextToString(r.UserAgent),
	}, nil
}

func toDomainSession(row sessionRow) domain.Session {
	return domain.Session{
		ID: row.id,
		UserID: row.userID,
		TokenHash: row.tokenHash,
		CreatedAt: row.createdAt,
		ExpiresAt: row.expiresAt,
		RevokedAt: row.revokedAt,
		DeviceName: row.deviceName,
		IPAddress: row.ipAddress,
		UserAgent: row.userAgent,
	}
}

func (r *sessionPostgresRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	err := r.q.RevokeSession(ctx, pgUUIDToPG(id))
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}