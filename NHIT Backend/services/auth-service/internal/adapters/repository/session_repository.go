package repository

import (
	"context"
	"fmt"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type sessionRepository struct {
	db *pgxpool.Pool
}

// Ensure sessionRepository implements ports.SessionRepository at compile time
var _ ports.SessionRepository = (*sessionRepository)(nil)

func NewSessionRepository(db *pgxpool.Pool) ports.SessionRepository {
	return &sessionRepository{db: db}
}

func (r *sessionRepository) Create(ctx context.Context, session *domain.Session) (*domain.Session, error) {
	query := `
		INSERT INTO sessions (session_id, user_id, session_token, created_at, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING session_id, user_id, session_token, created_at, expires_at
	`

	err := r.db.QueryRow(
		ctx,
		query,
		session.SessionID,
		session.UserID,
		session.SessionToken,
		session.CreatedAt,
		session.ExpiresAt,
	).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) GetByID(ctx context.Context, sessionID uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, session_token, created_at, expires_at
		FROM sessions
		WHERE session_id = $1 AND expires_at > NOW()
	`

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, sessionID).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) GetByToken(ctx context.Context, token string) (*domain.Session, error) {
	query := `
		SELECT session_id, user_id, session_token, created_at, expires_at
		FROM sessions
		WHERE session_token = $1 AND expires_at > NOW()
	`

	session := &domain.Session{}
	err := r.db.QueryRow(ctx, query, token).Scan(
		&session.SessionID,
		&session.UserID,
		&session.SessionToken,
		&session.CreatedAt,
		&session.ExpiresAt,
	)

	if err == pgx.ErrNoRows {
		return nil, fmt.Errorf("session not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get session: %w", err)
	}

	return session, nil
}

func (r *sessionRepository) Delete(ctx context.Context, sessionID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE session_id = $1`

	result, err := r.db.Exec(ctx, query, sessionID)
	if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	rows := result.RowsAffected()

	if rows == 0 {
		return fmt.Errorf("session not found")
	}

	return nil
}

func (r *sessionRepository) GetByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	query := `
		SELECT session_id, user_id, session_token, created_at, expires_at
		FROM sessions
		WHERE user_id = $1 AND expires_at > NOW()
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get sessions: %w", err)
	}
	defer rows.Close()

	var sessions []*domain.Session
	for rows.Next() {
		session := &domain.Session{}
		err := rows.Scan(
			&session.SessionID,
			&session.UserID,
			&session.SessionToken,
			&session.CreatedAt,
			&session.ExpiresAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan session: %w", err)
		}
		sessions = append(sessions, session)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to iterate sessions: %w", err)
	}

	return sessions, nil
}

func (r *sessionRepository) DeleteByUserID(ctx context.Context, userID uuid.UUID) error {
	query := `DELETE FROM sessions WHERE user_id = $1`

	_, err := r.db.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("failed to delete sessions: %w", err)
	}

	return nil
}
