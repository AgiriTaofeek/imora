package domain

import (
	"time"

	"github.com/google/uuid"
)

// Role is a user's RBAC role, per research/07-security/README.md#authorization.
type Role string

const (
	RoleEngineer          Role = "engineer"
	RoleComplianceOfficer Role = "compliance_officer"
	RolePlatformOperator  Role = "platform_operator"
	RoleAdmin             Role = "admin"
)

// User is a human account authenticated via local auth (Argon2id + optional TOTP), per
// docs/prd.md's M0 scope and research/07-security/README.md#authentication. Matches the users
// table already sketched in docs/architecture.md's Postgres section.
type User struct {
	ID           uuid.UUID
	Email        string
	PasswordHash string // Argon2id (m=19456, t=2, p=1) — see docs/coding-standards.md
	Role         Role
	TOTPSecret   *string // nil until MFA enrollment; surfaced but optional per docs/user-stories.md Flow A
	CreatedAt    time.Time
}
