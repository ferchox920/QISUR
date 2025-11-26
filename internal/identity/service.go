package identity

import (
	"context"
	"errors"
	"time"
)

// Service expone casos de uso de identidad.
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

// RegisterUserInput encapsula datos de registro.
type RegisterUserInput struct {
	Email    string
	Password string
	FullName string
}

// VerifyUserInput contiene datos del desafio de verificacion.
type VerifyUserInput struct {
	UserID UserID
	Code   string
}

// BlockUserInput captura bloqueos realizados por admin.
type BlockUserInput struct {
	AdminID string
	UserID  UserID
	Reason  string
}

// LoginInput contiene credenciales para autenticacion.
type LoginInput struct {
	Email    string
	Password string
}

// AuthToken contiene el token emitido tras autenticacion.
type AuthToken struct {
	Token string
}

// UpdateUserInput contiene campos editables e info del actor.
type UpdateUserInput struct {
	UserID    UserID
	UpdaterID string
	FullName  string
}

// UpdateUserRoleInput restringe el cambio de rol a operaciones admin.
type UpdateUserRoleInput struct {
	AdminID string
	UserID  UserID
	Role    RoleName
}

// AdminSeedInput se usa para pre-crear un admin desde la configuracion.
type AdminSeedInput struct {
	Email    string
	Password string
	FullName string
}

// ServiceDeps cablea la infraestructura al servicio de dominio.
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

const dummyPasswordHash = "$2a$10$7EqJtq98hPqEX7fNZaFWoOHi4bxmC8lzQju0aDY9.6e2cqE8X4Fi."

// NewService construye el servicio de identidad con dependencias inyectadas.
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
	if s.deps.VerificationCodeProvider != nil && s.deps.VerificationSender != nil {
		tx, err := s.deps.UserRepo.BeginTx(ctx)
		if err != nil {
			return User{}, err
		}
		defer tx.Rollback(ctx)

		created, err := tx.CreateUser(ctx, user)
		if err != nil {
			return User{}, err
		}
		code, err := s.deps.VerificationCodeProvider.Generate(ctx, created.ID)
		if err != nil {
			return User{}, err
		}
		exp := time.Now().Add(15 * time.Minute)
		if err := tx.SaveVerificationCode(ctx, created.ID, code, exp); err != nil {
			return User{}, err
		}
		if err := s.deps.VerificationSender.SendVerification(ctx, created.Email, code); err != nil {
			return User{}, err
		}
		if err := tx.Commit(ctx); err != nil {
			return User{}, err
		}
		return created, nil
	}
	created, err := s.deps.UserRepo.CreateUser(ctx, user)
	if err != nil {
		return User{}, err
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
	if input.Code == "" {
		return ErrInvalidVerificationCode
	}
	code, expiresAt, err := s.deps.UserRepo.GetVerificationCode(ctx, input.UserID)
	if err != nil {
		return err
	}
	if time.Now().After(expiresAt) {
		_ = s.deps.UserRepo.DeleteVerificationCode(ctx, input.UserID)
		// reenviar un codigo nuevo
		if s.deps.VerificationCodeProvider != nil && s.deps.VerificationSender != nil {
			newCode, genErr := s.deps.VerificationCodeProvider.Generate(ctx, string(input.UserID))
			if genErr == nil {
				exp := time.Now().Add(15 * time.Minute)
				if saveErr := s.deps.UserRepo.SaveVerificationCode(ctx, input.UserID, newCode, exp); saveErr == nil {
					if user, userErr := s.deps.UserRepo.GetByID(ctx, input.UserID); userErr == nil {
						_ = s.deps.VerificationSender.SendVerification(ctx, user.Email, newCode)
					}
				}
			}
		}
		return ErrInvalidVerificationCode
	}
	if code != input.Code {
		return ErrInvalidVerificationCode
	}
	if err := s.deps.UserRepo.SetVerification(ctx, input.UserID, true); err != nil {
		return err
	}
	return s.deps.UserRepo.DeleteVerificationCode(ctx, input.UserID)
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
		s.consumePasswordHash(input.Password)
		if errors.Is(err, ErrUserNotFound) {
			return AuthToken{}, ErrInvalidCredentials
		}
		return AuthToken{}, err
	}
	if err := s.deps.PasswordHasher.Compare(user.PasswordHash, input.Password); err != nil {
		return AuthToken{}, ErrInvalidCredentials
	}
	if user.Status == UserStatusBlocked || !user.IsVerified {
		return AuthToken{}, ErrInvalidCredentials
	}
	token, err := s.deps.TokenProvider.Generate(ctx, user)
	if err != nil {
		return AuthToken{}, err
	}
	return AuthToken{Token: token}, nil
}

func (s *service) consumePasswordHash(password string) {
	if s.deps.PasswordHasher == nil {
		return
	}
	_ = s.deps.PasswordHasher.Compare(dummyPasswordHash, password)
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
