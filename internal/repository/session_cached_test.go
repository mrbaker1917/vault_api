package repository

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"vault_api/internal/domain"
	"vault_api/internal/sessioncache"
)

type countingSessionRepo struct {
	getByIDCalls int
	sessions   map[uuid.UUID]domain.Session
}

func (s *countingSessionRepo) Create(_ context.Context, session domain.Session) (domain.Session, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	s.sessions[session.ID] = session
	return session, nil
}

func (s *countingSessionRepo) GetByTokenHash(_ context.Context, _ string) (domain.Session, error) {
	return domain.Session{}, ErrNotFound
}

func (s *countingSessionRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Session, error) {
	s.getByIDCalls++
	session, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, ErrNotFound
	}
	return session, nil
}

func (s *countingSessionRepo) Revoke(_ context.Context, id uuid.UUID) error {
	delete(s.sessions, id)
	return nil
}

func (s *countingSessionRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]domain.Session, error) {
	items := make([]domain.Session, 0)
	for _, session := range s.sessions {
		if session.UserID == userID {
			items = append(items, session)
		}
	}
	return items, nil
}

func (s *countingSessionRepo) RevokeByID(_ context.Context, id, userID uuid.UUID) error {
	session, ok := s.sessions[id]
	if !ok || session.UserID != userID {
		return ErrNotFound
	}
	delete(s.sessions, id)
	return nil
}

func (s *countingSessionRepo) RevokeAllExcept(_ context.Context, userID, exceptSessionID uuid.UUID) error {
	for id, session := range s.sessions {
		if session.UserID == userID && id != exceptSessionID {
			delete(s.sessions, id)
		}
	}
	return nil
}

func TestCachedSessionRepositoryUsesCacheOnSecondGet(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	base := &countingSessionRepo{sessions: make(map[uuid.UUID]domain.Session)}
	repo := NewCachedSessionRepository(base, sessioncache.New(client, sessioncache.DefaultTTL))

	sessionID := uuid.New()
	userID := uuid.New()
	base.sessions[sessionID] = domain.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	ctx := context.Background()

	first, err := repo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("first get: %v", err)
	}
	if first.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, first.UserID)
	}
	if base.getByIDCalls != 1 {
		t.Fatalf("expected one base lookup, got %d", base.getByIDCalls)
	}

	second, err := repo.GetByID(ctx, sessionID)
	if err != nil {
		t.Fatalf("second get: %v", err)
	}
	if second.UserID != userID {
		t.Fatalf("expected cached user id %s, got %s", userID, second.UserID)
	}
	if base.getByIDCalls != 1 {
		t.Fatalf("expected cached second get, base calls = %d", base.getByIDCalls)
	}
}

func TestCachedSessionRepositoryInvalidatesOnRevoke(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start miniredis: %v", err)
	}
	defer mr.Close()

	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer client.Close()

	base := &countingSessionRepo{sessions: make(map[uuid.UUID]domain.Session)}
	repo := NewCachedSessionRepository(base, sessioncache.New(client, sessioncache.DefaultTTL))

	sessionID := uuid.New()
	userID := uuid.New()
	base.sessions[sessionID] = domain.Session{
		ID:        sessionID,
		UserID:    userID,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	ctx := context.Background()
	if _, err := repo.GetByID(ctx, sessionID); err != nil {
		t.Fatalf("warm cache: %v", err)
	}

	if err := repo.RevokeByID(ctx, sessionID, userID); err != nil {
		t.Fatalf("revoke: %v", err)
	}

	base.getByIDCalls = 0
	if _, err := repo.GetByID(ctx, sessionID); err != ErrNotFound {
		t.Fatalf("expected not found after revoke, got %v", err)
	}
	if base.getByIDCalls != 1 {
		t.Fatalf("expected base lookup after cache invalidation, got %d", base.getByIDCalls)
	}
}
