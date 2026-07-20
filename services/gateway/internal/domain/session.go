package domain

import (
	"time"

	"github.com/google/uuid"
)

// Session is a human/dashboard authentication session (Domain B in services/gateway/DESIGN.md),
// stored server-side in Redis per ADR 0008 — the cookie carries only the SessionID, never this
// struct's contents.
type Session struct {
	SessionID  uuid.UUID
	UserID     uuid.UUID
	Role       Role
	CreatedAt  time.Time
	ExpiresAt  time.Time
	LastSeenAt time.Time
}
