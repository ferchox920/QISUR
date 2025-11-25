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
	VerificationCodeVerifier VerificationCodeVerifier
	TokenProvider            TokenProvider
}

type service struct {
	deps ServiceDeps
}

// NewService constructs the identity service with injected dependencies.
func NewService(deps ServiceDeps) Service {
	return &service{deps: deps}
}

func (s *service) register(ctx context.Context, input RegisterUserInput, role RoleName) (User, error) {
	if s.deps.PasswordHasher == nil || s.deps.RoleRepo == nil {
		return User{}, ErrNotImplemented
	}
	if input.Email == "" || input.Password == "" || input.FullName == "" {
		return User{}, ErrInvalidCredentials
	}
	if _, err := s.deps.UserRepo.GetByEmail(ctx, input.Email); err == nil {
		return User{}, ErrEmailAlreadyRegistered
	}
	hashed, err := s.deps.PasswordHasher.Hash(input.Password)
	if err != nil {
		return User{}, err
	}
	if err := s.deps.RoleRepo.EnsureRole(ctx, role); err != nil {
		return User{}, err
	}
	user := User{
		Email:        input.Email,
		FullName:     input.FullName,
		PasswordHash: hashed,
		Role:         role,
		Status:       UserStatusPendingVerification,
		IsVerified:   false,
	}
	created, err := s.deps.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return User{}, err
	}
	if s.deps.VerificationCodeProvider != nil && s.deps.VerificationSender != nil {
		code, err := s.deps.VerificationCodeProvider.Generate(ctx, created.ID)
		if err == nil {
			_ = s.deps.VerificationSender.SendVerification(ctx, created.Email, code)
		}
	}
	return created, nil
}

func (s *service) seedAdmin(ctx context.Context, seed AdminSeedInput) error {
	if seed.Email == "" || seed.Password == "" {
		return nil
	}
	if s.deps.PasswordHasher == nil || s.deps.RoleRepo == nil {
		return ErrNotImplemented
	}
	_, err := s.deps.UserRepo.GetByEmail(ctx, seed.Email)
	if err == nil {
		return nil
	}
	hashed, err := s.deps.PasswordHasher.Hash(seed.Password)
	if err != nil {
		return err
	}
	if err := s.deps.RoleRepo.EnsureRole(ctx, RoleAdmin); err != nil {
		return err
	}
	_, err = s.deps.UserRepo.CreateUser(ctx, User{
		Email:        seed.Email,
		FullName:     seed.FullName,
		PasswordHash: hashed,
		Role:         RoleAdmin,
		Status:       UserStatusActive,
		IsVerified:   true,
	})
	return err
}

func (s *service) RegisterClient(ctx context.Context, input RegisterUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	return s.register(ctx, input, RoleClient)
}

func (s *service) RegisterStandardUser(ctx context.Context, input RegisterUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	return s.register(ctx, input, RoleUser)
}

func (s *service) VerifyUser(ctx context.Context, input VerifyUserInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	if s.deps.VerificationCodeVerifier == nil {
		return ErrNotImplemented
	}
	if input.Code == "" {
		return ErrInvalidVerificationCode
	}
	valid, err := s.deps.VerificationCodeVerifier.Verify(ctx, string(input.UserID), input.Code)
	if err != nil {
		return err
	}
	if !valid {
		return ErrInvalidVerificationCode
	}
	return s.deps.UserRepo.SetVerification(ctx, input.UserID, true)
}

func (s *service) BlockUser(ctx context.Context, input BlockUserInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	return s.deps.UserRepo.UpdateStatus(ctx, input.UserID, UserStatusBlocked)
}

func (s *service) SeedAdmin(ctx context.Context, seed AdminSeedInput) error {
	if s.deps.UserRepo == nil {
		return ErrRepositoryNotConfigured
	}
	return s.seedAdmin(ctx, seed)
}

func (s *service) Login(ctx context.Context, input LoginInput) (AuthToken, error) {
	if s.deps.UserRepo == nil {
		return AuthToken{}, ErrRepositoryNotConfigured
	}
	if s.deps.PasswordHasher == nil || s.deps.TokenProvider == nil {
		return AuthToken{}, ErrNotImplemented
	}
	user, err := s.deps.UserRepo.GetByEmail(ctx, input.Email)
	if err != nil {
		return AuthToken{}, err
	}
	if user.Status == UserStatusBlocked {
		return AuthToken{}, ErrUserBlocked
	}
	if !user.IsVerified {
		return AuthToken{}, ErrUserNotVerified
	}
	if err := s.deps.PasswordHasher.Compare(user.PasswordHash, input.Password); err != nil {
		return AuthToken{}, ErrInvalidCredentials
	}
	token, err := s.deps.TokenProvider.Generate(ctx, user)
	if err != nil {
		return AuthToken{}, err
	}
	return AuthToken{Token: token}, nil
}

func (s *service) UpdateUser(ctx context.Context, input UpdateUserInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	user, err := s.deps.UserRepo.GetByID(ctx, input.UserID)
	if err != nil {
		return User{}, err
	}
	if input.FullName != "" {
		user.FullName = input.FullName
	}
	return s.deps.UserRepo.UpdateUserProfile(ctx, user)
}

func (s *service) UpdateUserRole(ctx context.Context, input UpdateUserRoleInput) (User, error) {
	if s.deps.UserRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	if s.deps.RoleRepo == nil {
		return User{}, ErrRepositoryNotConfigured
	}
	if err := s.deps.RoleRepo.EnsureRole(ctx, input.Role); err != nil {
		return User{}, err
	}
	if err := s.deps.RoleRepo.AssignRole(ctx, input.UserID, input.Role); err != nil {
		return User{}, err
	}
	return s.deps.UserRepo.GetByID(ctx, input.UserID)
}
