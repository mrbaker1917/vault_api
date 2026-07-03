package domain

import (
	"time"

	"github.com/google/uuid"
)

type VaultItem struct {
	ID            uuid.UUID
	UserID        uuid.UUID
	EncryptedData []byte
	ItemType      string
	Title         string
	Folder        string
	Tags          []string
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
	Version       int32
}
