package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/jackc/pgx/v5/pgxpool"
)

type emailVerificationRepository struct {
	db *pgxpool.Pool
}

func NewEmailVerificationRepository(db *pgxpool.Pool) *emailVerificationRepository {
	return &emailVerificationRepository{db: db}
}

func (r *emailVerificationRepository) Create(ctx context.Context, userID uuid.UUID, expiresAt time.Time) (*domain.EmailVerificationToken, error) {
	// First, delete any existing verification token for this user
	deleteQuery := `DELETE FROM email_verification_tokens WHERE user_id = $1`
	_, err := r.db.Exec(ctx, deleteQuery, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete existing verification token: %w", err)
	}

	// Now insert the new token
	token := uuid.New()
	query := `
		INSERT INTO email_verification_tokens (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
		RETURNING token, user_id, expires_at, created_at
	`

	verification := &domain.EmailVerificationToken{}
	err = r.db.QueryRow(
		ctx,
		query,
		token,
		userID,
		expiresAt,
		time.Now(),
	).Scan(
		&verification.Token,
		&verification.UserID,
		&verification.ExpiresAt,
		&verification.CreatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create email verification token: %w", err)
	}

	return verification, nil
}

func (r *emailVerificationRepository) Verify(ctx context.Context, userID uuid.UUID, token uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM email_verification_tokens
		WHERE user_id = $1 AND token = $2 AND expires_at > NOW()
	`

	var count int
	err := r.db.QueryRow(ctx, query, userID, token).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to verify email token: %w", err)
	}

	return count > 0, nil
}

func (r *emailVerificationRepository) Delete(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM email_verification_tokens WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete email verification token: %w", err)
	}

	return nil
}
