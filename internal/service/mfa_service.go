package service

import (
	"context"
	"errors"
	"fmt"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

var ErrMFANotEnabled = errors.New("mfa not enabled")
var ErrInvalidTOTPCode = errors.New("invalid totp code")

type MFAService struct {}