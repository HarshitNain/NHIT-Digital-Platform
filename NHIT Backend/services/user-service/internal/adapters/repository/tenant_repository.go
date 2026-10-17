package repository

import (
    "context"

    "github.com/google/uuid"
    "github.com/jackc/pgx/v5/pgtype"
    sqlc "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/adapters/repository/sqlc/generated"
    "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
    "github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
)

type tenantRepository struct {
    queries *sqlc.Queries
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(queries *sqlc.Queries) ports.TenantRepository {
    return &tenantRepository{queries: queries}
}

func (r *tenantRepository) Create(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error) {
    params := sqlc.CreateTenantParams{
        TenantID: pgtype.UUID{Bytes: tenant.TenantID, Valid: true},
        Name:     tenant.Name,
        Email:    tenant.Email,
        Password: tenant.Password,
    }

    dbTenant, err := r.queries.CreateTenant(ctx, params)
    if err != nil {
        return nil, err
    }

    return toDomainTenant(dbTenant), nil
}

func (r *tenantRepository) GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
    dbTenant, err := r.queries.GetTenant(ctx, pgtype.UUID{Bytes: tenantID, Valid: true})
    if err != nil {
        return nil, err
    }

    return toDomainTenant(dbTenant), nil
}

func toDomainTenant(dbTenant *sqlc.Tenant) *domain.Tenant {
    return &domain.Tenant{
        TenantID:  uuid.UUID(dbTenant.TenantID.Bytes),
        Name:      dbTenant.Name,
        Email:     dbTenant.Email,
        Password:  dbTenant.Password,
        CreatedAt: dbTenant.CreatedAt.Time,
        UpdatedAt: dbTenant.UpdatedAt.Time,
    }
}
