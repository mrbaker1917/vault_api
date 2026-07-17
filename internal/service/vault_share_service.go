package service

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

type ShareVaultItemInput struct {
	Email            string
	EncryptedItemKey []byte
	Permission       string
}

type ListSharedVaultItemsResult struct {
	Items  []domain.SharedVaultItemWithItem
	Total  int64
	Limit  int32
	Offset int32
}

func (s *VaultService) ShareItem(
	ctx context.Context,
	ownerID, itemID uuid.UUID,
	audit AuditContext,
	input ShareVaultItemInput,
) (domain.SharedVaultItem, error) {
	if err := crypto.ValidateEncryptedItemKey(input.EncryptedItemKey); err != nil {
		return domain.SharedVaultItem{}, fmt.Errorf("%w: %w", ErrInvalidEncryptedItemKey, err)
	}

	permission := strings.ToLower(strings.TrimSpace(input.Permission))
	if permission != domain.SharePermissionRead && permission != domain.SharePermissionWrite {
		return domain.SharedVaultItem{}, ErrInvalidSharePermission
	}

	email := strings.TrimSpace(strings.ToLower(input.Email))
	if email == "" {
		return domain.SharedVaultItem{}, fmt.Errorf("recipient email is required")
	}

	item, err := s.vaultItems.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.SharedVaultItem{}, ErrNotFound
		}
		return domain.SharedVaultItem{}, fmt.Errorf("get vault item: %w", err)
	}
	if item.UserID != ownerID {
		return domain.SharedVaultItem{}, ErrNotFound
	}

	recipient, err := s.users.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.SharedVaultItem{}, ErrNotFound
		}
		return domain.SharedVaultItem{}, fmt.Errorf("get recipient: %w", err)
	}
	if recipient.ID == ownerID {
		return domain.SharedVaultItem{}, ErrCannotShareWithSelf
	}

	if _, err := s.sharedItems.GetByVaultAndRecipient(ctx, itemID, recipient.ID); err == nil {
		return domain.SharedVaultItem{}, ErrAlreadyShared
	} else if !errors.Is(err, repository.ErrNotFound) {
		return domain.SharedVaultItem{}, fmt.Errorf("check existing share: %w", err)
	}

	share, err := s.sharedItems.Create(ctx, domain.SharedVaultItem{
		VaultItemID:      itemID,
		OwnerID:          ownerID,
		SharedWithUserID: recipient.ID,
		EncryptedItemKey: input.EncryptedItemKey,
		Permission:       permission,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		return domain.SharedVaultItem{}, fmt.Errorf("create share: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, ownerID, audit, AuditVaultItemShare, "vault_item", &itemID, map[string]any{
			"shared_with_user_id": recipient.ID.String(),
			"permission":            permission,
		})
	}

	return share, nil
}

func (s *VaultService) RevokeShare(
	ctx context.Context,
	ownerID, itemID, recipientID uuid.UUID,
	audit AuditContext,
) error {
	item, err := s.vaultItems.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("get vault item: %w", err)
	}
	if item.UserID != ownerID {
		return ErrNotFound
	}

	if err := s.sharedItems.Delete(ctx, itemID, ownerID, recipientID); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return ErrNotFound
		}
		return fmt.Errorf("revoke share: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, ownerID, audit, AuditVaultItemShareRevoke, "vault_item", &itemID, map[string]any{
			"shared_with_user_id": recipientID.String(),
		})
	}

	return nil
}

func (s *VaultService) ListSharedWithMe(ctx context.Context, userID uuid.UUID, limit, offset int32) (ListSharedVaultItemsResult, error) {
	if limit <= 0 {
		limit = defaultVaultListLimit
	}
	if limit > maxVaultListLimit {
		limit = maxVaultListLimit
	}
	if offset < 0 {
		offset = 0
	}

	result, err := s.sharedItems.ListForRecipient(ctx, userID, limit, offset)
	if err != nil {
		return ListSharedVaultItemsResult{}, fmt.Errorf("list shared vault items: %w", err)
	}

	return ListSharedVaultItemsResult{
		Items:  result.Items,
		Total:  result.Total,
		Limit:  limit,
		Offset: offset,
	}, nil
}
