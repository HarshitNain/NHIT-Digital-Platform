package ports

import (
	"context"

	"github.com/google/uuid"
)

// OrganizationServiceClient defines the interface for gRPC communication with organization service
type OrganizationServiceClient interface {
	CreateOrganization(ctx context.Context, tenantID uuid.UUID, name, code string, createdBy uuid.UUID) (uuid.UUID, error)
	DeleteOrganization(ctx context.Context, orgID uuid.UUID) error
	SetSuperAdmin(ctx context.Context, orgID, superAdminID uuid.UUID) error
	GetOrganization(ctx context.Context, orgID uuid.UUID) (*OrganizationInfo, error)
	ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*OrganizationInfo, error)
}

// OrganizationInfo represents organization information
type OrganizationInfo struct {
	OrgID         uuid.UUID
	TenantID      uuid.UUID
	Name          string
	Code          string
	IsActive      bool
	SuperAdminID  uuid.UUID
	UserCount     int32
	IsGlobalSchema bool
}
