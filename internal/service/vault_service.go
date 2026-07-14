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

const (
	defaultVaultListLimit = 50
	maxVaultListLimit     = 100
)

type ListVaultItemsFilter struct {
	Folder   string
	ItemType string
	Tag      string
	Title    string
	Limit    int32
	Offset   int32
}

type ListVaultItemsResult struct {
	Items  []domain.VaultItem
	Total  int64
	Limit  int32
	Offset int32
}

type VaultService struct {
	vaultItems repository.VaultItemRepository
	audit      *AuditService
}

func NewVaultService(vaultItems repository.VaultItemRepository, audit *AuditService) *VaultService {
	return &VaultService{vaultItems: vaultItems, audit: audit}
}

func (s *VaultService) CreateItem(ctx context.Context, userID uuid.UUID, audit AuditContext, input CreateVaultItemInput) (domain.VaultItem, error) {
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

	if s.audit != nil {
		itemID := item.ID
		s.audit.Log(ctx, userID, audit, AuditVaultItemCreate, "vault_item", &itemID, map[string]any{
			"item_type": item.ItemType,
			"title":     item.Title,
		})
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

func (s *VaultService) ListItems(ctx context.Context, userID uuid.UUID, filter ListVaultItemsFilter) (ListVaultItemsResult, error) {
	if filter.Limit <= 0 {
		filter.Limit = defaultVaultListLimit
	}
	if filter.Limit > maxVaultListLimit {
		filter.Limit = maxVaultListLimit
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}

	result, err := s.vaultItems.ListByUserID(ctx, userID, repository.ListVaultItemsFilter{
		Folder:   filter.Folder,
		ItemType: filter.ItemType,
		Tag:      filter.Tag,
		Title:    filter.Title,
		Limit:    filter.Limit,
		Offset:   filter.Offset,
	})
	if err != nil {
		return ListVaultItemsResult{}, fmt.Errorf("list vault items: %w", err)
	}

	return ListVaultItemsResult{
		Items:  result.Items,
		Total:  result.Total,
		Limit:  filter.Limit,
		Offset: filter.Offset,
	}, nil
}

type UpdateVaultItemInput struct {
	EncryptedData []byte
	ItemType      string
	Title         string
	Folder        string
	Tags          []string
	Version       int32
}

func (s *VaultService) UpdateItem(ctx context.Context, userID, itemID uuid.UUID, audit AuditContext, input UpdateVaultItemInput) (domain.VaultItem, error) {
	if _, err := s.GetItem(ctx, userID, itemID); err != nil {
		return domain.VaultItem{}, err
	}
	item, err := s.vaultItems.Update(ctx, domain.VaultItem{
		ID:            itemID,
		EncryptedData: input.EncryptedData,
		ItemType:      input.ItemType,
		Title:         input.Title,
		Folder:        input.Folder,
		Tags:          input.Tags,
		UpdatedAt:     time.Now(),
		Version:       input.Version,
	})
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("update vault item: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditVaultItemUpdate, "vault_item", &itemID, map[string]any{
			"item_type": item.ItemType,
			"title":     item.Title,
			"version":   item.Version,
		})
	}

	return item, nil
}

func (s *VaultService) DeleteItem(ctx context.Context, userID, itemID uuid.UUID, audit AuditContext, version int32) (domain.VaultItem, error) {
	if _, err := s.GetItem(ctx, userID, itemID); err != nil {
		return domain.VaultItem{}, err
	}
	item, err := s.vaultItems.Delete(ctx, itemID, version)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("delete vault item: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditVaultItemDelete, "vault_item", &itemID, map[string]any{
			"title":   item.Title,
			"version": item.Version,
		})
	}

	return item, nil
}

func (s *VaultService) RestoreItem(ctx context.Context, userID, itemID uuid.UUID, audit AuditContext, version int32) (domain.VaultItem, error) {
	item, err := s.vaultItems.Restore(ctx, itemID, version, userID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return domain.VaultItem{}, ErrNotFound
		}
		return domain.VaultItem{}, fmt.Errorf("restore vault item: %w", err)
	}

	if s.audit != nil {
		s.audit.Log(ctx, userID, audit, AuditVaultItemRestore, "vault_item", &itemID, map[string]any{
			"title":   item.Title,
			"version": item.Version,
		})
	}

	return item, nil
}
