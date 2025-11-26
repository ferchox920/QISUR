package identity

import (
	"context"
	"errors"
	"testing"
	"time"
)

type stubUserRepo struct{}

type stubTx struct{}

func (stubUserRepo) BeginTx(ctx context.Context) (UserTx, error)             { return &stubTx{}, nil }
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

func (stubTx) CreateUser(ctx context.Context, user User) (User, error) { return user, nil }
func (stubTx) SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error {
	return nil
}
func (stubTx) Commit(ctx context.Context) error   { return nil }
func (stubTx) Rollback(ctx context.Context) error { return nil }

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

func TestRegister_EnforcesPasswordPolicy(t *testing.T) {
	svc := NewService(ServiceDeps{
		UserRepo:       stubUserRepo{},
		RoleRepo:       stubUserRepo{},
		PasswordHasher: stubHasher{},
	})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "short", FullName: "Test"}); err == nil {
		t.Fatalf("expected error for short password")
	}
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "alllowercase1", FullName: "Test"}); err == nil {
		t.Fatalf("expected error for weak password")
	}
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "Strong123", FullName: "Test"}); err != nil {
		t.Fatalf("unexpected error for strong password: %v", err)
	}
}

type loginRepo struct {
	stubUserRepo
	user User
	err  error
}

func (l loginRepo) GetByEmail(ctx context.Context, email string) (User, error) {
	if l.err != nil {
		return User{}, l.err
	}
	return l.user, nil
}

func TestLogin_MasksUserNotFound(t *testing.T) {
	hasher := &trackingHasher{}
	svc := NewService(ServiceDeps{
		UserRepo:       loginRepo{err: ErrUserNotFound},
		PasswordHasher: hasher,
		TokenProvider:  stubTokenProvider{token: "tok"},
	})
	if _, err := svc.Login(context.Background(), LoginInput{Email: "missing@example.com", Password: "secret"}); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got %v", err)
	}
	if hasher.compareCount != 1 {
		t.Fatalf("expected hasher compare to be invoked for timing, got %d", hasher.compareCount)
	}
}

func TestLogin_BlockedReturnsGenericError(t *testing.T) {
	hasher := &trackingHasher{}
	svc := NewService(ServiceDeps{
		UserRepo: loginRepo{user: User{
			Email:        "blocked@example.com",
			PasswordHash: "hash",
			Status:       UserStatusBlocked,
			IsVerified:   true,
		}},
		PasswordHasher: hasher,
		TokenProvider:  stubTokenProvider{token: "tok"},
	})
	if _, err := svc.Login(context.Background(), LoginInput{Email: "blocked@example.com", Password: "secret"}); err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials for blocked user, got %v", err)
	}
	if hasher.compareCount != 1 {
		t.Fatalf("expected hasher compare to be invoked before status check, got %d", hasher.compareCount)
	}
}

