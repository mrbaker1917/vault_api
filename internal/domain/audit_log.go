package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type AuditLog struct {
	ID           uuid.UUID
	UserID       *uuid.UUID
	Action       string
	ResourceType string
	ResourceID   *uuid.UUID
	IPAddress    string
	UserAgent    string
	Metadata     json.RawMessage
	CreatedAt    time.Time
}
