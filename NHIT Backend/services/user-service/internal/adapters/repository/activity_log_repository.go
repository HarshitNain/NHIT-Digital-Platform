package repository

import (
	"context"

	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type activityLogRepository struct {
	queries *sqlc.Queries
}

func NewActivityLogRepository(queries *sqlc.Queries) ports.ActivityLogRepository {
	return &activityLogRepository{queries: queries}
}

func (r *activityLogRepository) Create(ctx context.Context, log *domain.ActivityLog) (*domain.ActivityLog, error) {
	params := sqlc.CreateActivityLogParams{
		Name:        log.Name,
		Description: log.Description,
	}

	createdLog, err := r.queries.CreateActivityLog(ctx, params)
	if err != nil {
		return nil, err
	}

	return &domain.ActivityLog{
		ID:          createdLog.ID,
		Name:        createdLog.Name,
		Description: createdLog.Description,
		CreatedAt:   createdLog.CreatedAt.Time,
	}, nil
}

func (r *activityLogRepository) List(ctx context.Context, limit, offset int32) ([]*domain.ActivityLog, int64, error) {
	total, err := r.queries.CountActivityLogs(ctx)
	if err != nil {
		return nil, 0, err
	}

	params := sqlc.ListActivityLogsParams{
		Limit:  limit,
		Offset: offset,
	}

	logs, err := r.queries.ListActivityLogs(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	result := make([]*domain.ActivityLog, len(logs))
	for i, log := range logs {
		result[i] = &domain.ActivityLog{
			ID:          log.ID,
			Name:        log.Name,
			Description: log.Description,
			CreatedAt:   log.CreatedAt.Time,
		}
	}

	return result, total, nil
}

func (r *activityLogRepository) Count(ctx context.Context) (int64, error) {
	return r.queries.CountActivityLogs(ctx)
}
