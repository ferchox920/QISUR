package identity

import "context"

// Service exposes identity use cases.
type Service interface {
	RegisterClient(ctx context.Context, input RegisterUserInput) (User, error)
	RegisterStandardUser(ctx context.Context, input RegisterUserInput) (User, error)
	VerifyUser(ctx context.Context, input VerifyUserInput) error
	BlockUser(ctx context.Context, input BlockUserInput) error
	Login(ctx context.Context, input LoginInput) (AuthToken, error)
	SeedAdmin(ctx context.Context, seed AdminSeedInput) error
	UpdateUser(ctx context.Context, input UpdateUserInput) (User, error)
	UpdateUserRole(ctx context.Context, input UpdateUserRoleInput) (User, error)
}

// RegisterUserInput encapsulates signup data.
type RegisterUserInput struct {
	Email    string
	Password string
	FullName string
}

// VerifyUserInput contains verification challenge data.
type VerifyUserInput struct {
	UserID UserID
	Code   string
}

// BlockUserInput captures admin-driven blocks.
type BlockUserInput struct {
	AdminID string
	UserID  UserID
	Reason  string
}

// LoginInput holds credentials for authentication.
type LoginInput struct {
	Email    string
	Password string
}

// AuthToken captures authentication token response.
type AuthToken struct {
	Token string
}

// UpdateUserInput contains updatable fields and actor info.
type UpdateUserInput struct {
	UserID    UserID
	UpdaterID string
	FullName  string
}

// UpdateUserRoleInput restricts role change to admin operations.
type UpdateUserRoleInput struct {
	AdminID string
	UserID  UserID
	Role    RoleName
}

// AdminSeedInput is used to pre-create an admin from configuration.
type AdminSeedInput struct {
	Email    string
	Password string
	FullName string
}

// ServiceDeps wires infrastructure to the domain service.
type ServiceDeps struct {
	UserRepo                 UserRepository
	RoleRepo                 RoleRepository
	PasswordHasher           PasswordHasher
	VerificationSender       VerificationSender
	VerificationCodeProvider VerificationCodeGenerator
	TokenProvider            TokenProvider
}

type service struct {
	deps ServiceDeps
}

// NewService constructs the identity service with injected dependencies.
func NewService(deps ServiceDeps) Service {
	return &service{deps: deps}
}

func (s *service) RegisterClient(ctx context.Context, input RegisterUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	// TODO: implement client registration flow (hash password, assign role, enqueue verification).
	return User{}, ErrNotImplemented
}

func (s *service) RegisterStandardUser(ctx context.Context, input RegisterUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	// TODO: implement standard user registration flow with role assignment and verification.
	return User{}, ErrNotImplemented
}

func (s *service) VerifyUser(ctx context.Context, input VerifyUserInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	// TODO: validate OTP and mark user as verified.
	return ErrNotImplemented
}

func (s *service) BlockUser(ctx context.Context, input BlockUserInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	// TODO: ensure admin permissions and block target user.
	return ErrNotImplemented
}

func (s *service) SeedAdmin(ctx context.Context, seed AdminSeedInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	// TODO: upsert admin user with admin role using seed values.
	return ErrNotImplemented
}

func (s *service) Login(ctx context.Context, input LoginInput) (AuthToken, error) {
	if s.deps.UserRepo == nil {
		return AuthToken{}, ErrRepositoryNotConfigured
	}
	if s.deps.PasswordHasher == nil || s.deps.TokenProvider == nil {
		return AuthToken{}, ErrNotImplemented
	}
	// TODO: fetch user, validate password, status, verification, block checks, then issue token.
	return AuthToken{}, ErrNotImplemented
}

func (s *service) UpdateUser(ctx context.Context, input UpdateUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	// TODO: authorize updater, then apply updates to user fields (no role change here).
	return User{}, ErrNotImplemented
}

func (s *service) UpdateUserRole(ctx context.Context, input UpdateUserRoleInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	// TODO: ensure admin permissions, update role via RoleRepo/UserRepo.
	return User{}, ErrNotImplemented
}
