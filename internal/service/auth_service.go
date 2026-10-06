package service

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrRefreshTokenReuse = errors.New("refresh token reuse detected")
var ErrEmailAlreadyExists = errors.New("email already exists")
var ErrPasswordUnchanged = errors.New("new password must differ from current password")

const (
	accessTokenTTL            = 15 * time.Minute
	refreshTokenRotationGrace = 30 * time.Second
)

type AuthService struct {
	users           repository.UserRepository
	sessions        repository.SessionRepository
	jwtSecret       string
	audit           *AuditService
	passwordChecker crypto.PasswordBreachChecker
}

func NewAuthService(
	users repository.UserRepository,
	sessions repository.SessionRepository,
	jwtSecret string,
	audit *AuditService,
	passwordChecker crypto.PasswordBreachChecker,
) *AuthService {
	return &AuthService{
		users:           users,
		sessions:        sessions,
		jwtSecret:       jwtSecret,
		audit:           audit,
		passwordChecker: passwordChecker,
	}
}

func (s *AuthService) validateNewPassword(ctx context.Context, password string) error {
	if err := crypto.ValidatePasswordStrength(password); err != nil {
		return err
	}
	if s.passwordChecker == nil {
		return nil
	}

	breached, err := s.passwordChecker.IsBreached(ctx, password)
	if err != nil {
		slog.Warn("password breach check failed; allowing signup", "error", err)
		return nil
	}
	if breached {
		return ErrCompromisedPassword
	}
	return nil
}

func (s *AuthService) GetProfile(ctx context.Context, userID uuid.UUID) (domain.User, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user: %w", err)
	}
	return user, nil
}

func (s *AuthService) Signup(ctx context.Context, email, password string, audit AuditContext) (domain.User, error) {
	if err := s.validateNewPassword(ctx, password); err != nil {
		return domain.User{}, err
	}

	hashPassword, err := crypto.HashPassword(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		Email:        email,
		PasswordHash: hashPassword,
	}

	_, err = s.users.GetByEmail(ctx, email)
	if err == nil {
		return domain.User{}, ErrEmailAlreadyExists
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}
	user, err = s.users.Create(ctx, user)
	if err != nil {
		return domain.User{}, fmt.Errorf("create user: %w", err)
	}

	if s.audit != nil {
		userID := user.ID
		s.audit.Log(ctx, user.ID, audit, AuditAuthSignup, "user", &userID, map[string]any{
			"email": user.Email,
		})
	}

	return user, nil
}

type LoginDeviceInfo struct {
	DeviceName string
	IPAddress  string
	UserAgent  string
}

func (s *AuthService) Login(ctx context.Context, email, password, totpCode string, device LoginDeviceInfo) (accessToken, refreshToken string, err error) {
	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("get user by email: %w", err)
	}

	ok, err := crypto.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		return "", "", fmt.Errorf("check password hash: %w", err)
	}
	if !ok {
		return "", "", ErrInvalidCredentials
	}

	if user.MfaEnabled {
		if strings.TrimSpace(totpCode) == "" {
			return "", "", ErrMFARequired
		}
		if !s.validateUserTOTP(user.MfaSecret, totpCode) {
			return "", "", ErrInvalidTOTPCode
		}
	}

	accessToken, refreshToken, _, err = s.CreateSessionTokens(ctx, user, device)
	if err != nil {
		return "", "", err
	}

	if s.audit != nil {
		userID := user.ID
		s.audit.Log(ctx, user.ID, AuditContext{
			IPAddress: device.IPAddress,
			UserAgent: device.UserAgent,
		}, AuditAuthLogin, "user", &userID, map[string]any{
			"device_name": device.DeviceName,
			"mfa_used":    user.MfaEnabled,
		})
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) CreateSessionTokens(ctx context.Context, user domain.User, device LoginDeviceInfo) (accessToken, refreshToken string, sessionID uuid.UUID, err error) {
	refreshToken, err = crypto.GenerateRefreshToken()
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("generate refresh token: %w", err)
	}
	tokenHash, err := crypto.HashToken(refreshToken)
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("hash token: %w", err)
	}
	session, err := s.sessions.Create(ctx, domain.Session{
		UserID:     user.ID,
		TokenHash:  tokenHash,
		DeviceName: device.DeviceName,
		IPAddress:  device.IPAddress,
		UserAgent:  device.UserAgent,
		CreatedAt:  time.Now(),
		ExpiresAt:  time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("create session: %w", err)
	}
	accessToken, err = crypto.MakeAccessToken(
		user.ID,
		session.ID,
		s.jwtSecret,
		accessTokenTTL,
	)
	if err != nil {
		return "", "", uuid.Nil, fmt.Errorf("make access token: %w", err)
	}
	return accessToken, refreshToken, session.ID, nil
}

