package services

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// SessionCleanupService handles automatic cleanup of expired sessions and refresh tokens
type SessionCleanupService struct {
	pool            *pgxpool.Pool
	cleanupInterval time.Duration
	logger          *slog.Logger
}

// NewSessionCleanupService creates a new session cleanup service
func NewSessionCleanupService(pool *pgxpool.Pool, cleanupInterval time.Duration, logger *slog.Logger) *SessionCleanupService {
	return &SessionCleanupService{
		pool:            pool,
		cleanupInterval: cleanupInterval,
		logger:          logger,
	}
}

// Start begins the automatic cleanup process
// It runs in a goroutine and stops when the context is cancelled
func (s *SessionCleanupService) Start(ctx context.Context) {
	s.logger.Info("Session cleanup service started", "interval", s.cleanupInterval)

	ticker := time.NewTicker(s.cleanupInterval)
	defer ticker.Stop()

	// Run cleanup immediately on startup
	s.cleanup(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info("Session cleanup service stopping")
			return
		case <-ticker.C:
			s.cleanup(ctx)
		}
	}
}

// cleanup performs the actual cleanup of expired sessions and refresh tokens
func (s *SessionCleanupService) cleanup(ctx context.Context) {
	startTime := time.Now()

	// Cleanup expired sessions
	sessionsDeleted, err := s.cleanupSessions(ctx)
	if err != nil {
		s.logger.Error("Error cleaning up sessions", "error", err)
	}

	// Cleanup expired refresh tokens
	tokensDeleted, err := s.cleanupRefreshTokens(ctx)
	if err != nil {
		s.logger.Error("Error cleaning up refresh tokens", "error", err)
	}

	duration := time.Since(startTime)
	s.logger.Info("Cleanup complete",
		"deleted_sessions", sessionsDeleted,
		"deleted_refresh_tokens", tokensDeleted,
		"duration", duration)
}

// cleanupSessions deletes expired sessions from the database
func (s *SessionCleanupService) cleanupSessions(ctx context.Context) (int64, error) {
	query := `DELETE FROM sessions WHERE expires_at < NOW()`

	result, err := s.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired sessions: %w", err)
	}

	return result.RowsAffected(), nil
}

// cleanupRefreshTokens deletes expired refresh tokens from the database
func (s *SessionCleanupService) cleanupRefreshTokens(ctx context.Context) (int64, error) {
	query := `DELETE FROM refresh_tokens WHERE expires_at < NOW()`

	result, err := s.pool.Exec(ctx, query)
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired refresh tokens: %w", err)
	}

	return result.RowsAffected(), nil
}
