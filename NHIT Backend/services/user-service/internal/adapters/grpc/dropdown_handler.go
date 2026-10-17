package grpc

import (
	"context"

	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	designationpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

// GetDepartmentsDropdown fetches organization-specific departments from department-service
func (h *UserHandler) GetDepartmentsDropdown(ctx context.Context, req *userpb.GetDropdownRequest) (*userpb.DepartmentsDropdownResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok && req.OrgId != "" {
		orgID = req.OrgId
	}

	if orgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}

	// Forward metadata (including authorization) to department-service
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Call department-service via gRPC
	deptClient := departmentpb.NewDepartmentServiceClient(h.deptConn)
	deptResp, err := deptClient.ListDepartments(ctx, &departmentpb.ListDepartmentsRequest{
		Page:     1,
		PageSize: 1000, // Get all departments for dropdown
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch departments from department-service: %v", err)
	}

	// Convert to dropdown format
	var departments []*userpb.DropdownItem
	for _, dept := range deptResp.Departments {
		departments = append(departments, &userpb.DropdownItem{
			Id:   dept.Id,
			Name: dept.Name,
		})
	}

	return &userpb.DepartmentsDropdownResponse{
		Departments: departments,
	}, nil
}

// GetDesignationsDropdown fetches organization-specific designations from designation-service
func (h *UserHandler) GetDesignationsDropdown(ctx context.Context, req *userpb.GetDropdownRequest) (*userpb.DesignationsDropdownResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok && req.OrgId != "" {
		orgID = req.OrgId
	}

	if orgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}

	// Forward metadata (including authorization) to designation-service
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ctx = metadata.NewOutgoingContext(ctx, md)
	}

	// Call designation-service via gRPC
	desigClient := designationpb.NewDesignationServiceClient(h.desigConn)
	desigResp, err := desigClient.ListDesignations(ctx, &designationpb.ListDesignationsRequest{
		Page:     1,
		PageSize: 1000, // Get all designations for dropdown
	})
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch designations from designation-service: %v", err)
	}

	// Convert to dropdown format
	var designations []*userpb.DropdownItem
	for _, desig := range desigResp.Designations {
		designations = append(designations, &userpb.DropdownItem{
			Id:   desig.Id,
			Name: desig.Name,
		})
	}

	return &userpb.DesignationsDropdownResponse{
		Designations: designations,
	}, nil
}

// GetRolesDropdown fetches organization-specific roles from user-service database
func (h *UserHandler) GetRolesDropdown(ctx context.Context, req *userpb.GetDropdownRequest) (*userpb.RolesDropdownResponse, error) {
	orgID, ok := middleware.GetOrgIDFromContext(ctx)
	if !ok && req.OrgId != "" {
		orgID = req.OrgId
	}

	if orgID == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}

	// Query roles from local database (organization-specific)
	// Roles can be organization-specific (parent_org_id = org_id) or system roles (parent_org_id IS NULL)
	rows, err := h.db.Query(ctx, `
		SELECT role_id, name 
		FROM roles 
		WHERE parent_org_id = $1 OR parent_org_id IS NULL
		ORDER BY name ASC
	`, orgID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to fetch roles: %v", err)
	}
	defer rows.Close()

	var roles []*userpb.DropdownItem
	for rows.Next() {
		var id, name string
		if err := rows.Scan(&id, &name); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan role: %v", err)
		}
		roles = append(roles, &userpb.DropdownItem{
			Id:   id,
			Name: name,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, status.Errorf(codes.Internal, "error iterating roles: %v", err)
	}

	return &userpb.RolesDropdownResponse{
		Roles: roles,
	}, nil
}
