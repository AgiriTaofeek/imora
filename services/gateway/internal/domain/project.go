package domain

import (
	"time"

	"github.com/google/uuid"
)

// Project is a customer's top-level container for browser-sdk data capture — the entity a
// ProjectKey authorizes writes on behalf of. See services/gateway/DESIGN.md.
type Project struct {
	ID        uuid.UUID
	Name      string
	OwnerID   uuid.UUID
	CreatedAt time.Time
	Active    bool
}

// ProjectKey authenticates browser-sdk write traffic for its Project (Domain A in
// services/gateway/DESIGN.md). A Project owns one-to-many ProjectKeys so a key can be rotated
// without a window where the project has zero working keys — see ADR 0009.
type ProjectKey struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	CreatedAt time.Time
	Active    bool
	RevokedAt *time.Time // nil until revoked
}
