package postgres

import (
	"context"
	"errors"
	"fmt"

	"catalog-api/internal/identity"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// IdentityRepository persists identity data in Postgres.
type IdentityRepository struct {
	pool *pgxpool.Pool
}

func NewIdentityRepository(pool *pgxpool.Pool) *IdentityRepository {
	return &IdentityRepository{pool: pool}
}

func (r *IdentityRepository) CreateUser(ctx context.Context, user identity.User) (identity.User, error) {
	if r.pool == nil {
		return identity.User{}, identity.ErrRepositoryNotConfigured
	}
	query := `
		INSERT INTO users (email, full_name, password_hash, role, status, is_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, full_name, password_hash, role, status, is_verified, created_at, updated_at
	`
	row := r.pool.QueryRow(ctx, query,
		user.Email,
		user.FullName,
		user.PasswordHash,
		user.Role,
		user.Status,
		user.IsVerified,
	)
	created, err := scanUser(row)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return identity.User{}, identity.ErrEmailAlreadyRegistered
		}
		return identity.User{}, err
	}
	return created, nil
}

func (r *IdentityRepository) GetByEmail(ctx context.Context, email string) (identity.User, error) {
	if r.pool == nil {
		return identity.User{}, identity.ErrRepositoryNotConfigured
	}
	query := `
		SELECT id, email, full_name, password_hash, role, status, is_verified, created_at, updated_at
		FROM users
		WHERE email = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, email)
	return scanUser(row)
}

func (r *IdentityRepository) GetByID(ctx context.Context, id identity.UserID) (identity.User, error) {
	if r.pool == nil {
		return identity.User{}, identity.ErrRepositoryNotConfigured
	}
	query := `
		SELECT id, email, full_name, password_hash, role, status, is_verified, created_at, updated_at
		FROM users
		WHERE id = $1
		LIMIT 1
	`
	row := r.pool.QueryRow(ctx, query, id)
	return scanUser(row)
}

func (r *IdentityRepository) SetVerification(ctx context.Context, userID identity.UserID, verified bool) error {
	if r.pool == nil {
		return identity.ErrRepositoryNotConfigured
	}
	cmdTag, err := r.pool.Exec(ctx, `UPDATE users SET is_verified = $1 WHERE id = $2`, verified, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for verification user_id=%s", userID)
	}
	return nil
}

func (r *IdentityRepository) UpdateStatus(ctx context.Context, userID identity.UserID, status identity.UserStatus) error {
	if r.pool == nil {
		return identity.ErrRepositoryNotConfigured
	}
	cmdTag, err := r.pool.Exec(ctx, `UPDATE users SET status = $1 WHERE id = $2`, status, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated for status user_id=%s", userID)
	}
	return nil
}

func (r *IdentityRepository) EnsureRole(ctx context.Context, role identity.RoleName) error {
	if r.pool == nil {
		return identity.ErrRepositoryNotConfigured
	}
	_, err := r.pool.Exec(ctx, `INSERT INTO roles (name) VALUES ($1) ON CONFLICT DO NOTHING`, role)
	return err
}

func (r *IdentityRepository) AssignRole(ctx context.Context, userID identity.UserID, role identity.RoleName) error {
	if r.pool == nil {
		return identity.ErrRepositoryNotConfigured
	}
	cmdTag, err := r.pool.Exec(ctx, `UPDATE users SET role = $1 WHERE id = $2`, role, userID)
	if err != nil {
		return err
	}
	if cmdTag.RowsAffected() == 0 {
		return fmt.Errorf("no rows updated when assigning role user_id=%s", userID)
	}
	return nil
}

func (r *IdentityRepository) UpdateUserProfile(ctx context.Context, user identity.User) (identity.User, error) {
	if r.pool == nil {
		return identity.User{}, identity.ErrRepositoryNotConfigured
	}
	query := `UPDATE users SET full_name = $1, updated_at = NOW() WHERE id = $2 RETURNING id, email, full_name, password_hash, role, status, is_verified, created_at, updated_at`
	row := r.pool.QueryRow(ctx, query, user.FullName, user.ID)
	return scanUser(row)
}

func scanUser(row pgx.Row) (identity.User, error) {
	var u identity.User
	if err := row.Scan(
		&u.ID,
		&u.Email,
		&u.FullName,
		&u.PasswordHash,
		&u.Role,
		&u.Status,
		&u.IsVerified,
		&u.CreatedAt,
		&u.UpdatedAt,
	); err != nil {
		return identity.User{}, err
	}
	return u, nil
}
