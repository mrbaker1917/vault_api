package service

import (
	"context"
	"errors"
	"fmt"

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
	users repository.UserRepository
	audit *AuditService
}

func NewMFAService(users repository.UserRepository, audit *AuditService) *MFAService {
	return &MFAService{users: users, audit: audit}
}

type MFASetup struct {
	Secret     string
	OTPAuthURL string
}

func (s *MFAService) EnableMFA(ctx context.Context, userID uuid.UUID, audit AuditContext) (MFASetup, error) {
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

	setup, err := crypto.GenerateTOTPSecret(user.Email)
	if err != nil {
		return MFASetup{}, fmt.Errorf("generate totp secret: %w", err)
	}

	if err := s.users.EnableMFASecret(ctx, userID, setup.Secret); err != nil {
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
	if user.MfaSecret == nil {
		return ErrMFANotEnabled
	}
	if !crypto.ValidateTOTPCode(*user.MfaSecret, code) {
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
	if user.MfaSecret == nil || !crypto.ValidateTOTPCode(*user.MfaSecret, code) {
		return ErrInvalidTOTPCode
	}
	if err := s.users.DisableMFA(ctx, userID); err != nil {
		return fmt.Errorf("disable mfa: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditMFADisable, "user", &userID, nil)
	}

	return nil
}
