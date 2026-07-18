package service

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

type stubAuthUserRepo struct {
	users map[uuid.UUID]domain.User
}

func newStubAuthUserRepo() *stubAuthUserRepo {
	return &stubAuthUserRepo{users: make(map[uuid.UUID]domain.User)}
}

func (s *stubAuthUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *stubAuthUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, repository.ErrNotFound
}

func (s *stubAuthUserRepo) GetByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	user, ok := s.users[id]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *stubAuthUserRepo) EnableMFASecret(_ context.Context, id uuid.UUID, secret string) error {
	user, ok := s.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	user.MfaSecret = &secret
	s.users[id] = user
	return nil
}

func (s *stubAuthUserRepo) ConfirmMFA(_ context.Context, id uuid.UUID) error {
	user, ok := s.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	user.MfaEnabled = true
	s.users[id] = user
	return nil
}

func (s *stubAuthUserRepo) DisableMFA(_ context.Context, id uuid.UUID) error {
	user, ok := s.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	user.MfaEnabled = false
	user.MfaSecret = nil
	s.users[id] = user
	return nil
}

func (s *stubAuthUserRepo) UpdatePassword(_ context.Context, id uuid.UUID, passwordHash string) error {
	user, ok := s.users[id]
	if !ok {
		return repository.ErrNotFound
	}
	user.PasswordHash = passwordHash
	s.users[id] = user
	return nil
}

type stubAuthSessionRepo struct {
	sessions          map[uuid.UUID]domain.Session
	revokedExceptCall int
}

func newStubAuthSessionRepo() *stubAuthSessionRepo {
	return &stubAuthSessionRepo{sessions: make(map[uuid.UUID]domain.Session)}
}

func (s *stubAuthSessionRepo) Create(_ context.Context, session domain.Session) (domain.Session, error) {
	if session.ID == uuid.Nil {
		session.ID = uuid.New()
	}
	s.sessions[session.ID] = session
	return session, nil
}

func (s *stubAuthSessionRepo) GetByTokenHash(_ context.Context, tokenHash string) (domain.Session, error) {
	for _, session := range s.sessions {
		if session.TokenHash == tokenHash {
			return session, nil
		}
	}
	return domain.Session{}, repository.ErrNotFound
}

func (s *stubAuthSessionRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Session, error) {
	session, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, repository.ErrNotFound
	}
	return session, nil
}

func (s *stubAuthSessionRepo) Revoke(_ context.Context, id uuid.UUID) error {
	delete(s.sessions, id)
	return nil
}

func (s *stubAuthSessionRepo) ListByUserID(_ context.Context, userID uuid.UUID) ([]domain.Session, error) {
	out := make([]domain.Session, 0)
	for _, session := range s.sessions {
		if session.UserID == userID {
			out = append(out, session)
		}
	}
	return out, nil
}

func (s *stubAuthSessionRepo) RevokeByID(_ context.Context, id, userID uuid.UUID) error {
	session, ok := s.sessions[id]
	if !ok || session.UserID != userID {
		return repository.ErrNotFound
	}
	delete(s.sessions, id)
	return nil
}

func (s *stubAuthSessionRepo) RevokeAllExcept(_ context.Context, userID, exceptSessionID uuid.UUID) error {
	s.revokedExceptCall++
	for id, session := range s.sessions {
		if session.UserID == userID && id != exceptSessionID {
			delete(s.sessions, id)
		}
	}
	return nil
}

func TestAuthServiceChangePassword(t *testing.T) {
	users := newStubAuthUserRepo()
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(users, sessions, "test-secret", nil)

	userID := uuid.New()
	currentSessionID := uuid.New()
	otherSessionID := uuid.New()

	hash, err := crypto.HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users.users[userID] = domain.User{
		ID:           userID,
		Email:        "user@example.com",
		PasswordHash: hash,
	}
	sessions.sessions[currentSessionID] = domain.Session{ID: currentSessionID, UserID: userID}
	sessions.sessions[otherSessionID] = domain.Session{ID: otherSessionID, UserID: userID}

	err = svc.ChangePassword(context.Background(), userID, currentSessionID, "old-password", "new-password-456", "", AuditContext{})
	if err != nil {
		t.Fatalf("change password: %v", err)
	}

	ok, err := crypto.CheckPasswordHash("new-password-456", users.users[userID].PasswordHash)
	if err != nil || !ok {
		t.Fatal("expected updated password hash to match new password")
	}
	if sessions.revokedExceptCall != 1 {
		t.Fatalf("expected revoke other sessions once, got %d", sessions.revokedExceptCall)
	}
	if _, ok := sessions.sessions[currentSessionID]; !ok {
		t.Fatal("expected current session to remain active")
	}
	if _, ok := sessions.sessions[otherSessionID]; ok {
		t.Fatal("expected other session to be revoked")
	}
}

func TestAuthServiceChangePasswordRejectsWrongCurrentPassword(t *testing.T) {
	users := newStubAuthUserRepo()
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(users, sessions, "test-secret", nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("old-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users.users[userID] = domain.User{ID: userID, PasswordHash: hash}

	err = svc.ChangePassword(context.Background(), userID, uuid.New(), "wrong-password", "new-password-456", "", AuditContext{})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceChangePasswordRejectsUnchangedPassword(t *testing.T) {
	users := newStubAuthUserRepo()
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(users, sessions, "test-secret", nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("same-password")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users.users[userID] = domain.User{ID: userID, PasswordHash: hash}

	err = svc.ChangePassword(context.Background(), userID, uuid.New(), "same-password", "same-password", "", AuditContext{})
	if !errors.Is(err, ErrPasswordUnchanged) {
		t.Fatalf("expected ErrPasswordUnchanged, got %v", err)
	}
}
