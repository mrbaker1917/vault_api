package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash     string
	CreatedAt     time.Time
	ExpiresAt     time.Time
	RevokedAt     *time.Time // nil until revoked
	DeviceName    string
	IPAddress     string
	UserAgent     string
}