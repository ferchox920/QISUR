package identity

import "time"

// UserStatus captures lifecycle states for accounts.
type UserStatus string

const (
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusBlocked             UserStatus = "blocked"
)

// User aggregates identity data without leaking transport or storage concerns.
type User struct {
	ID           string
	Email        string
	FullName     string
	PasswordHash string
	Role         RoleName
	Status       UserStatus
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
