package service

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"vault_api/internal/crypto"
	"vault_api/internal/domain"
	"vault_api/internal/repository"
)

func validEncryptedBlob(extra ...byte) []byte {
	data := []byte{crypto.EncryptedBlobVersion1, 0xAA, 0xBB}
	return append(data, extra...)
}

type stubVaultItemRepo struct {
	items map[uuid.UUID]domain.VaultItem
}

func newStubVaultItemRepo() *stubVaultItemRepo {
	return &stubVaultItemRepo{items: make(map[uuid.UUID]domain.VaultItem)}
}

func (s *stubVaultItemRepo) Create(_ context.Context, item domain.VaultItem) (domain.VaultItem, error) {
	if item.ID == uuid.Nil {
		item.ID = uuid.New()
	}
	s.items[item.ID] = item
	return item, nil
}

func (s *stubVaultItemRepo) GetByID(_ context.Context, id uuid.UUID) (domain.VaultItem, error) {
	item, ok := s.items[id]
	if !ok || item.DeletedAt != nil {
		return domain.VaultItem{}, repository.ErrNotFound
	}
	return item, nil
}

func (s *stubVaultItemRepo) ListByUserID(_ context.Context, userID uuid.UUID, filter repository.ListVaultItemsFilter) (repository.ListVaultItemsResult, error) {
	items := make([]domain.VaultItem, 0)
	for _, item := range s.items {
		if item.UserID != userID || item.DeletedAt != nil {
			continue
		}
		items = append(items, item)
	}
	total := int64(len(items))
	if filter.Offset >= int32(len(items)) {
		return repository.ListVaultItemsResult{Items: []domain.VaultItem{}, Total: total}, nil
	}
	end := int(filter.Offset + filter.Limit)
	if end > len(items) {
		end = len(items)
	}
	return repository.ListVaultItemsResult{
		Items: items[filter.Offset:end],
		Total: total,
	}, nil
}

func (s *stubVaultItemRepo) Update(_ context.Context, item domain.VaultItem) (domain.VaultItem, error) {
	existing, ok := s.items[item.ID]
	if !ok || existing.DeletedAt != nil || existing.Version != item.Version {
		return domain.VaultItem{}, repository.ErrNotFound
	}
	item.UserID = existing.UserID
	item.CreatedAt = existing.CreatedAt
	item.Version = existing.Version + 1
	s.items[item.ID] = item
	return item, nil
}

func (s *stubVaultItemRepo) Delete(_ context.Context, id uuid.UUID, version int32) (domain.VaultItem, error) {
	item, ok := s.items[id]
	if !ok || item.DeletedAt != nil || item.Version != version {
		return domain.VaultItem{}, repository.ErrNotFound
	}
	now := time.Now()
	item.DeletedAt = &now
	item.Version++
	s.items[id] = item
	return item, nil
}

func (s *stubVaultItemRepo) Restore(_ context.Context, id uuid.UUID, version int32, userID uuid.UUID) (domain.VaultItem, error) {
	item, ok := s.items[id]
	if !ok || item.UserID != userID || item.DeletedAt == nil || item.Version != version {
		return domain.VaultItem{}, repository.ErrNotFound
	}
	item.DeletedAt = nil
	item.Version++
	s.items[id] = item
	return item, nil
}

func TestVaultServiceCreateItemValidatesEncryptedBlob(t *testing.T) {
	svc := NewVaultService(newStubVaultItemRepo(), nil)
	userID := uuid.New()

	_, err := svc.CreateItem(context.Background(), userID, AuditContext{}, CreateVaultItemInput{
		EncryptedData: []byte{0x02, 0x01},
		ItemType:      "login",
	})
	if !errors.Is(err, ErrInvalidEncryptedBlob) {
		t.Fatalf("expected ErrInvalidEncryptedBlob, got %v", err)
	}
}

func TestVaultServiceCreateGetUpdateDeleteRestore(t *testing.T) {
	repo := newStubVaultItemRepo()
	svc := NewVaultService(repo, nil)
	userID := uuid.New()
	otherUserID := uuid.New()

	created, err := svc.CreateItem(context.Background(), userID, AuditContext{}, CreateVaultItemInput{
		EncryptedData: validEncryptedBlob(),
		ItemType:      "login",
		Title:         "GitHub",
	})
	if err != nil {
		t.Fatalf("create item: %v", err)
	}
	if created.Version != 1 {
		t.Fatalf("expected version 1, got %d", created.Version)
	}

	got, err := svc.GetItem(context.Background(), userID, created.ID)
	if err != nil {
		t.Fatalf("get item: %v", err)
	}
	if got.Title != "GitHub" {
		t.Fatalf("expected title GitHub, got %q", got.Title)
	}

	_, err = svc.GetItem(context.Background(), otherUserID, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for other user, got %v", err)
	}

	updated, err := svc.UpdateItem(context.Background(), userID, created.ID, AuditContext{}, UpdateVaultItemInput{
		EncryptedData: validEncryptedBlob(0xCC),
		ItemType:      "login",
		Title:         "GitHub Updated",
		Version:       1,
	})
	if err != nil {
		t.Fatalf("update item: %v", err)
	}
	if updated.Version != 2 {
		t.Fatalf("expected version 2, got %d", updated.Version)
	}

	_, err = svc.UpdateItem(context.Background(), userID, created.ID, AuditContext{}, UpdateVaultItemInput{
		EncryptedData: validEncryptedBlob(0xDD),
		ItemType:      "login",
		Title:         "Stale version",
		Version:       1,
	})
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound for stale version, got %v", err)
	}

	deleted, err := svc.DeleteItem(context.Background(), userID, created.ID, AuditContext{}, updated.Version)
	if err != nil {
		t.Fatalf("delete item: %v", err)
	}
	if deleted.DeletedAt == nil {
		t.Fatal("expected deleted_at to be set")
	}

	_, err = svc.GetItem(context.Background(), userID, created.ID)
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted item to be hidden, got %v", err)
	}

	restored, err := svc.RestoreItem(context.Background(), userID, created.ID, AuditContext{}, deleted.Version)
	if err != nil {
		t.Fatalf("restore item: %v", err)
	}
	if restored.DeletedAt != nil {
		t.Fatal("expected deleted_at to be cleared after restore")
	}
	if restored.Version != deleted.Version+1 {
		t.Fatalf("expected version %d, got %d", deleted.Version+1, restored.Version)
	}
}

func TestVaultServiceListItemsAppliesDefaultLimit(t *testing.T) {
	repo := newStubVaultItemRepo()
	svc := NewVaultService(repo, nil)
	userID := uuid.New()

	for i := 0; i < 3; i++ {
		_, err := svc.CreateItem(context.Background(), userID, AuditContext{}, CreateVaultItemInput{
			EncryptedData: validEncryptedBlob(byte(i)),
			ItemType:      "login",
			Title:         "Item",
		})
		if err != nil {
			t.Fatalf("create item %d: %v", i, err)
		}
	}

	result, err := svc.ListItems(context.Background(), userID, ListVaultItemsFilter{})
	if err != nil {
		t.Fatalf("list items: %v", err)
	}
	if result.Limit != defaultVaultListLimit {
		t.Fatalf("expected default limit %d, got %d", defaultVaultListLimit, result.Limit)
	}
	if result.Total != 3 {
		t.Fatalf("expected total 3, got %d", result.Total)
	}
	if len(result.Items) != 3 {
		t.Fatalf("expected 3 items, got %d", len(result.Items))
	}
}
