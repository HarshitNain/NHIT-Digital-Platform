package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userRepository struct {
	db *pgxpool.Pool
}

// Ensure userRepository implements ports.UserRepository at compile time
var _ ports.UserRepository = (*userRepository)(nil)

func NewUserRepository(db *pgxpool.Pool) ports.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) GetByEmail(ctx context.Context, tenantID uuid.UUID, email string) (*ports.UserData, error) {
	query := `
		SELECT user_id, tenant_id, email, name, password, email_verified_at, is_active
		FROM users
		WHERE tenant_id = $1 AND email = $2
	`

	user := &ports.UserData{}
	err := r.db.QueryRow(ctx, query, tenantID, email).Scan(
		&user.UserID,
		&user.TenantID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *ports.UserData) (*ports.UserData, error) {
	query := `
		INSERT INTO users (user_id, tenant_id, email, name, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING user_id, tenant_id, email, name, password, email_verified_at, is_active, created_at, updated_at
	`

	now := time.Now()
	createdUser := &ports.UserData{}

	err := r.db.QueryRow(ctx, query,
		user.UserID,
		user.TenantID,
		user.Email,
		user.Name,
		user.Password,
		now,
		now,
	).Scan(
		&createdUser.UserID,
		&createdUser.TenantID,
		&createdUser.Email,
		&createdUser.Name,
		&createdUser.Password,
		&createdUser.EmailVerifiedAt,
		&createdUser.IsActive,
		&createdUser.CreatedAt,
		&createdUser.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

// GetByEmailGlobal gets user by email across all tenants - for tenant-agnostic login
func (r *userRepository) GetByEmailGlobal(ctx context.Context, email string) (*ports.UserData, error) {
	query := `
		SELECT user_id, tenant_id, email, name, password, email_verified_at, is_active
		FROM users
		WHERE email = $1
		LIMIT 1
	`

	user := &ports.UserData{}
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.UserID,
		&user.TenantID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user, nil
}

func (r *userRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, hashedPassword string) error {
	query := `
		UPDATE users
		SET password = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.db.Exec(ctx, query, hashedPassword, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	rows := result.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID, ipAddress, userAgent string) error {
	query := `
		UPDATE users
		SET last_login_at = $1, last_login_ip = $2, updated_at = $3
		WHERE user_id = $4
	`

	_, err := r.db.Exec(ctx, query, time.Now(), ipAddress, time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last login: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateLastLogout(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_logout_at = $1, updated_at = $2
		WHERE user_id = $3
	`

	_, err := r.db.Exec(ctx, query, time.Now(), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to update last logout: %w", err)
	}

	return nil
}

func (r *userRepository) VerifyEmail(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users
		SET email_verified_at = $1, updated_at = $2
		WHERE user_id = $3
	`

	result, err := r.db.Exec(ctx, query, time.Now(), time.Now(), userID)
	if err != nil {
		return fmt.Errorf("failed to verify email: %w", err)
	}

	rows := result.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM users WHERE user_id = $1`

	result, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete user: %w", err)
	}

	rows := result.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("user not found")
	}

	return nil
}

func (r *userRepository) GetByID(ctx context.Context, userID uuid.UUID) (*ports.UserData, error) {
	query := `
		SELECT user_id, tenant_id, email, name, password, email_verified_at, is_active
		FROM users
		WHERE user_id = $1
	`

	user := &ports.UserData{}
	err := r.db.QueryRow(ctx, query, userID).Scan(
		&user.UserID,
		&user.TenantID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.EmailVerifiedAt,
		&user.IsActive,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get user by ID: %w", err)
	}

	return user, nil
}

func (r *userRepository) UpdateTenantPassword(ctx context.Context, email string, hashedPassword string) error {
	query := `
		UPDATE tenants
		SET password = $1, updated_at = $2
		WHERE email = $3
	`

	_, err := r.db.Exec(ctx, query, hashedPassword, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to update tenant password: %w", err)
	}

	return nil
}

func (r *userRepository) UpdateOrganizationSuperAdminPassword(ctx context.Context, email string, hashedPassword string) error {
	query := `
		UPDATE organizations
		SET super_admin_password = $1, updated_at = $2
		WHERE super_admin_email = $3
	`

	_, err := r.db.Exec(ctx, query, hashedPassword, time.Now(), email)
	if err != nil {
		return fmt.Errorf("failed to update organization super admin password: %w", err)
	}

	return nil
}