func (s *AuthService) RevokeOtherSessions(ctx context.Context, userID, keepSessionID uuid.UUID) error {
	if err := s.sessions.RevokeAllExcept(ctx, userID, keepSessionID); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}
	return nil
}

func (s *AuthService) validateUserTOTP(stored *string, code string) bool {
	if stored == nil {
		return false
	}
	plain, err := crypto.DecryptMFASecret(*stored, s.jwtSecret)
	if err != nil {
		return false
	}
	return crypto.ValidateTOTPCode(plain, code)
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (accessToken, newRefreshToken string, err error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", "", ErrInvalidCredentials
	}
	tokenHash, err := crypto.HashToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("hash token: %w", err)
	}

	newRefreshToken, newTokenHash, err := s.newRefreshTokenPair()
	if err != nil {
		return "", "", err
	}

	session, err := s.sessions.GetByTokenHash(ctx, tokenHash)
	if err == nil {
		accessToken, newRefreshToken, err = s.rotateAndIssue(ctx, session, tokenHash, newRefreshToken, newTokenHash)
		return accessToken, newRefreshToken, err
	}
	if !errors.Is(err, repository.ErrNotFound) {
		return "", "", fmt.Errorf("get session by token hash: %w", err)
	}

	session, err = s.sessions.GetByPreviousTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", fmt.Errorf("get session by previous token hash: %w", err)
	}

	rotated, err := s.sessions.RotateFromPreviousToken(ctx, session.ID, tokenHash, newTokenHash, refreshTokenRotationGrace)
	if err != nil {
		return "", "", fmt.Errorf("rotate session token from previous: %w", err)
	}
	if rotated {
		accessToken, err = crypto.MakeAccessToken(session.UserID, session.ID, s.jwtSecret, accessTokenTTL)
		if err != nil {
			return "", "", fmt.Errorf("make access token: %w", err)
		}
		return accessToken, newRefreshToken, nil
	}

	if revokeErr := s.sessions.Revoke(ctx, session.ID); revokeErr != nil {
		slog.Warn("failed to revoke session after refresh token reuse", "session_id", session.ID, "error", revokeErr)
	}
	if s.audit != nil {
		sessionID := session.ID
		s.audit.Log(ctx, session.UserID, AuditContext{}, AuditAuthRefreshReuse, "session", &sessionID, nil)
	}
	return "", "", ErrRefreshTokenReuse
}

func (s *AuthService) newRefreshTokenPair() (refreshToken, tokenHash string, err error) {
	refreshToken, err = crypto.GenerateRefreshToken()
	if err != nil {
		return "", "", fmt.Errorf("generate refresh token: %w", err)
	}
	tokenHash, err = crypto.HashToken(refreshToken)
	if err != nil {
		return "", "", fmt.Errorf("hash token: %w", err)
	}
	return refreshToken, tokenHash, nil
}

func (s *AuthService) rotateAndIssue(
	ctx context.Context,
	session domain.Session,
	currentTokenHash, newRefreshToken, newTokenHash string,
) (accessToken, rotatedRefreshToken string, err error) {
	rotated, err := s.sessions.RotateToken(ctx, session.ID, currentTokenHash, newTokenHash)
	if err != nil {
		return "", "", fmt.Errorf("rotate session token: %w", err)
	}
	if !rotated {
		graceRotated, graceErr := s.sessions.RotateFromPreviousToken(ctx, session.ID, currentTokenHash, newTokenHash, refreshTokenRotationGrace)
		if graceErr != nil {
			return "", "", fmt.Errorf("rotate session token from previous: %w", graceErr)
		}
		if !graceRotated {
			return "", "", ErrInvalidCredentials
		}
	}

	accessToken, err = crypto.MakeAccessToken(session.UserID, session.ID, s.jwtSecret, accessTokenTTL)
	if err != nil {
		return "", "", fmt.Errorf("make access token: %w", err)
	}
	return accessToken, newRefreshToken, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID, userID uuid.UUID, audit AuditContext) error {
	if err := s.sessions.Revoke(ctx, sessionID); err != nil {
		return err
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditAuthLogout, "session", &sessionID, nil)
	}
	return nil
}

