package identity

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubUserRepo struct{}

func (stubUserRepo) CreateUser(ctx context.Context, user User) (User, error) { return user, nil }
func (stubUserRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	return User{}, ErrUserNotFound
}
func (stubUserRepo) GetByID(ctx context.Context, id UserID) (User, error) {
	return User{ID: id, Email: string(id)}, nil
}
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
func (stubUserRepo) SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error {
	return nil
}
func (stubUserRepo) GetVerificationCode(ctx context.Context, userID UserID) (string, time.Time, error) {
	return "123456", time.Now().Add(time.Hour), nil
}
func (stubUserRepo) DeleteVerificationCode(ctx context.Context, userID UserID) error { return nil }

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

type codeRepo struct {
	stubUserRepo
	code      string
	expires   time.Time
	deleted   bool
	email     string
	saved     bool
	savedCode string
}

func (c *codeRepo) GetVerificationCode(ctx context.Context, userID UserID) (string, time.Time, error) {
	return c.code, c.expires, nil
}

func (c *codeRepo) DeleteVerificationCode(ctx context.Context, userID UserID) error {
	c.deleted = true
	return nil
}

func (c *codeRepo) GetByID(ctx context.Context, id UserID) (User, error) {
	return User{ID: id, Email: c.email}, nil
}

func (c *codeRepo) SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error {
	c.saved = true
	c.savedCode = code
	return nil
}

func TestVerifyUser_InvalidCode(t *testing.T) {
	repo := &codeRepo{code: "123456", expires: time.Now().Add(time.Hour)}
	svc := NewService(ServiceDeps{
		UserRepo: repo,
	})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id", Code: "000000"}); err != ErrInvalidVerificationCode {
		t.Fatalf("expected ErrInvalidVerificationCode, got %v", err)
	}
}

func TestVerifyUser_ExpiredCode(t *testing.T) {
	repo := &codeRepo{code: "123456", expires: time.Now().Add(-time.Minute), email: "a@b.c"}
	sender := &trackingSender{}
	newCode := "999888"
	svc := NewService(ServiceDeps{
		UserRepo:                 repo,
		VerificationCodeProvider: fixedCodeProvider{code: newCode},
		VerificationSender:       sender,
	})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id", Code: "123456"}); err != ErrInvalidVerificationCode {
		t.Fatalf("expected ErrInvalidVerificationCode for expired code, got %v", err)
	}
	if !repo.deleted {
		t.Fatalf("expected DeleteVerificationCode to be called on expiration")
	}
	if !repo.saved || repo.savedCode != newCode {
		t.Fatalf("expected new code to be saved on expiration, saved=%v code=%s", repo.saved, repo.savedCode)
	}
	if sender.sentCode != newCode || sender.sentTo != "a@b.c" {
		t.Fatalf("expected resend of new code, got to=%s code=%s", sender.sentTo, sender.sentCode)
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
	saved   bool
	code    string
}

func (t *trackingRepo) DeleteUser(ctx context.Context, userID UserID) error {
	t.deleted = true
	return nil
}

func (t *trackingRepo) SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error {
	t.saved = true
	t.code = code
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
	if !repo.saved || repo.code != code {
		t.Fatalf("expected verification code to be saved, got saved=%v code=%s", repo.saved, repo.code)
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
