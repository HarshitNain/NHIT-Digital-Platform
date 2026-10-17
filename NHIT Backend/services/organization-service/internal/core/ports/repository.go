package ports

import (
	"context"
	"errors"
	"time"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
)

// ErrNotFound indicates the requested organization record does not exist.
var ErrNotFound = errors.New("organization not found")

// OrganizationModel represents the DB-layer model for an organization.
// This will be mapped from SQLC generated structs.
type OrganizationModel struct {
	OrgID        string
	TenantID     string
	ParentOrgID  *string
	Name         string
	Code         string
	DatabaseName string
	Description  *string
	Logo         *string

	// Parent-only fields
	SuperAdminName  *string
	SuperAdminEmail *string
	SuperAdminPass  *string

	InitialProjects []string

	Status    pb.OrganizationStatus
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Pagination metadata returned by repository
type PaginationResult struct {
	CurrentPage int32
	PageSize    int32
	TotalItems  int32
	TotalPages  int32
}

// Repository interface to be implemented by SQLC adapter
type Repository interface {

	// ================
	// CREATE
	// ================
	CreateOrganization(ctx context.Context, org OrganizationModel) (OrganizationModel, error)

	// ================
	// FETCH
	// ================
	GetOrganizationByID(ctx context.Context, orgID string) (OrganizationModel, error)
	GetOrganizationByCode(ctx context.Context, code string) (OrganizationModel, error)

	// ================
	// LIST
	// ================
	ListOrganizations(ctx context.Context, offset, limit int) ([]OrganizationModel, int, error)
	ListOrganizationsByTenant(ctx context.Context, tenantID string, offset, limit int) ([]OrganizationModel, int, error)
	ListChildOrganizations(ctx context.Context, parentOrgID string, offset, limit int) ([]OrganizationModel, int, error)

	// ================
	// UPDATE
	// ================
	UpdateOrganization(ctx context.Context, org OrganizationModel) (OrganizationModel, error)

	// ================
	// DELETE
	// ================
	DeleteOrganization(ctx context.Context, orgID string) error
}
