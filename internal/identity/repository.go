package identity

import (
	"context"
	"time"
)

// UserRepository holds persistence contracts for users.
type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id UserID) (User, error)
	SetVerification(ctx context.Context, userID UserID, verified bool) error
	UpdateStatus(ctx context.Context, userID UserID, status UserStatus) error
	UpdateUserProfile(ctx context.Context, user User) (User, error)
	DeleteUser(ctx context.Context, userID UserID) error
	SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error
	GetVerificationCode(ctx context.Context, userID UserID) (code string, expiresAt time.Time, err error)
	DeleteVerificationCode(ctx context.Context, userID UserID) error
}

// RoleRepository holds contracts for role management.
type RoleRepository interface {
	EnsureRole(ctx context.Context, role RoleName) error
	AssignRole(ctx context.Context, userID UserID, role RoleName) error
}
