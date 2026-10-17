package repository

import (
	"context"

	sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
)

type permissionRepository struct {
	queries *sqlc.Queries
}

func NewPermissionRepository(queries *sqlc.Queries) ports.PermissionRepository {
	return &permissionRepository{queries: queries}
}

func (r *permissionRepository) ListAll(ctx context.Context) ([]*domain.Permission, error) {
	dbPerms, err := r.queries.ListPermissions(ctx)
	if err != nil {
		return nil, err
	}

	perms := make([]*domain.Permission, len(dbPerms))
	for i, p := range dbPerms {
		perms[i] = toDomainPermission(p)
	}
	return perms, nil
}

func (r *permissionRepository) ListByModule(ctx context.Context, module *string) ([]*domain.Permission, error) {
	dbPerms, err := r.queries.ListPermissionsByModule(ctx, module)
	if err != nil {
		return nil, err
	}

	perms := make([]*domain.Permission, len(dbPerms))
	for i, p := range dbPerms {
		perms[i] = toDomainPermission(p)
	}
	return perms, nil
}

func toDomainPermission(p *sqlc.Permission) *domain.Permission {
	var module string
	if p.Module != nil {
		module = *p.Module
	}

	var action string
	if p.Action != nil {
		action = *p.Action
	}

	return &domain.Permission{
		PermissionID:       uuid.UUID(p.PermissionID.Bytes),
		Name:               p.Name,
		Description:        p.Description,
		Module:             module,
		Action:             action,
		IsSystemPermission: p.IsSystemPermission,
	}
}
