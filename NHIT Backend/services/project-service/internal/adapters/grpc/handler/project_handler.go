package handler

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/project-service/internal/core/ports"
)

type projectHandler struct {
	pb.UnimplementedProjectServiceServer
	service ports.ProjectService
}

// NewProjectHandler creates a new project handler
func NewProjectHandler(service ports.ProjectService) pb.ProjectServiceServer {
	return &projectHandler{
		service: service,
	}
}

// GetProject retrieves a project by ID
func (h *projectHandler) GetProject(ctx context.Context, req *pb.GetProjectRequest) (*pb.GetProjectResponse, error) {
	projectID, err := uuid.Parse(req.ProjectId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid project ID: %v", err)
	}

	project, err := h.service.GetProject(ctx, projectID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "project not found: %v", err)
	}

	return &pb.GetProjectResponse{
		Project: toProtoProject(project),
	}, nil
}

// CreateProject creates a new project
func (h *projectHandler) CreateProject(ctx context.Context, req *pb.CreateProjectRequest) (*pb.CreateProjectResponse, error) {
	fmt.Printf("DEBUG CreateProject: TenantID='%s', OrgID='%s', ProjectName='%s', CreatedBy='%s'\n", 
		req.TenantId, req.OrgId, req.ProjectName, req.CreatedBy)
	
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant ID: %v", err)
	}

	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}

	if req.ProjectName == "" {
		return nil, status.Error(codes.InvalidArgument, "project name is required")
	}

	project, err := h.service.CreateProject(ctx, tenantID, orgID, req.ProjectName, req.CreatedBy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create project: %v", err)
	}

	return &pb.CreateProjectResponse{
		Project: toProtoProject(project),
	}, nil
}

// ListProjectsByOrganization lists all projects for an organization with pagination
func (h *projectHandler) ListProjectsByOrganization(ctx context.Context, req *pb.ListProjectsByOrganizationRequest) (*pb.ListProjectsByOrganizationResponse, error) {
	fmt.Printf("DEBUG PROJECT HANDLER: ListProjectsByOrganization called for OrgID=%s, Page=%d, PageSize=%d\n", req.OrgId, req.Page, req.PageSize)
	
	orgID, err := uuid.Parse(req.OrgId)
	if err != nil {
		fmt.Printf("DEBUG PROJECT HANDLER: Invalid OrgID format: %v\n", err)
		return nil, status.Errorf(codes.InvalidArgument, "invalid organization ID: %v", err)
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	projects, totalCount, err := h.service.ListProjectsByOrganization(ctx, orgID, page, pageSize)
	if err != nil {
		fmt.Printf("DEBUG PROJECT HANDLER: Service returned error: %v\n", err)
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}
	
	fmt.Printf("DEBUG PROJECT HANDLER: Found %d projects (Total: %d) for OrgID=%s\n", len(projects), totalCount, req.OrgId)

	var pbProjects []*pb.Project
	for _, p := range projects {
		pbProjects = append(pbProjects, toProtoProject(p))
	}

	totalPages := (int32(totalCount) + pageSize - 1) / pageSize

	return &pb.ListProjectsByOrganizationResponse{
		Projects:   pbProjects,
		TotalCount: int32(totalCount),
		Pagination: &pb.PaginationMetadata{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  int32(totalCount),
			TotalPages:  totalPages,
		},
	}, nil
}

// Helper function to convert domain project to proto project
func toProtoProject(p *domain.Project) *pb.Project {
	project := &pb.Project{
		ProjectId:   p.ProjectID.String(),
		TenantId:    p.TenantID.String(),
		OrgId:       p.OrgID.String(),
		ProjectName: p.ProjectName,
		CreatedBy:   p.CreatedBy,
		CreatedAt:   timestamppb.New(p.CreatedAt),
		UpdatedAt:   timestamppb.New(p.UpdatedAt),
	}

	return project
}
