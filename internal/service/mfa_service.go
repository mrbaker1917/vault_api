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

var (
	ErrMFANotEnabled     = errors.New("mfa not enabled")
	ErrMFAAlreadyEnabled = errors.New("mfa already enabled")
	ErrInvalidTOTPCode   = errors.New("invalid totp code")
	ErrMFARequired       = errors.New("mfa required")
)

type MFAService struct {
	users         repository.UserRepository
	recoveryCodes repository.RecoveryCodeRepository
	jwtSecret     string
	audit         *AuditService
}

func NewMFAService(
	users repository.UserRepository,
	recoveryCodes repository.RecoveryCodeRepository,
	jwtSecret string,
	audit *AuditService,
) *MFAService {
	return &MFAService{
		users:         users,
		recoveryCodes: recoveryCodes,
		jwtSecret:     jwtSecret,
		audit:         audit,
	}
}

type MFASetup struct {
	Secret     string
	OTPAuthURL string
}

func (s *MFAService) EnableMFA(ctx context.Context, userID uuid.UUID, audit AuditContext, password string) (MFASetup, error) {
	password = strings.TrimSpace(password)
	if password == "" {
		return MFASetup{}, ErrInvalidCredentials
	}

	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return MFASetup{}, ErrNotFound
		}
		return MFASetup{}, fmt.Errorf("get user by id: %w", err)
	}
	if user.MfaEnabled {
		return MFASetup{}, ErrMFAAlreadyEnabled
	}

	ok, err := crypto.CheckPasswordHash(password, user.PasswordHash)
	if err != nil {
		return MFASetup{}, fmt.Errorf("check password hash: %w", err)
	}
	if !ok {
		return MFASetup{}, ErrInvalidCredentials
	}

	setup, err := crypto.GenerateTOTPSecret(user.Email)
	if err != nil {
		return MFASetup{}, fmt.Errorf("generate totp secret: %w", err)
	}

	encryptedSecret, err := crypto.EncryptMFASecret(setup.Secret, s.jwtSecret)
	if err != nil {
		return MFASetup{}, fmt.Errorf("encrypt mfa secret: %w", err)
	}

	if err := s.users.EnableMFASecret(ctx, userID, encryptedSecret); err != nil {
		return MFASetup{}, fmt.Errorf("store mfa secret: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditMFAEnable, "user", &userID, nil)
	}

	return MFASetup{
		Secret:     setup.Secret,
		OTPAuthURL: setup.OTPAuthURL,
	}, nil
}

func (s *MFAService) VerifyMFA(ctx context.Context, userID uuid.UUID, audit AuditContext, code string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get user by id: %w", err)
	}
	if user.MfaEnabled {
		return ErrMFAAlreadyEnabled
	}
	plainSecret, ok := s.resolveTOTPSecret(user.MfaSecret)
	if !ok {
		return ErrMFANotEnabled
	}
	if !crypto.ValidateTOTPCode(plainSecret, code) {
		return ErrInvalidTOTPCode
	}
	if err := s.users.ConfirmMFA(ctx, userID); err != nil {
		return fmt.Errorf("confirm mfa: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditMFAVerify, "user", &userID, nil)
	}

	return nil
}

func (s *MFAService) DisableMFA(ctx context.Context, userID uuid.UUID, audit AuditContext, code string) error {
	user, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get user by id: %w", err)
	}
	if !user.MfaEnabled {
		return ErrMFANotEnabled
	}
	plainSecret, ok := s.resolveTOTPSecret(user.MfaSecret)
	if !ok || !crypto.ValidateTOTPCode(plainSecret, code) {
		return ErrInvalidTOTPCode
	}
	if err := s.users.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("disable mfa: %w", err)
	}
	if s.recoveryCodes != nil {
		if err := s.recoveryCodes.DeleteByUserID(ctx, userID); err != nil {
			return fmt.Errorf("delete recovery codes: %w", err)
		}
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditMFADisable, "user", &userID, nil)
	}

	return nil
}

func (s *MFAService) resolveTOTPSecret(stored *string) (string, bool) {
	if stored == nil {
		return "", false
	}
	plain, err := crypto.DecryptMFASecret(*stored, s.jwtSecret)
	if err != nil || plain == "" {
		return "", false
	}
	return plain, true
}
