package service

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

const (
	AuditAuthSignup         = "auth.signup"
	AuditAuthLogin          = "auth.login"
	AuditAuthLogout         = "auth.logout"
	AuditAuthPasswordChange = "auth.password.change"
	AuditAuthSessionRevoke  = "auth.session.revoke"
	AuditVaultItemCreate    = "vault.item.create"
	AuditVaultItemUpdate    = "vault.item.update"
	AuditVaultItemDelete    = "vault.item.delete"
	AuditVaultItemRestore   = "vault.item.restore"
	AuditVaultItemShare     = "vault.item.share"
	AuditVaultItemShareRevoke = "vault.item.share.revoke"
	AuditMFAEnable          = "mfa.enable"
	AuditMFAVerify          = "mfa.verify"
	AuditMFADisable         = "mfa.disable"
	AuditRecoveryGenerate   = "recovery.generate"
	AuditRecoveryVerify     = "recovery.verify"
)

const (
	defaultAuditLimit = 50
	maxAuditLimit     = 100
)

func DefaultAuditLimit() int32 {
	return defaultAuditLimit
}

type AuditContext struct {
	IPAddress string
	UserAgent string
}

type AuditService struct {
	auditLogs repository.AuditLogRepository
}

func NewAuditService(auditLogs repository.AuditLogRepository) *AuditService {
	return &AuditService{auditLogs: auditLogs}
}

func (s *AuditService) Log(
	ctx context.Context,
	userID uuid.UUID,
	audit AuditContext,
	action, resourceType string,
	resourceID *uuid.UUID,
	metadata map[string]any,
) {
	var metadataJSON json.RawMessage
	if metadata != nil {
		encoded, err := json.Marshal(metadata)
		if err != nil {
			slog.Warn("audit log metadata encode failed", "action", action, "error", err)
			return
		}
		metadataJSON = encoded
	}

	uid := userID
	_, err := s.auditLogs.Create(ctx, domain.AuditLog{
		UserID:       &uid,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		IPAddress:    audit.IPAddress,
		UserAgent:    audit.UserAgent,
		Metadata:     metadataJSON,
	})
	if err != nil {
		slog.Warn("audit log write failed", "action", action, "error", err)
	}
}

func (s *AuditService) ListLogs(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]domain.AuditLog, error) {
	if limit <= 0 {
		limit = defaultAuditLimit
	}
	if limit > maxAuditLimit {
		limit = maxAuditLimit
	}
	if offset < 0 {
		offset = 0
	}

	logs, err := s.auditLogs.ListByUserID(ctx, userID, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list audit logs: %w", err)
	}
	return logs, nil
}
