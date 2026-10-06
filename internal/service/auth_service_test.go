package service

import (
	"context"
	"errors"
	"testing"
	"time"

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
		if session.TokenHash == tokenHash && session.RevokedAt == nil {
			return session, nil
		}
	}
	return domain.Session{}, repository.ErrNotFound
}

func (s *stubAuthSessionRepo) GetByPreviousTokenHash(_ context.Context, tokenHash string) (domain.Session, error) {
	for _, session := range s.sessions {
		if session.PreviousTokenHash == tokenHash && session.RevokedAt == nil {
			return session, nil
		}
	}
	return domain.Session{}, repository.ErrNotFound
}

func (s *stubAuthSessionRepo) RotateToken(_ context.Context, sessionID uuid.UUID, currentTokenHash, newTokenHash string) (bool, error) {
	session, ok := s.sessions[sessionID]
	if !ok || session.RevokedAt != nil || session.TokenHash != currentTokenHash {
		return false, nil
	}
	now := time.Now()
	session.PreviousTokenHash = session.TokenHash
	session.TokenHash = newTokenHash
	session.TokenRotatedAt = &now
	s.sessions[sessionID] = session
	return true, nil
}

func (s *stubAuthSessionRepo) RotateFromPreviousToken(
	_ context.Context,
	sessionID uuid.UUID,
	previousTokenHash, newTokenHash string,
	grace time.Duration,
) (bool, error) {
	session, ok := s.sessions[sessionID]
	if !ok || session.RevokedAt != nil || session.PreviousTokenHash != previousTokenHash {
		return false, nil
	}
	if session.TokenRotatedAt == nil || time.Since(*session.TokenRotatedAt) > grace {
		return false, nil
	}
	now := time.Now()
	session.PreviousTokenHash = session.TokenHash
	session.TokenHash = newTokenHash
	session.TokenRotatedAt = &now
	s.sessions[sessionID] = session
	return true, nil
}

func (s *stubAuthSessionRepo) GetByID(_ context.Context, id uuid.UUID) (domain.Session, error) {
	session, ok := s.sessions[id]
	if !ok {
		return domain.Session{}, repository.ErrNotFound
	}
	return session, nil
}

