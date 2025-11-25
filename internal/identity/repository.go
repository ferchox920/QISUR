package identity

import "context"

// UserRepository holds persistence contracts for users.
type UserRepository interface {
	CreateUser(ctx context.Context, user User) (User, error)
	GetByEmail(ctx context.Context, email string) (User, error)
	GetByID(ctx context.Context, id UserID) (User, error)
	SetVerification(ctx context.Context, userID UserID, verified bool) error
	UpdateStatus(ctx context.Context, userID UserID, status UserStatus) error
	UpdateUserProfile(ctx context.Context, user User) (User, error)
}

// RoleRepository holds contracts for role management.
type RoleRepository interface {
	EnsureRole(ctx context.Context, role RoleName) error
	AssignRole(ctx context.Context, userID UserID, role RoleName) error
}
