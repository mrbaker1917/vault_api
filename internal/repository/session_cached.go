package repository

import (
	"context"
	"log/slog"

	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/sessioncache"
)

type cachedSessionRepository struct {
	base  SessionRepository
	cache *sessioncache.Cache
}

func NewCachedSessionRepository(base SessionRepository, cache *sessioncache.Cache) SessionRepository {
	return &cachedSessionRepository{
		base:  base,
		cache: cache,
	}
}

func (r *cachedSessionRepository) Create(ctx context.Context, session domain.Session) (domain.Session, error) {
	return r.base.Create(ctx, session)
}

func (r *cachedSessionRepository) GetByTokenHash(ctx context.Context, tokenHash string) (domain.Session, error) {
	return r.base.GetByTokenHash(ctx, tokenHash)
}

func (r *cachedSessionRepository) GetByID(ctx context.Context, id uuid.UUID) (domain.Session, error) {
	userID, ok, err := r.cache.Get(ctx, id)
	if err != nil {
		slog.Warn("session cache get failed; falling back to database", "session_id", id, "error", err)
	} else if ok {
		return domain.Session{ID: id, UserID: userID}, nil
	}

	session, err := r.base.GetByID(ctx, id)
	if err != nil {
		return domain.Session{}, err
	}

	if err := r.cache.Set(ctx, session.ID, session.UserID, session.ExpiresAt); err != nil {
		slog.Warn("session cache set failed", "session_id", session.ID, "error", err)
	}

	return session, nil
}

func (r *cachedSessionRepository) Revoke(ctx context.Context, id uuid.UUID) error {
	if err := r.base.Revoke(ctx, id); err != nil {
		return err
	}
	r.invalidate(ctx, id)
	return nil
}

func (r *cachedSessionRepository) ListByUserID(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	return r.base.ListByUserID(ctx, userID)
}

func (r *cachedSessionRepository) RevokeByID(ctx context.Context, id, userID uuid.UUID) error {
	if err := r.base.RevokeByID(ctx, id, userID); err != nil {
		return err
	}
	r.invalidate(ctx, id)
	return nil
}

func (r *cachedSessionRepository) RevokeAllExcept(ctx context.Context, userID, exceptSessionID uuid.UUID) error {
	sessions, err := r.base.ListByUserID(ctx, userID)
	if err != nil {
		return err
	}

	if err := r.base.RevokeAllExcept(ctx, userID, exceptSessionID); err != nil {
		return err
	}

	for _, session := range sessions {
		if session.ID != exceptSessionID {
			r.invalidate(ctx, session.ID)
		}
	}
	return nil
}

func (r *cachedSessionRepository) invalidate(ctx context.Context, sessionID uuid.UUID) {
	if err := r.cache.Delete(ctx, sessionID); err != nil {
		slog.Warn("session cache delete failed", "session_id", sessionID, "error", err)
	}
}
