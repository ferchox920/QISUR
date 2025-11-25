package identity

import (
	"context"
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
func (stubUserRepo) EnsureRole(ctx context.Context, role RoleName) error                { return nil }
func (stubUserRepo) AssignRole(ctx context.Context, userID UserID, role RoleName) error { return nil }

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
