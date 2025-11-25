package identity

import (
	"context"
	"errors"
	"testing"
)

type stubUserRepo struct{}

func (stubUserRepo) CreateUser(ctx context.Context, user User) (User, error) { return User{}, nil }
func (stubUserRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	return User{}, nil
}
func (stubUserRepo) GetByID(ctx context.Context, id UserID) (User, error) { return User{}, nil }
func (stubUserRepo) SetVerification(ctx context.Context, userID UserID, verified bool) error {
	return nil
}
func (stubUserRepo) UpdateStatus(ctx context.Context, userID UserID, status UserStatus) error {
	return nil
}
func (stubUserRepo) UpdateUserProfile(ctx context.Context, user User) (User, error) {
	return user, nil
}
func (stubUserRepo) EnsureRole(ctx context.Context, role RoleName) error                { return nil }
func (stubUserRepo) AssignRole(ctx context.Context, userID UserID, role RoleName) error { return nil }
func (stubUserRepo) DeleteUser(ctx context.Context, userID UserID) error                { return nil }

type stubVerifier struct {
	valid bool
	err   error
}

func (s stubVerifier) Verify(ctx context.Context, userID string, code string) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	return s.valid, nil
}

func TestRegisterClient_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestRegisterStandardUser_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if _, err := svc.RegisterStandardUser(context.Background(), RegisterUserInput{}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestLogin_NotImplementedWhenHasherOrTokenMissing(t *testing.T) {
	svc := NewService(ServiceDeps{
		UserRepo: stubUserRepo{},
		RoleRepo: stubUserRepo{},
	})
	if _, err := svc.Login(context.Background(), LoginInput{Email: "a", Password: "b"}); err != ErrNotImplemented {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestVerifyUser_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id"}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestVerifyUser_VerifierRequired(t *testing.T) {
	svc := NewService(ServiceDeps{UserRepo: stubUserRepo{}})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id", Code: "code"}); err != ErrNotImplemented {
		t.Fatalf("expected ErrNotImplemented, got %v", err)
	}
}

func TestVerifyUser_InvalidCode(t *testing.T) {
	svc := NewService(ServiceDeps{
		UserRepo:                 stubUserRepo{},
		VerificationCodeVerifier: stubVerifier{valid: false},
	})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id", Code: "bad"}); err != ErrInvalidVerificationCode {
		t.Fatalf("expected ErrInvalidVerificationCode, got %v", err)
	}
}

type failingSender struct{}

func (failingSender) SendVerification(ctx context.Context, email, code string) error {
	return errors.New("send failed")
}

type trackingSender struct {
	sentTo   string
	sentCode string
}

func (t *trackingSender) SendVerification(ctx context.Context, email, code string) error {
	t.sentTo = email
	t.sentCode = code
	return nil
}

type trackingRepo struct {
	stubUserRepo
	deleted bool
}

func (t *trackingRepo) DeleteUser(ctx context.Context, userID UserID) error {
	t.deleted = true
	return nil
}

type fixedCodeProvider struct {
	code string
	err  error
}

func (p fixedCodeProvider) Generate(ctx context.Context, userID string) (string, error) {
	return p.code, p.err
}

func TestRegister_RollsBackOnSendFailure(t *testing.T) {
	repo := &trackingRepo{}
	svc := NewService(ServiceDeps{
		UserRepo:                 repo,
		RoleRepo:                 repo,
		PasswordHasher:           stubHasher{},
		VerificationCodeProvider: fixedCodeProvider{code: "123456"},
		VerificationSender:       failingSender{},
	})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "secret123", FullName: "Test"}); err == nil {
		t.Fatalf("expected error due to send failure")
	}
	if !repo.deleted {
		t.Fatalf("expected DeleteUser to be called on rollback")
	}
}

func TestRegister_UsesVerificationSenderWhenConfigured(t *testing.T) {
	repo := &trackingRepo{}
	sender := &trackingSender{}
	code := "654321"
	svc := NewService(ServiceDeps{
		UserRepo:                 repo,
		RoleRepo:                 repo,
		PasswordHasher:           stubHasher{},
		VerificationCodeProvider: fixedCodeProvider{code: code},
		VerificationSender:       sender,
	})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "secret123", FullName: "Test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.sentTo != "a@b.c" || sender.sentCode != code {
		t.Fatalf("sender not invoked as expected: %+v", sender)
	}
}

func TestBlockUser_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if err := svc.BlockUser(context.Background(), BlockUserInput{UserID: "id"}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestUpdateUser_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if _, err := svc.UpdateUser(context.Background(), UpdateUserInput{UserID: "id"}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

func TestUpdateUserRole_RepoRequired(t *testing.T) {
	svc := NewService(ServiceDeps{})
	if _, err := svc.UpdateUserRole(context.Background(), UpdateUserRoleInput{UserID: "id"}); err != ErrRepositoryNotConfigured {
		t.Fatalf("expected ErrRepositoryNotConfigured, got %v", err)
	}
}

type stubHasher struct{}

func (stubHasher) Hash(password string) (string, error) { return "hashed", nil }
func (stubHasher) Compare(hash, password string) error  { return nil }
