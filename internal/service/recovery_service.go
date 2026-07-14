package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"vault_api/internal/crypto"
	"vault_api/internal/repository"
)

var ErrInvalidRecoveryCode = errors.New("invalid recovery code")
var ErrRecoveryRequiresMFA = errors.New("recovery codes require mfa to be enabled")

type RecoveryService struct {
	users         repository.UserRepository
	recoveryCodes repository.RecoveryCodeRepository
	auth          *AuthService
	audit         *AuditService
}

func NewRecoveryService(
	users repository.UserRepository,
	recoveryCodes repository.RecoveryCodeRepository,
	auth *AuthService,
	audit *AuditService,
) *RecoveryService {
	return &RecoveryService{
		users:         users,
		recoveryCodes: recoveryCodes,
		auth:          auth,
		audit:         audit,
	}
}

func (s *RecoveryService) GenerateRecoveryCodes(ctx context.Context, userID uuid.UUID, audit AuditContext) ([]string, error) {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}
	if !user.MfaEnabled {
		return nil, ErrRecoveryRequiresMFA
	}

	if err := s.recoveryCodes.DeleteByUserID(ctx, userID); err != nil {
		return nil, fmt.Errorf("delete existing recovery codes: %w", err)
	}

	plaintextCodes := make([]string, 0, crypto.RecoveryCodeCount())
	for range crypto.RecoveryCodeCount() {
		code, err := crypto.GenerateRecoveryCode()
		if err != nil {
			return nil, fmt.Errorf("generate recovery code: %w", err)
		}

		codeHash, err := crypto.HashToken(normalizeRecoveryCode(code))
		if err != nil {
			return nil, fmt.Errorf("hash recovery code: %w", err)
		}

		if _, err := s.recoveryCodes.Create(ctx, userID, codeHash); err != nil {
			return nil, fmt.Errorf("store recovery code: %w", err)
		}
		plaintextCodes = append(plaintextCodes, code)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditRecoveryGenerate, "user", &userID, map[string]any{
			"code_count": len(plaintextCodes),
		})
	}

	return plaintextCodes, nil
}

func (s *RecoveryService) VerifyRecoveryLogin(
	ctx context.Context,
	email, password, recoveryCode string,
	device LoginDeviceInfo,
) (accessToken, refreshToken string, err error) {
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
	if !user.MfaEnabled {
		return "", "", ErrRecoveryRequiresMFA
	}

	normalized := normalizeRecoveryCode(recoveryCode)
	if normalized == "" {
		return "", "", ErrInvalidRecoveryCode
	}

	codeHash, err := crypto.HashToken(normalized)
	if err != nil {
		return "", "", fmt.Errorf("hash recovery code: %w", err)
	}

	unusedCodes, err := s.recoveryCodes.ListUnusedByUserID(ctx, user.ID)
	if err != nil {
		return "", "", fmt.Errorf("list recovery codes: %w", err)
	}

	var matchedID uuid.UUID
	for _, stored := range unusedCodes {
		if stored.CodeHash == codeHash {
			matchedID = stored.ID
			break
		}
	}
	if matchedID == uuid.Nil {
		return "", "", ErrInvalidRecoveryCode
	}

	if err := s.recoveryCodes.MarkUsed(ctx, matchedID); err != nil {
		return "", "", fmt.Errorf("mark recovery code used: %w", err)
	}

	accessToken, refreshToken, err = s.auth.CreateSessionTokens(ctx, user, device)
	if err != nil {
		return "", "", err
	}

	if s.audit != nil {
		userID := user.ID
		s.audit.Log(ctx, user.ID, AuditContext{
			IPAddress: device.IPAddress,
			UserAgent: device.UserAgent,
		}, AuditRecoveryVerify, "user", &userID, map[string]any{
			"device_name": device.DeviceName,
		})
	}

	return accessToken, refreshToken, nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToUpper(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
}
