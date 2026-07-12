package service

import (
	"context"
	"fmt"
	"errors"
	"time"
	"strings"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
	"github.com/google/uuid"
)

var ErrInvalidCredentials = errors.New("invalid credentials")
var ErrEmailAlreadyExists = errors.New("email already exists")

const accessTokenTTL = 15 * time.Minute

type AuthService struct {
	users     repository.UserRepository
	sessions  repository.SessionRepository
	jwtSecret string
}

func NewAuthService(users repository.UserRepository, sessions repository.SessionRepository, jwtSecret string) *AuthService {
	return &AuthService{
		users:     users,
		sessions:  sessions,
		jwtSecret: jwtSecret,
	}
}

func (s *AuthService) Signup(ctx context.Context, email, password string) (domain.User, error) {
	hashPassword, err := crypto.HashPassword(password)
	if err != nil {
		return domain.User{}, fmt.Errorf("hash password: %w", err)
	}
	user := domain.User{
		Email: email,
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
		if user.MfaSecret == nil || !crypto.ValidateTOTPCode(*user.MfaSecret, totpCode) {
			return "", "", ErrInvalidTOTPCode
		}
	}

	refreshToken, err = crypto.GenerateRefreshToken()
    if err != nil {
        return "", "", fmt.Errorf("generate refresh token: %w", err)
    }
    tokenHash, err := crypto.HashToken(refreshToken)
	if err != nil {
        return "", "", fmt.Errorf("hash token: %w", err)
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
        return "", "", fmt.Errorf("create session: %w", err)
    }
    accessToken, err = crypto.MakeAccessToken(
        user.ID,
        session.ID,
        s.jwtSecret,
        accessTokenTTL,
    )
    if err != nil {
        return "", "", fmt.Errorf("make access token: %w", err)
    }
    return accessToken, refreshToken, nil
}

func (s *AuthService) Refresh(ctx context.Context, refreshToken string) (accessToken string, err error) {
	if strings.TrimSpace(refreshToken) == "" {
		return "", ErrInvalidCredentials
	}
	tokenHash, err := crypto.HashToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("hash token: %w", err)
	}
	session, err := s.sessions.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", fmt.Errorf("get session by token hash: %w", err)
	}

	accessToken, err = crypto.MakeAccessToken(
		session.UserID,
		session.ID,
		s.jwtSecret,
		accessTokenTTL,
	)
	if err != nil {
		return "", fmt.Errorf("make access token: %w", err)
	}
	return accessToken, nil
}

func (s *AuthService) Logout(ctx context.Context, sessionID uuid.UUID) error {
	return s.sessions.Revoke(ctx, sessionID)
}

func (s *AuthService) ListSessions(ctx context.Context, userID uuid.UUID) ([]domain.Session, error) {
	return s.sessions.ListByUserID(ctx, userID)
}

func (s *AuthService) RevokeSession(ctx context.Context, sessionID, userID uuid.UUID) error {
	err := s.sessions.RevokeByID(ctx, sessionID, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}