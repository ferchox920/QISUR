package identity

import "time"

// UserID representa un identificador de usuario; se espera UUID.
type UserID = string

// UserStatus describe estados de ciclo de vida para cuentas.
type UserStatus string

const (
	UserStatusPendingVerification UserStatus = "pending_verification"
	UserStatusActive              UserStatus = "active"
	UserStatusBlocked             UserStatus = "blocked"
)

// User agrupa datos de identidad sin mezclar detalles de transporte o storage.
type User struct {
	ID           UserID
	Email        string
	FullName     string
	PasswordHash string
	Role         RoleName
	Status       UserStatus
	IsVerified   bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