func (s *stubAuthSessionRepo) Revoke(_ context.Context, id uuid.UUID) error {
	session, ok := s.sessions[id]
	if !ok {
		return repository.ErrNotFound
	}
	now := time.Now()
	session.RevokedAt = &now
	s.sessions[id] = session
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
	svc := NewAuthService(users, sessions, "test-secret", nil, nil)

	userID := uuid.New()
	currentSessionID := uuid.New()
	otherSessionID := uuid.New()

	hash, err := crypto.HashPassword("Old-Password12")
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

	err = svc.ChangePassword(context.Background(), userID, currentSessionID, "Old-Password12", "New-Password456", "", AuditContext{})
	if err != nil {
		t.Fatalf("change password: %v", err)
	}

	ok, err := crypto.CheckPasswordHash("New-Password456", users.users[userID].PasswordHash)
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
	svc := NewAuthService(users, sessions, "test-secret", nil, nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("Old-Password12")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users.users[userID] = domain.User{ID: userID, PasswordHash: hash}

	err = svc.ChangePassword(context.Background(), userID, uuid.New(), "Wrong-Password1", "New-Password456", "", AuditContext{})
	if !errors.Is(err, ErrInvalidCredentials) {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthServiceChangePasswordRejectsUnchangedPassword(t *testing.T) {
	users := newStubAuthUserRepo()
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(users, sessions, "test-secret", nil, nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("Same-Password1")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	users.users[userID] = domain.User{ID: userID, PasswordHash: hash}

	err = svc.ChangePassword(context.Background(), userID, uuid.New(), "Same-Password1", "Same-Password1", "", AuditContext{})
	if !errors.Is(err, ErrPasswordUnchanged) {
		t.Fatalf("expected ErrPasswordUnchanged, got %v", err)
	}
}

type stubPasswordChecker struct {
	breached bool
	err      error
}

func (s stubPasswordChecker) IsBreached(_ context.Context, _ string) (bool, error) {
	return s.breached, s.err
}

func TestAuthServiceSignupRejectsWeakPassword(t *testing.T) {
	svc := NewAuthService(newStubAuthUserRepo(), newStubAuthSessionRepo(), "test-secret", nil, nil)

	_, err := svc.Signup(context.Background(), "user@example.com", "short", AuditContext{})
	if !errors.Is(err, crypto.ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestAuthServiceSignupRejectsCompromisedPassword(t *testing.T) {
	svc := NewAuthService(
		newStubAuthUserRepo(),
		newStubAuthSessionRepo(),
		"test-secret",
		nil,
		stubPasswordChecker{breached: true},
	)

	_, err := svc.Signup(context.Background(), "user@example.com", "StrongPass123", AuditContext{})
	if !errors.Is(err, ErrCompromisedPassword) {
		t.Fatalf("expected ErrCompromisedPassword, got %v", err)
	}
}

func TestAuthServiceRefreshRotatesToken(t *testing.T) {
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(newStubAuthUserRepo(), sessions, "test-secret", nil, nil)

	userID := uuid.New()
	sessionID := uuid.New()
	refreshToken, err := crypto.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	tokenHash, err := crypto.HashToken(refreshToken)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}
	sessions.sessions[sessionID] = domain.Session{
		ID:        sessionID,
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(time.Hour),
	}

	accessToken, newRefreshToken, err := svc.Refresh(context.Background(), refreshToken)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if accessToken == "" || newRefreshToken == "" || newRefreshToken == refreshToken {
		t.Fatalf("expected rotated tokens, got access=%q refresh=%q", accessToken, newRefreshToken)
	}

	newHash, err := crypto.HashToken(newRefreshToken)
	if err != nil {
		t.Fatalf("hash new token: %v", err)
	}
	stored := sessions.sessions[sessionID]
	if stored.TokenHash != newHash {
		t.Fatalf("expected stored token hash to update")
	}
	if stored.PreviousTokenHash != tokenHash {
		t.Fatalf("expected previous token hash to be retained")
	}
}

func TestAuthServiceRefreshRejectsReusedToken(t *testing.T) {
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(newStubAuthUserRepo(), sessions, "test-secret", nil, nil)

	userID := uuid.New()
	sessionID := uuid.New()
	oldRefresh, err := crypto.GenerateRefreshToken()
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}
	oldHash, err := crypto.HashToken(oldRefresh)
	if err != nil {
		t.Fatalf("hash token: %v", err)
	}
	rotatedAt := time.Now().Add(-time.Minute)
	sessions.sessions[sessionID] = domain.Session{
		ID:                sessionID,
		UserID:            userID,
		TokenHash:         "current-hash",
		PreviousTokenHash: oldHash,
		TokenRotatedAt:    &rotatedAt,
		ExpiresAt:         time.Now().Add(time.Hour),
	}

	_, _, err = svc.Refresh(context.Background(), oldRefresh)
	if !errors.Is(err, ErrRefreshTokenReuse) {
		t.Fatalf("expected ErrRefreshTokenReuse, got %v", err)
	}
	if sessions.sessions[sessionID].RevokedAt == nil {
		t.Fatal("expected session to be revoked after reuse")
	}
}

func TestAuthServiceSignupAllowsPasswordWhenBreachCheckFails(t *testing.T) {
	svc := NewAuthService(
		newStubAuthUserRepo(),
		newStubAuthSessionRepo(),
		"test-secret",
		nil,
		stubPasswordChecker{err: errors.New("hibp unavailable")},
	)

	user, err := svc.Signup(context.Background(), "user@example.com", "StrongPass123", AuditContext{})
	if err != nil {
		t.Fatalf("expected signup to succeed when breach check fails, got %v", err)
	}
	if user.Email != "user@example.com" {
		t.Fatalf("expected user email, got %q", user.Email)
	}
}

func TestAuthServiceResetAccountPassword(t *testing.T) {
	users := newStubAuthUserRepo()
	sessions := newStubAuthSessionRepo()
	svc := NewAuthService(users, sessions, "test-secret", nil, nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("Old-Password12")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	secret := "totp-secret"
	users.users[userID] = domain.User{
		ID:           userID,
		Email:        "person@example.com",
		PasswordHash: hash,
		MfaEnabled:   true,
		MfaSecret:    &secret,
	}
	sessionID := uuid.New()
	sessions.sessions[sessionID] = domain.Session{ID: sessionID, UserID: userID}

	updated, err := svc.ResetAccountPassword(context.Background(), " person@example.com ", "New-Password12", false)
	if err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if !updated.MfaEnabled {
		t.Fatal("expected MFA to stay enabled")
	}
	ok, err := crypto.CheckPasswordHash("New-Password12", users.users[userID].PasswordHash)
	if err != nil || !ok {
		t.Fatalf("new password hash mismatch: ok=%v err=%v", ok, err)
	}
	if _, stillThere := sessions.sessions[sessionID]; stillThere {
		t.Fatal("expected existing session to be revoked")
	}
}

func TestAuthServiceResetAccountPasswordDisableMFA(t *testing.T) {
	users := newStubAuthUserRepo()
	svc := NewAuthService(users, newStubAuthSessionRepo(), "test-secret", nil, nil)

	userID := uuid.New()
	hash, err := crypto.HashPassword("Old-Password12")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	secret := "totp-secret"
	users.users[userID] = domain.User{
		ID:           userID,
		Email:        "person@example.com",
		PasswordHash: hash,
		MfaEnabled:   true,
		MfaSecret:    &secret,
	}

	updated, err := svc.ResetAccountPassword(context.Background(), "person@example.com", "New-Password12", true)
	if err != nil {
		t.Fatalf("reset password: %v", err)
	}
	if updated.MfaEnabled || updated.MfaSecret != nil {
		t.Fatal("expected MFA to be turned off")
	}
	if users.users[userID].MfaEnabled || users.users[userID].MfaSecret != nil {
		t.Fatal("expected stored MFA to be cleared")
	}
}

func TestAuthServiceResetAccountPasswordRejectsUnknownEmailAndWeakPassword(t *testing.T) {
	users := newStubAuthUserRepo()
	svc := NewAuthService(users, newStubAuthSessionRepo(), "test-secret", nil, nil)

	_, err := svc.ResetAccountPassword(context.Background(), "missing@example.com", "New-Password12", false)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	userID := uuid.New()
	users.users[userID] = domain.User{ID: userID, Email: "person@example.com"}
	_, err = svc.ResetAccountPassword(context.Background(), "person@example.com", "short", false)
	if !errors.Is(err, crypto.ErrWeakPassword) {
		t.Fatalf("expected weak password error, got %v", err)
	}
}
