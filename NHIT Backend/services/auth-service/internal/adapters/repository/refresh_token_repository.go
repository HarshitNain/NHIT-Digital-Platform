package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type refreshTokenRepository struct {
	db *pgxpool.Pool
}

func NewRefreshTokenRepository(db *pgxpool.Pool) *refreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(ctx context.Context, token *domain.RefreshToken) error {
	query := `
		INSERT INTO refresh_tokens (token, user_id, expires_at, created_at)
		VALUES ($1, $2, $3, $4)
	`

	_, err := r.db.Exec(
		ctx,
		query,
		token.Token,
		token.UserID,
		token.ExpiresAt,
		token.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to create refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) GetUserIDByToken(ctx context.Context, token string) (uuid.UUID, error) {
	query := `
		SELECT user_id
		FROM refresh_tokens
		WHERE token = $1 AND expires_at > NOW()
	`

	var userID uuid.UUID
	err := r.db.QueryRow(ctx, query, token).Scan(&userID)

	if err == pgx.ErrNoRows {
		return uuid.Nil, fmt.Errorf("refresh token not found or expired")
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("failed to get refresh token: %w", err)
	}

	return userID, nil
}

func (r *refreshTokenRepository) Delete(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`

	_, err := r.db.Exec(ctx, query, token)
	if err != nil {
		return fmt.Errorf("failed to delete refresh token: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM refresh_tokens WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete refresh tokens: %w", err)
	}

	return nil
}

func (r *refreshTokenRepository) UpdateExpiresAt(ctx context.Context, token string, expiresAt time.Time) error {
	query := `UPDATE refresh_tokens SET expires_at = $2 WHERE token = $1`

	_, err := r.db.Exec(ctx, query, token, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to update refresh token expiry: %w", err)
	}

	return nil
}