func TestLogin_Success(t *testing.T) {
	hasher := &trackingHasher{}
	svc := NewService(ServiceDeps{
		UserRepo: loginRepo{user: User{
			ID:           "u1",
			Email:        "user@example.com",
			PasswordHash: "hash",
			Status:       UserStatusActive,
			IsVerified:   true,
		}},
		PasswordHasher: hasher,
		TokenProvider:  stubTokenProvider{token: "tok123"},
	})
	token, err := svc.Login(context.Background(), LoginInput{Email: "user@example.com", Password: "secret"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.Token != "tok123" {
		t.Fatalf("expected token tok123, got %s", token.Token)
	}
	if hasher.compareCount != 1 {
		t.Fatalf("expected hasher compare once, got %d", hasher.compareCount)
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
	svc := NewService(ServiceDeps{
		UserRepo: repo,
	})
	if err := svc.VerifyUser(context.Background(), VerifyUserInput{UserID: "id", Code: "123456"}); err != ErrInvalidVerificationCode {
		t.Fatalf("expected ErrInvalidVerificationCode for expired code, got %v", err)
	}
	if !repo.deleted {
		t.Fatalf("expected DeleteVerificationCode to be called on expiration")
	}
	if repo.saved {
		t.Fatalf("did not expect new code to be saved automatically")
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
	saved      bool
	code       string
	committed  bool
	rolledBack bool
	tx         *trackingTx
}

func (t *trackingRepo) BeginTx(ctx context.Context) (UserTx, error) {
	tx := &trackingTx{repo: t}
	t.tx = tx
	return tx, nil
}

type trackingTx struct {
	repo *trackingRepo
	done bool
}

func (t *trackingTx) CreateUser(ctx context.Context, user User) (User, error) {
	user.ID = "generated-id"
	return user, nil
}

func (t *trackingTx) SaveVerificationCode(ctx context.Context, userID UserID, code string, expiresAt time.Time) error {
	t.repo.saved = true
	t.repo.code = code
	return nil
}

func (t *trackingTx) Commit(ctx context.Context) error {
	if t.done {
		return nil
	}
	t.done = true
	t.repo.committed = true
	return nil
}

func (t *trackingTx) Rollback(ctx context.Context) error {
	if t.done {
		return nil
	}
	t.done = true
	t.repo.rolledBack = true
	return nil
}

type fixedCodeProvider struct {
	code string
	err  error
}

func (p fixedCodeProvider) Generate(ctx context.Context, userID string) (string, error) {
	return p.code, p.err
}

func TestRegister_SendFailureAfterCommitReturnsError(t *testing.T) {
	repo := &trackingRepo{}
	sender := &flakySender{failures: 2}
	svc := NewService(ServiceDeps{
		UserRepo:                 repo,
		RoleRepo:                 repo,
		PasswordHasher:           stubHasher{},
		VerificationCodeProvider: fixedCodeProvider{code: "123456"},
		VerificationSender:       sender,
	})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "Secret123", FullName: "Test"}); err == nil {
		t.Fatalf("expected error due to send failure")
	}
	if !repo.committed || repo.rolledBack {
		t.Fatalf("expected transaction committed before send, got committed=%v rolledBack=%v", repo.committed, repo.rolledBack)
	}
	if sender.sent != 2 {
		t.Fatalf("expected two send attempts, got %d", sender.sent)
	}
}

func TestRegister_SendRetriesOnceAndSucceeds(t *testing.T) {
	repo := &trackingRepo{}
	sender := &flakySender{failures: 1}
	code := "999000"
	svc := NewService(ServiceDeps{
		UserRepo:                 repo,
		RoleRepo:                 repo,
		PasswordHasher:           stubHasher{},
		VerificationCodeProvider: fixedCodeProvider{code: code},
		VerificationSender:       sender,
	})
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "Secret123", FullName: "Test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.sent != 2 {
		t.Fatalf("expected two send attempts (one retry), got %d", sender.sent)
	}
	if sender.sentCode != code || sender.sentTo != "a@b.c" {
		t.Fatalf("expected successful send to a@b.c with code %s, got to=%s code=%s", code, sender.sentTo, sender.sentCode)
	}
	if !repo.committed || repo.rolledBack {
		t.Fatalf("expected transaction committed, got committed=%v rolledBack=%v", repo.committed, repo.rolledBack)
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
	if _, err := svc.RegisterClient(context.Background(), RegisterUserInput{Email: "a@b.c", Password: "Secret123", FullName: "Test"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sender.sentTo != "a@b.c" || sender.sentCode != code {
		t.Fatalf("sender not invoked as expected: %+v", sender)
	}
	if !repo.saved || repo.code != code {
		t.Fatalf("expected verification code to be saved, got saved=%v code=%s", repo.saved, repo.code)
	}
	if !repo.committed || repo.rolledBack {
		t.Fatalf("expected transaction committed, got committed=%v rolledBack=%v", repo.committed, repo.rolledBack)
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

type trackingHasher struct {
	compareCount int
	compareErr   error
}

func (t *trackingHasher) Hash(password string) (string, error) { return "hash", nil }
func (t *trackingHasher) Compare(hash, password string) error {
	t.compareCount++
	return t.compareErr
}

type stubTokenProvider struct {
	token string
	err   error
}

func (s stubTokenProvider) Generate(ctx context.Context, user User) (string, error) {
	if s.err != nil {
		return "", s.err
	}
	return s.token, nil
}

type flakySender struct {
	failures int
	sent     int
	sentTo   string
	sentCode string
}

func (f *flakySender) SendVerification(ctx context.Context, email, code string) error {
	if f.sent < f.failures {
		f.sent++
		return errors.New("send failed")
	}
	f.sent++
	f.sentTo = email
	f.sentCode = code
	return nil
}
