package repository

import (
	"context"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type SharedVaultItemRepository interface {
	Create(ctx context.Context, share domain.SharedVaultItem) (domain.SharedVaultItem, error)
	GetByVaultAndRecipient(ctx context.Context, vaultItemID, recipientID uuid.UUID) (domain.SharedVaultItem, error)
	Delete(ctx context.Context, vaultItemID, ownerID, recipientID uuid.UUID) error
	ListForRecipient(ctx context.Context, recipientID uuid.UUID, limit, offset int32) (ListSharedVaultItemsResult, error)
}

type ListSharedVaultItemsResult struct {
	Items []domain.SharedVaultItemWithItem
	Total int64
}
