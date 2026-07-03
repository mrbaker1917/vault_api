package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

type CreateVaultItemInput struct {
	EncryptedData []byte
	ItemType      string
	Title         string
	Folder        string
	Tags          []string
}

type VaultService struct {
	vaultItems repository.VaultItemRepository
}

func NewVaultService(vaultItems repository.VaultItemRepository) *VaultService {
	return &VaultService{vaultItems: vaultItems}
}

func (s *VaultService) CreateItem(ctx context.Context, userID uuid.UUID, input CreateVaultItemInput) (domain.VaultItem, error) {
	now := time.Now()
	item, err := s.vaultItems.Create(ctx, domain.VaultItem{
		UserID:        userID,
		EncryptedData: input.EncryptedData,
		ItemType:      input.ItemType,
		Title:         input.Title,
		Folder:        input.Folder,
		Tags:          input.Tags,
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	})
	if err != nil {
		return domain.VaultItem{}, fmt.Errorf("create vault item: %w", err)
	}
	return item, nil
}

func (s *VaultService) GetItem(ctx context.Context, userID, itemID uuid.UUID) (domain.VaultItem, error) {
	item, err := s.vaultItems.GetByID(ctx, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("get vault item: %w", err)
	}
	if item.UserID != userID {
		return domain.VaultItem{}, ErrNotFound
	}
	return item, nil
}

func (s *VaultService) ListItems(ctx context.Context, userID uuid.UUID) ([]domain.VaultItem, error) {
	items, err := s.vaultItems.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list vault items: %w", err)
	}
	return items, nil
}
