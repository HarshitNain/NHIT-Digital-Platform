package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// This is a compile-time assertion to ensure passwordResetRepository implements the interface
var _ ports.PasswordResetRepository = (*passwordResetRepository)(nil)

// CreateWithOTP creates a password reset entry with an OTP
func (r *passwordResetRepository) CreateWithOTP(
	ctx context.Context,
	userID uuid.UUID,
	otp string,
	expiresAt time.Time,
) (*domain.PasswordReset, error) {
	query := `
		INSERT INTO password_resets (id, otp, user_id, reset_type, expires_at, created_at, used)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, otp, user_id, reset_type, expires_at, created_at, used
	`

	resetID := uuid.New()
	resetType := "otp"
	used := false
	reset := &domain.PasswordReset{}

	err := r.db.QueryRow(
		ctx,
		query,
		resetID,
		otp,
		userID,
		resetType,
		expiresAt,
		time.Now(),
		used,
	).Scan(
		&reset.ID,
		&reset.OTP,
		&reset.UserID,
		&reset.ResetType,
		&reset.ExpiresAt,
		&reset.CreatedAt,
		&reset.Used,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create password reset with OTP: %w", err)
	}

	return reset, nil
}

// GetByUserIDAndOTP retrieves a password reset entry by user ID and OTP
func (r *passwordResetRepository) GetByUserIDAndOTP(
	ctx context.Context,
	userID uuid.UUID,
	otp string,
) (*domain.PasswordReset, error) {
	query := `
		SELECT id, otp, user_id, reset_type, expires_at, created_at, used
		FROM password_resets
		WHERE user_id = $1 AND otp = $2 AND reset_type = 'otp' AND expires_at > NOW() AND used = FALSE
	`

	reset := &domain.PasswordReset{}
	err := r.db.QueryRow(ctx, query, userID, otp).Scan(
		&reset.ID,
		&reset.OTP,
		&reset.UserID,
		&reset.ResetType,
		&reset.ExpiresAt,
		&reset.CreatedAt,
		&reset.Used,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("password reset OTP not found or expired")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get password reset by OTP: %w", err)
	}

	return reset, nil
}
