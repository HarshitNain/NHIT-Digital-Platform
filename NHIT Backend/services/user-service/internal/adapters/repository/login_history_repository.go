package repository

import (
	"context"
	"fmt"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

type loginHistoryRepository struct {
	queries *sqlc.Queries
}

// NewLoginHistoryRepository creates a new login history repository
func NewLoginHistoryRepository(queries *sqlc.Queries) ports.LoginHistoryRepository {
	return &loginHistoryRepository{queries: queries}
}

// Create creates a new login history entry
func (r *loginHistoryRepository) Create(ctx context.Context, history *domain.UserLoginHistory) (*domain.UserLoginHistory, error) {
	params := sqlc.CreateLoginHistoryParams{
		UserID:    uuid.NullUUID{UUID: history.UserID, Valid: true},
		IpAddress: history.IPAddress,
		UserAgent: history.UserAgent,
		LoginTime: pgtype.Timestamptz{Time: history.LoginTime, Valid: true},
	}

	createdHistory, err := r.queries.CreateLoginHistory(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("failed to create login history: %w", err)
	}

	return &domain.UserLoginHistory{
		HistoryID: createdHistory.HistoryID,
		UserID:    createdHistory.UserID.UUID,
		IPAddress: createdHistory.IpAddress,
		UserAgent: createdHistory.UserAgent,
		LoginTime: createdHistory.LoginTime.Time,
	}, nil
}

// ListByUser lists login history for a specific user with total count
func (r *loginHistoryRepository) ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.UserLoginHistory, int64, error) {
	// Get total count
	total, err := r.queries.CountUserLoginHistories(ctx, uuid.NullUUID{UUID: userID, Valid: true})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count login histories: %w", err)
	}

	params := sqlc.ListUserLoginHistoriesParams{
		UserID: uuid.NullUUID{UUID: userID, Valid: true},
		Limit:  limit,
		Offset: offset,
	}

	histories, err := r.queries.ListUserLoginHistories(ctx, params)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list login histories: %w", err)
	}

	result := make([]*domain.UserLoginHistory, len(histories))
	for i, h := range histories {
		result[i] = &domain.UserLoginHistory{
			HistoryID: h.HistoryID,
			UserID:    h.UserID.UUID,
			IPAddress: h.IpAddress,
			UserAgent: h.UserAgent,
			LoginTime: h.LoginTime.Time,
		}
	}

	return result, total, nil
}
