package repository

import (
	"fmt"
	"net/netip"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

func pgBoolToPG(v bool) pgtype.Bool {
	return pgtype.Bool{Bool: v, Valid: true}
}

func pgBoolFromPG(b pgtype.Bool) bool {
	if !b.Valid {
		return false
	}
	return b.Bool
}

func pgTextFromPtr(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgTextFromString(s string) pgtype.Text {
	if s == "" {
		return pgtype.Text{}
	}
	return pgtype.Text{String: s, Valid: true}
}

func pgTextToPtr(t pgtype.Text) *string {
	if !t.Valid {
		return nil
	}
	return &t.String
}

func pgTextToString(t pgtype.Text) string {
	if !t.Valid {
		return ""
	}
	return t.String
}

func pgTimestampToPG(t time.Time) pgtype.Timestamp {
	return pgtype.Timestamp{Time: t, Valid: true}
}

func pgTimestampFromPG(t pgtype.Timestamp) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}

func pgTimestampToPtr(t pgtype.Timestamp) *time.Time {
	if !t.Valid {
		return nil
	}
	return &t.Time
}

func pgUUIDToPG(id uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: id, Valid: true}
}

func uuidFromPG(u pgtype.UUID) (uuid.UUID, error) {
	if !u.Valid {
		return uuid.Nil, fmt.Errorf("null uuid")
	}
	return uuid.FromBytes(u.Bytes[:])
}

func netipAddrFromString(s string) (*netip.Addr, error) {
	if s == "" {
		return nil, nil
	}
	addr, err := netip.ParseAddr(s)
	if err != nil {
		return nil, err
	}
	return &addr, nil
}

func netipAddrToString(a *netip.Addr) string {
	if a == nil {
		return ""
	}
	return a.String()
}

func pgInt4ToPG(v int32) pgtype.Int4 {
	return pgtype.Int4{Int32: v, Valid: true}
}

func pgInt4FromPG(v pgtype.Int4) int32 {
	if !v.Valid {
		return 0
	}
	return v.Int32
}
