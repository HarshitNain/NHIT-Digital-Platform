package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/domain"
)

// Pagination
type PaginationParams struct {
	Page     int32
	PageSize int32
}

type PaginatedOrganizations struct {
	Organizations []domain.Organization
	TotalCount    int
	CurrentPage   int32
	PageSize      int32
	TotalPages    int32
}

// Service defines business logic use-cases
type OrganizationService interface {

	// Create Parent or Child Organization
	CreateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)

	// Get
	GetOrganizationByID(ctx context.Context, orgID string) (domain.Organization, error)
	GetOrganizationByCode(ctx context.Context, code string) (domain.Organization, error)

	// List
	ListOrganizations(ctx context.Context, params PaginationParams) (PaginatedOrganizations, error)
	ListOrganizationsByTenant(ctx context.Context, tenantID string, params PaginationParams) (PaginatedOrganizations, error)
	ListChildOrganizations(ctx context.Context, parentOrgID string, params PaginationParams) (PaginatedOrganizations, error)

	// Update
	UpdateOrganization(ctx context.Context, org domain.Organization) (domain.Organization, error)

	// Delete
	DeleteOrganization(ctx context.Context, orgID string) error
}
