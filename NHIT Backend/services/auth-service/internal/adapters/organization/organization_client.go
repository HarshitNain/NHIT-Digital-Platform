package organization

import (
	"context"
	"fmt"

	"github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
)

type organizationClient struct {
	client organizationpb.OrganizationServiceClient
}

func NewOrganizationClient(client organizationpb.OrganizationServiceClient) ports.OrganizationServiceClient {
	return &organizationClient{
		client: client,
	}
}

func (c *organizationClient) CreateOrganization(ctx context.Context, tenantID uuid.UUID, name, code string, createdBy uuid.UUID) (uuid.UUID, error) {
	// Implementation not needed for switch organization, but required by interface
	return uuid.Nil, fmt.Errorf("not implemented")
}

func (c *organizationClient) DeleteOrganization(ctx context.Context, orgID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (c *organizationClient) SetSuperAdmin(ctx context.Context, orgID, superAdminID uuid.UUID) error {
	return fmt.Errorf("not implemented")
}

func (c *organizationClient) GetOrganization(ctx context.Context, orgID uuid.UUID) (*ports.OrganizationInfo, error) {
	// Extract metadata from incoming context and propagate it
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	resp, err := c.client.GetOrganization(ctx, &organizationpb.GetOrganizationRequest{
		OrgId: orgID.String(),
	})
	if err != nil {
		return nil, err
	}

	return convertToOrganization(resp.Organization), nil
}

func (c *organizationClient) ListUserOrganizations(ctx context.Context, userID uuid.UUID) ([]*ports.OrganizationInfo, error) {
	// Extract tenant_id from context metadata
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, fmt.Errorf("missing metadata in context")
	}
	
	tenantIDs := md.Get("tenant_id")
	if len(tenantIDs) == 0 {
		return nil, fmt.Errorf("missing tenant_id in context")
	}

	// Propagate metadata to outgoing context
	ctx = metadata.NewOutgoingContext(ctx, md)
	
	// Call ListOrganizationsByTenant
	resp, err := c.client.ListOrganizationsByTenant(ctx, &organizationpb.ListOrganizationsByTenantRequest{
		TenantId: tenantIDs[0],
		Page:     1,
		PageSize: 100, // Fetch reasonable amount
	})
	if err != nil {
		return nil, err
	}

	var orgs []*ports.OrganizationInfo
	for _, org := range resp.Organizations {
		orgs = append(orgs, convertToOrganization(org))
	}
	
	return orgs, nil
}

func convertToOrganization(org *organizationpb.Organization) *ports.OrganizationInfo {
	if org == nil {
		return nil
	}
	
	orgID, _ := uuid.Parse(org.OrgId)
	tenantID, _ := uuid.Parse(org.TenantId)
	
	return &ports.OrganizationInfo{
		OrgID:    orgID,
		TenantID: tenantID,
		Name:     org.Name,
		Code:     org.Code,
		IsActive: org.Status == organizationpb.OrganizationStatus_activated,
	}
}
