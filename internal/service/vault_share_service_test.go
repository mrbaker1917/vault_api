package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

type stubSharedVaultItemRepo struct {
	shares map[uuid.UUID]domain.SharedVaultItem
}

func newStubSharedVaultItemRepo() *stubSharedVaultItemRepo {
	return &stubSharedVaultItemRepo{shares: make(map[uuid.UUID]domain.SharedVaultItem)}
}

func (s *stubSharedVaultItemRepo) Create(_ context.Context, share domain.SharedVaultItem) (domain.SharedVaultItem, error) {
	if share.ID == uuid.Nil {
		share.ID = uuid.New()
	}
	s.shares[share.ID] = share
	return share, nil
}

func (s *stubSharedVaultItemRepo) GetByVaultAndRecipient(_ context.Context, vaultItemID, recipientID uuid.UUID) (domain.SharedVaultItem, error) {
	for _, share := range s.shares {
		if share.VaultItemID == vaultItemID && share.SharedWithUserID == recipientID {
			return share, nil
		}
	}
	return domain.SharedVaultItem{}, repository.ErrNotFound
}

func (s *stubSharedVaultItemRepo) Delete(_ context.Context, vaultItemID, ownerID, recipientID uuid.UUID) error {
	for id, share := range s.shares {
		if share.VaultItemID == vaultItemID && share.OwnerID == ownerID && share.SharedWithUserID == recipientID {
			delete(s.shares, id)
			return nil
		}
	}
	return repository.ErrNotFound
}

func (s *stubSharedVaultItemRepo) ListForRecipient(_ context.Context, recipientID uuid.UUID, limit, offset int32) (repository.ListSharedVaultItemsResult, error) {
	items := make([]domain.SharedVaultItemWithItem, 0)
	for _, share := range s.shares {
		if share.SharedWithUserID != recipientID {
			continue
		}
		items = append(items, domain.SharedVaultItemWithItem{Share: share})
	}
	total := int64(len(items))
	if offset >= int32(len(items)) {
		return repository.ListSharedVaultItemsResult{Items: []domain.SharedVaultItemWithItem{}, Total: total}, nil
	}
	end := int(offset + limit)
	if end > len(items) {
		end = len(items)
	}
	return repository.ListSharedVaultItemsResult{
		Items: items[offset:end],
		Total: total,
	}, nil
}

type stubUserRepo struct {
	users map[uuid.UUID]domain.User
}

func newStubUserRepo() *stubUserRepo {
	return &stubUserRepo{users: make(map[uuid.UUID]domain.User)}
}

func (s *stubUserRepo) Create(_ context.Context, user domain.User) (domain.User, error) {
	if user.ID == uuid.Nil {
		user.ID = uuid.New()
	}
	s.users[user.ID] = user
	return user, nil
}

func (s *stubUserRepo) GetByEmail(_ context.Context, email string) (domain.User, error) {
	for _, user := range s.users {
		if user.Email == email {
			return user, nil
		}
	}
	return domain.User{}, repository.ErrNotFound
}

func (s *stubUserRepo) GetByID(_ context.Context, id uuid.UUID) (domain.User, error) {
	user, ok := s.users[id]
	if !ok {
		return domain.User{}, repository.ErrNotFound
	}
	return user, nil
}

func (s *stubUserRepo) EnableMFASecret(_ context.Context, _ uuid.UUID, _ string) error { return nil }
func (s *stubUserRepo) ConfirmMFA(_ context.Context, _ uuid.UUID) error                { return nil }
func (s *stubUserRepo) DisableMFA(_ context.Context, _ uuid.UUID) error                { return nil }

