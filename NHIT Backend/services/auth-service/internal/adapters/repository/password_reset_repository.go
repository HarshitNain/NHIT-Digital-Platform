package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type passwordResetRepository struct {
	db *pgxpool.Pool
}

func NewPasswordResetRepository(db *pgxpool.Pool) ports.PasswordResetRepository {
	return &passwordResetRepository{db: db}
}

func (r *passwordResetRepository) Create(ctx context.Context, userID uuid.UUID, token uuid.UUID, expiresAt time.Time) (*domain.PasswordReset, error) {
	query := `
		INSERT INTO password_resets (id, token, user_id, reset_type, expires_at, created_at, used)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, token, user_id, reset_type, expires_at, created_at, used
	`

	resetID := uuid.New()
	resetType := "token"
	used := false
	reset := &domain.PasswordReset{}

	err := r.db.QueryRow(
		ctx,
		query,
		resetID,
		token,
		userID,
		resetType,
		expiresAt,
		time.Now(),
		used,
	).Scan(
		&reset.ID,
		&reset.Token,
		&reset.UserID,
		&reset.ResetType,
		&reset.ExpiresAt,
		&reset.CreatedAt,
		&reset.Used,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create password reset: %w", err)
	}

	return reset, nil
}

func (r *passwordResetRepository) GetByToken(ctx context.Context, token uuid.UUID) (*domain.PasswordReset, error) {
	query := `
		SELECT id, token, user_id, reset_type, expires_at, created_at, used
		FROM password_resets
		WHERE token = $1 AND expires_at > NOW() AND used = FALSE AND reset_type = 'token'
	`

	reset := &domain.PasswordReset{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&reset.ID,
		&reset.Token,
		&reset.UserID,
		&reset.ResetType,
		&reset.ExpiresAt,
		&reset.CreatedAt,
		&reset.Used,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("password reset token not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset: %w", err)
	}

	return reset, nil
}

func (r *passwordResetRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM password_resets WHERE id = $1`

	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete password reset: %w", err)
	}

	return nil
}