func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	return s.sessions.ListByUserID(ctx, userID)
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID, audit AuditContext) error {
	err := s.sessions.RevokeByID(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("revoke session: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditAuthSessionRevoke, "session", &sessionID, nil)
	}
	return nil
}

func (s *AuthService) ChangePassword(
	ctx context.Context,
	userID, sessionID uuid.UUID,
	currentPassword, newPassword, totpCode string,
	audit AuditContext,
) error {
	currentPassword = strings.TrimSpace(currentPassword)
	newPassword = strings.TrimSpace(newPassword)
	if currentPassword == "" || newPassword == "" {
		return ErrInvalidCredentials
	}
	if currentPassword == newPassword {
		return ErrPasswordUnchanged
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get user: %w", err)
	}

	ok, err := crypto.CheckPasswordHash(currentPassword, user.PasswordHash)
	if err != nil {
		return fmt.Errorf("check password hash: %w", err)
	}
	if !ok {
		return ErrInvalidCredentials
	}

	if err := s.validateNewPassword(ctx, newPassword); err != nil {
		return err
	}

	if user.MfaEnabled {
		if strings.TrimSpace(totpCode) == "" {
			return ErrMFARequired
		}
		if !s.validateUserTOTP(user.MfaSecret, totpCode) {
			return ErrInvalidTOTPCode
		}
	}

	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("hash password: %w", err)
	}

	if err := s.users.UpdatePassword(ctx, userID, passwordHash); err != nil {
		return fmt.Errorf("update password: %w", err)
	}

	if err := s.sessions.RevokeAllExcept(ctx, userID, sessionID); err != nil {
		return fmt.Errorf("revoke other sessions: %w", err)
	}

	if s.audit != nil {
		uid := userID
		s.audit.Log(ctx, userID, audit, AuditAuthPasswordChange, "user", &uid, map[string]any{
			"mfa_used": user.MfaEnabled,
		})
	}

	return nil
}

// ResetAccountPassword sets a new account password for email, signs out every session,
// and optionally turns off MFA. The vault master password is not changed.
func (s *AuthService) ResetAccountPassword(ctx context.Context, email, newPassword string, disableMFA bool) (domain.User, error) {
	email = strings.TrimSpace(email)
	if email == "" {
		return domain.User{}, ErrNotFound
	}

	user, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.User{}, ErrNotFound
		}
		return domain.User{}, fmt.Errorf("get user by email: %w", err)
	}

	if err := s.validateNewPassword(ctx, newPassword); err != nil {
		return domain.User{}, err
	}

	passwordHash, err := crypto.HashPassword(newPassword)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	if err := s.users.UpdatePassword(ctx, user.ID, passwordHash); err != nil {
		return domain.User{}, fmt.Errorf("update password: %w", err)
	}

	if disableMFA {
		if err := s.users.DisableMFA(ctx, user.ID); err != nil {
			return domain.User{}, fmt.Errorf("disable mfa: %w", err)
		}
		user.MfaEnabled = false
		user.MfaSecret = nil
	}

	if err := s.sessions.RevokeAllExcept(ctx, user.ID, uuid.Nil); err != nil {
		return domain.User{}, fmt.Errorf("revoke sessions: %w", err)
	}

	if s.audit != nil {
		userID := user.ID
		s.audit.Log(ctx, user.ID, AuditContext{}, AuditAuthPasswordChange, "user", &userID, map[string]any{
			"operator_reset": true,
			"mfa_disabled":   disableMFA,
		})
	}

	return user, nil
}