func TestVaultServiceShareItem(t *testing.T) {
	vaultRepo := newStubVaultItemRepo()
	shareRepo := newStubSharedVaultItemRepo()
	userRepo := newStubUserRepo()
	svc := NewVaultService(vaultRepo, shareRepo, userRepo, nil)

	ownerID := uuid.New()
	recipientID := uuid.New()
	userRepo.users[ownerID] = domain.User{ID: ownerID, Email: "owner@example.com"}
	userRepo.users[recipientID] = domain.User{ID: recipientID, Email: "recipient@example.com"}

	item, err := svc.CreateItem(context.Background(), ownerID, AuditContext{}, CreateVaultItemInput{
		EncryptedData: validEncryptedBlob(),
		ItemType:      "login",
		Title:         "Shared Secret",
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	share, err := svc.ShareItem(context.Background(), ownerID, item.ID, AuditContext{}, ShareVaultItemInput{
		Email:            "recipient@example.com",
		EncryptedItemKey: []byte("wrapped-key"),
		Permission:       domain.SharePermissionRead,
	})
	if err != nil {
		t.Fatalf("share item: %v", err)
	}
	if share.SharedWithUserID != recipientID {
		t.Fatalf("expected recipient %s, got %s", recipientID, share.SharedWithUserID)
	}

	_, err = svc.ShareItem(context.Background(), ownerID, item.ID, AuditContext{}, ShareVaultItemInput{
		Email:            "recipient@example.com",
		EncryptedItemKey: []byte("wrapped-key"),
		Permission:       domain.SharePermissionRead,
	})
	if !errors.Is(err, ErrAlreadyShared) {
		t.Fatalf("expected ErrAlreadyShared, got %v", err)
	}

	_, err = svc.ShareItem(context.Background(), ownerID, item.ID, AuditContext{}, ShareVaultItemInput{
		Email:            "owner@example.com",
		EncryptedItemKey: []byte("wrapped-key"),
		Permission:       domain.SharePermissionRead,
	})
	if !errors.Is(err, ErrCannotShareWithSelf) {
		t.Fatalf("expected ErrCannotShareWithSelf, got %v", err)
	}

	result, err := svc.ListSharedWithMe(context.Background(), recipientID, 50, 0)
	if err != nil {
		t.Fatalf("list shared items: %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected total 1, got %d", result.Total)
	}

	if err := svc.RevokeShare(context.Background(), ownerID, item.ID, recipientID, AuditContext{}); err != nil {
		t.Fatalf("revoke share: %v", err)
	}

	result, err = svc.ListSharedWithMe(context.Background(), recipientID, 50, 0)
	if err != nil {
		t.Fatalf("list shared items after revoke: %v", err)
	}
	if result.Total != 0 {
		t.Fatalf("expected total 0 after revoke, got %d", result.Total)
	}
}

func TestVaultServiceShareItemRejectsInvalidPermission(t *testing.T) {
	vaultRepo := newStubVaultItemRepo()
	shareRepo := newStubSharedVaultItemRepo()
	userRepo := newStubUserRepo()
	svc := NewVaultService(vaultRepo, shareRepo, userRepo, nil)

	ownerID := uuid.New()
	userRepo.users[ownerID] = domain.User{ID: ownerID, Email: "owner@example.com"}

	item, err := svc.CreateItem(context.Background(), ownerID, AuditContext{}, CreateVaultItemInput{
		EncryptedData: validEncryptedBlob(),
		ItemType:      "login",
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	_, err = svc.ShareItem(context.Background(), ownerID, item.ID, AuditContext{}, ShareVaultItemInput{
		Email:            "missing@example.com",
		EncryptedItemKey: []byte("wrapped-key"),
		Permission:       "admin",
	})
	if !errors.Is(err, ErrInvalidSharePermission) {
		t.Fatalf("expected ErrInvalidSharePermission, got %v", err)
	}
}

func TestVaultServiceShareItemRejectsEmptyKey(t *testing.T) {
	svc := NewVaultService(newStubVaultItemRepo(), newStubSharedVaultItemRepo(), newStubUserRepo(), nil)
	ownerID := uuid.New()

	_, err := svc.ShareItem(context.Background(), ownerID, uuid.New(), AuditContext{}, ShareVaultItemInput{
		Email:            "user@example.com",
		EncryptedItemKey: nil,
		Permission:       domain.SharePermissionRead,
	})
	if !errors.Is(err, ErrInvalidEncryptedItemKey) {
		t.Fatalf("expected ErrInvalidEncryptedItemKey, got %v", err)
	}
}

func TestVaultServiceRevokeShareNotFound(t *testing.T) {
	vaultRepo := newStubVaultItemRepo()
	svc := NewVaultService(vaultRepo, newStubSharedVaultItemRepo(), newStubUserRepo(), nil)
	ownerID := uuid.New()

	item, err := vaultRepo.Create(context.Background(), domain.VaultItem{
		UserID:        ownerID,
		EncryptedData: validEncryptedBlob(),
		ItemType:      "login",
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		Version:       1,
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}

	err = svc.RevokeShare(context.Background(), ownerID, item.ID, uuid.New(), AuditContext{})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}
