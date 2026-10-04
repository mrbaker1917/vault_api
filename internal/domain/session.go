package domain

import (
	"time"

	"github.com/google/uuid"
)

type Session struct {
	ID                uuid.UUID
	UserID            uuid.UUID
	TokenHash         string
	PreviousTokenHash string
	TokenRotatedAt    *time.Time
	CreatedAt         time.Time
	ExpiresAt         time.Time
	RevokedAt         *time.Time // nil until revoked
	DeviceName        string
	IPAddress         string
	UserAgent         string
}