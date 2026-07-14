package repository

import (
	"context"

	"github.com/google/uuid"
	"vault_api/internal/domain"
)

type VaultItemRepository interface {
	Create(ctx context.Context, item domain.VaultItem) (domain.VaultItem, error)
	GetByID(ctx context.Context, id uuid.UUID) (domain.VaultItem, error)
	ListByUserID(ctx context.Context, userID uuid.UUID, filter ListVaultItemsFilter) (ListVaultItemsResult, error)
	Update(ctx context.Context, item domain.VaultItem) (domain.VaultItem, error)
	Delete(ctx context.Context, id uuid.UUID, version int32) (domain.VaultItem, error)
	Restore(ctx context.Context, id uuid.UUID, version int32, userID uuid.UUID) (domain.VaultItem, error)
}

type ListVaultItemsFilter struct {
	Folder   string
	ItemType string
	Tag      string
	Title    string
	Limit    int32
	Offset   int32
}

type ListVaultItemsResult struct {
	Items []domain.VaultItem
	Total int64
}
