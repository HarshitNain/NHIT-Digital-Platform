package repository

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
)

type roleRepository struct {
	queries *sqlc.Queries
}

func NewRoleRepository(queries *sqlc.Queries) ports.RoleRepository {
	return &roleRepository{queries: queries}
}
func (r *roleRepository) Create(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	params := sqlc.CreateRoleParams{
		TenantID:     toPgUUIDFromVal(role.TenantID),
		OrgID:        toPgUUID(role.OrgID),
		Name:         role.Name,
		Description:  role.Description,
		Permissions:  role.Permissions,
		IsSystemRole: role.IsSystemRole,
		CreatedBy:    role.CreatedBy,
	}

	dbRole, err := getQueries(ctx, r.queries).CreateRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainRole(dbRole), nil
}

func (r *roleRepository) GetByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	dbRole, err := getQueries(ctx, r.queries).GetRole(ctx, toPgUUIDFromVal(roleID))
	if err != nil {
		return nil, err
	}
	return toDomainRole(dbRole), nil
}
func (r *roleRepository) ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error) {
	params := sqlc.ListRolesByTenantParams{
		TenantID: toPgUUIDFromVal(tenantID),
		Limit:    limit,
		Offset:   offset,
	}
	q := getQueries(ctx, r.queries)
	dbRoles, err := q.ListRolesByTenant(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	total, err := q.CountRolesByTenant(ctx, toPgUUIDFromVal(tenantID))
	if err != nil {
		return nil, 0, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, total, nil
}
func (r *roleRepository) ListByTenantAndOrg(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) ([]*domain.Role, error) {
	params := sqlc.ListRolesByTenantAndOrgParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
	}

	dbRoles, err := getQueries(ctx, r.queries).ListRolesByTenantAndOrg(ctx, params)
	if err != nil {
		return nil, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, nil
}
func (r *roleRepository) ListByTenantAndOrgIncludingSystem(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error) {
	params := sqlc.ListRolesByOrganizationIncludingSystemParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
		Limit:    limit,
		Offset:   offset,
	}

	q := getQueries(ctx, r.queries)
	dbRoles, err := q.ListRolesByOrganizationIncludingSystem(ctx, params)
	if err != nil {
		return nil, 0, err
	}

	countParams := sqlc.CountRolesByOrganizationIncludingSystemParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
	}
	total, err := q.CountRolesByOrganizationIncludingSystem(ctx, countParams)
	if err != nil {
		return nil, 0, err
	}

	roles := make([]*domain.Role, len(dbRoles))
	for i, dbRole := range dbRoles {
		roles[i] = toDomainRole(dbRole)
	}
	return roles, total, nil
}

func (r *roleRepository) Update(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	params := sqlc.UpdateRoleParams{
		RoleID:      toPgUUIDFromVal(role.RoleID),
		Name:        role.Name,
		Description: role.Description,
		Permissions: role.Permissions,
	}

	dbRole, err := getQueries(ctx, r.queries).UpdateRole(ctx, params)
	if err != nil {
		return nil, err
	}

	return toDomainRole(dbRole), nil
}

func (r *roleRepository) Delete(ctx context.Context, roleID uuid.UUID) error {
	return getQueries(ctx, r.queries).DeleteRole(ctx, toPgUUIDFromVal(roleID))
}
func (r *roleRepository) Exists(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name string) (bool, error) {
	return getQueries(ctx, r.queries).RoleExists(ctx, sqlc.RoleExistsParams{
		TenantID: toPgUUIDFromVal(tenantID),
		OrgID:    toPgUUID(orgID),
		Name:     name,
	})
}

func (r *roleRepository) ExistsSystem(ctx context.Context, tenantID uuid.UUID, name string) (bool, error) {
	return getQueries(ctx, r.queries).RoleExistsSystem(ctx, sqlc.RoleExistsSystemParams{
		TenantID: toPgUUIDFromVal(tenantID),
		Name:     name,
	})
}

func toDomainRole(dbRole *sqlc.Role) *domain.Role {
	var orgID *uuid.UUID
	if dbRole.ParentOrgID.Valid {
		id := dbRole.ParentOrgID.UUID
		orgID = &id
	}

	var createdBy string
	if dbRole.CreatedBy != nil {
		createdBy = *dbRole.CreatedBy
	}

	return &domain.Role{
		RoleID:       dbRole.RoleID,
		TenantID:     dbRole.TenantID,
		OrgID:        orgID,
		Name:         dbRole.Name,
		Description:  dbRole.Description,
		Permissions:  dbRole.Permissions,
		IsSystemRole: dbRole.IsSystemRole,
		CreatedBy:    createdBy,
		CreatedAt:    dbRole.CreatedAt.Time,
		UpdatedAt:    dbRole.UpdatedAt.Time,
	}
}
