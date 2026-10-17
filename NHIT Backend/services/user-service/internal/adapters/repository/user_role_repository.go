package repository

import (
	"context"

	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
)

type userRoleRepository struct {
	queries *sqlc.Queries
}

// NewUserRoleRepository creates a new user role repository
func NewUserRoleRepository(queries *sqlc.Queries) ports.UserRoleRepository {
	return &userRoleRepository{queries: queries}
}

func (r *userRoleRepository) AssignRole(ctx context.Context, userID, roleID uuid.UUID) error {
	params := sqlc.AssignRoleToUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return getQueries(ctx, r.queries).AssignRoleToUser(ctx, params)
}

func (r *userRoleRepository) RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error {
	params := sqlc.RemoveRoleFromUserParams{
		UserID: userID,
		RoleID: roleID,
	}
	return getQueries(ctx, r.queries).RemoveRoleFromUser(ctx, params)
}

func (r *userRoleRepository) ListRolesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	dbRoles, err := getQueries(ctx, r.queries).ListDetailedRolesForUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}

	return roles, nil
}
