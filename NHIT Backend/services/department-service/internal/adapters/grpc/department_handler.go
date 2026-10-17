package grpc

import (
	"context"

	"strings"
	"log/slog"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"

	departmentpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/department-service/internal/core/ports"
)

type DepartmentHandler struct {
	departmentpb.UnimplementedDepartmentServiceServer
	service ports.DepartmentService
	logger  *slog.Logger
}

// NewDepartmentHandler creates a new department gRPC handler
func NewDepartmentHandler(service ports.DepartmentService, logger *slog.Logger) *DepartmentHandler {
	return &DepartmentHandler{
		service: service,
		logger:  logger,
	}
}

// helper to get first non-empty metadata value by keys
func firstMetadataValue(md metadata.MD, keys ...string) string {
	for _, k := range keys {
		if vals := md[strings.ToLower(k)]; len(vals) > 0 && vals[0] != "" {
			return vals[0]
		}
	}
	return ""
}

// getOrgIDFromContext extracts org_id from metadata
func getOrgIDFromContext(ctx context.Context) *uuid.UUID {
	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	// Check headers first (if context switching)
	orgIDStr := firstMetadataValue(md, "x-org-id", "org-id", "orgId", "org_id")
	if orgIDStr == "" {
		return nil
	}

	id, err := uuid.Parse(orgIDStr)
	if err != nil {
		return nil
	}
	return &id
}

// CreateDepartment creates a new department
func (h *DepartmentHandler) CreateDepartment(ctx context.Context, req *departmentpb.CreateDepartmentRequest) (*departmentpb.DepartmentResponse, error) {
	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "department name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "department description is required")
	}

	orgID := getOrgIDFromContext(ctx)
	// Optionally enforce orgID requirement?
	// User said "org_id is not fetching... and not storing". So we should store it.
	// We pass it to service (which passes to repo).

	// Create department
	dept, err := h.service.CreateDepartment(ctx, req.Name, req.Description, orgID)
	if err != nil {
		return nil, handleError(err)
	}

	return &departmentpb.DepartmentResponse{
		Department: domainToProto(dept),
	}, nil
}

// GetDepartment retrieves a department by ID
func (h *DepartmentHandler) GetDepartment(ctx context.Context, req *departmentpb.GetDepartmentRequest) (*departmentpb.DepartmentResponse, error) {
	// Parse ID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid department ID")
	}

	// Get department
	dept, err := h.service.GetDepartment(ctx, id)
	if err != nil {
		return nil, handleError(err)
	}

	return &departmentpb.DepartmentResponse{
		Department: domainToProto(dept),
	}, nil
}

// UpdateDepartment updates a department
func (h *DepartmentHandler) UpdateDepartment(ctx context.Context, req *departmentpb.UpdateDepartmentRequest) (*departmentpb.DepartmentResponse, error) {
	// Parse ID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid department ID")
	}

	// Validate request
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "department name is required")
	}
	if req.Description == "" {
		return nil, status.Error(codes.InvalidArgument, "department description is required")
	}

	// Update department
	dept, err := h.service.UpdateDepartment(ctx, id, req.Name, req.Description)
	if err != nil {
		return nil, handleError(err)
	}

	return &departmentpb.DepartmentResponse{
		Department: domainToProto(dept),
	}, nil
}

// DeleteDepartment deletes a department
func (h *DepartmentHandler) DeleteDepartment(ctx context.Context, req *departmentpb.DeleteDepartmentRequest) (*departmentpb.DeleteDepartmentResponse, error) {
	// Parse ID
	id, err := uuid.Parse(req.Id)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "invalid department ID")
	}

	// Delete department
	if err := h.service.DeleteDepartment(ctx, id); err != nil {
		return nil, handleError(err)
	}

	return &departmentpb.DeleteDepartmentResponse{
		Success: true,
		Message: "Department deleted successfully",
	}, nil
}

// ListDepartments lists departments with pagination
func (h *DepartmentHandler) ListDepartments(ctx context.Context, req *departmentpb.ListDepartmentsRequest) (*departmentpb.ListDepartmentsResponse, error) {
	// Set defaults
	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 10
	}

	orgID := getOrgIDFromContext(ctx)

	// List departments
	departments, total, err := h.service.ListDepartments(ctx, orgID, page, pageSize)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to list departments")
	}

	// Convert to proto
	protoDepts := make([]*departmentpb.Department, len(departments))
	for i, dept := range departments {
		protoDepts[i] = domainToProto(dept)
	}

	// Calculate total pages
	totalPages := (total + int32(pageSize) - 1) / int32(pageSize)

	return &departmentpb.ListDepartmentsResponse{
		Departments: protoDepts,
		TotalCount:  total,
		Pagination: &departmentpb.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}, nil
}

// domainToProto converts domain model to proto model
func domainToProto(dept *domain.Department) *departmentpb.Department {
	return &departmentpb.Department{
		Id:          dept.ID.String(),
		Name:        dept.Name,
		Description: dept.Description,
		CreatedAt:   timestamppb.New(dept.CreatedAt),
		UpdatedAt:   timestamppb.New(dept.UpdatedAt),
	}
}

// handleError converts domain errors to gRPC status errors
func handleError(err error) error {
	switch err {
	case domain.ErrDepartmentNotFound:
		return status.Error(codes.NotFound, err.Error())
	case domain.ErrDepartmentAlreadyExists:
		return status.Error(codes.AlreadyExists, err.Error())
	case domain.ErrDepartmentHasUsers:
		return status.Error(codes.FailedPrecondition, err.Error())
	case domain.ErrInvalidDepartmentID:
		return status.Error(codes.InvalidArgument, err.Error())
	case domain.ErrDepartmentNameRequired,
		domain.ErrDepartmentNameTooLong,
		domain.ErrDepartmentDescriptionRequired,
		domain.ErrDepartmentDescriptionTooLong:
		return status.Error(codes.InvalidArgument, err.Error())
	default:
		return status.Error(codes.Internal, "internal server error")
	}
}
