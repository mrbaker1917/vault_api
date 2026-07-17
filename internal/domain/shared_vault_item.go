package domain

import (
	"time"

	"github.com/google/uuid"
)

const (
	SharePermissionRead  = "read"
	SharePermissionWrite = "write"
)

type SharedVaultItem struct {
	ID               uuid.UUID
	VaultItemID      uuid.UUID
	OwnerID          uuid.UUID
	SharedWithUserID uuid.UUID
	EncryptedItemKey []byte
	Permission       string
	CreatedAt        time.Time
}

type SharedVaultItemWithItem struct {
	Share SharedVaultItem
	Item  VaultItem
}
