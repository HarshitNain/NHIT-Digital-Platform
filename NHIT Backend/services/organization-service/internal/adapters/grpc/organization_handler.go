package grpc

import (
	"context"
	"fmt"
	"bytes"
	"log/slog"
	"strings"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	pb "github.com/ShristiRnr/NHIT_Backend/api/pb/organizationpb"
	projectpb "github.com/ShristiRnr/NHIT_Backend/api/pb/projectpb"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/organization-service/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// OrganizationHandler implements the gRPC server and uses the Repository port.
type OrganizationHandler struct {
	pb.UnimplementedOrganizationServiceServer
	repo       ports.Repository
	db         *pgxpool.Pool
	authClient authpb.AuthServiceClient
	kafka      ports.KafkaPublisher
	logger     *slog.Logger
	minioClient *storage.MinIOClient
	// optional: clock or logger
}

// NewOrganizationHandler constructor
func NewOrganizationHandler(repo ports.Repository, db *pgxpool.Pool, authClient authpb.AuthServiceClient, kafka ports.KafkaPublisher, minioClient *storage.MinIOClient, logger *slog.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		repo:        repo,
		db:          db,
		authClient:  authClient,
		kafka:       kafka,
		logger:      logger,
		minioClient: minioClient,
	}
}

// helper: map repo model -> pb.Organization
func mapModelToProto(m ports.OrganizationModel) *pb.Organization {
	var parentID string
	if m.ParentOrgID != nil {
		parentID = *m.ParentOrgID
	}
	var desc, logo string
	if m.Description != nil {
		desc = *m.Description
	}
	if m.Logo != nil {
		logo = *m.Logo
	}
	isParent := m.ParentOrgID == nil
	var superAdmin *pb.SuperAdminDetails
	if isParent && (m.SuperAdminName != nil || m.SuperAdminEmail != nil) {
		superAdmin = &pb.SuperAdminDetails{
			Name:     safeStr(m.SuperAdminName),
			Email:    safeStr(m.SuperAdminEmail),
			Password: "", // do not return password
		}
	}
	createdBy := ""
	if m.SuperAdminName != nil {
		createdBy = safeStr(m.SuperAdminName)
	}
	return &pb.Organization{
		OrgId:           m.OrgID,
		TenantId:        m.TenantID,
		ParentOrgId:     parentID,
		Name:            m.Name,
		Code:            m.Code,
		DatabaseName:    m.DatabaseName,
		Description:     desc,
		Logo:            logo,
		SuperAdmin:      superAdmin,
		InitialProjects: []string{}, // Empty array since domain model doesn't have this field
		Status:          m.Status,
		CreatedAt:       toProtoTs(m.CreatedAt),
		UpdatedAt:       toProtoTs(m.UpdatedAt),
		CreatedBy:       createdBy,
	}
}

func safeStr(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// helper: to proto timestamp
func toProtoTs(t time.Time) *timestamppb.Timestamp {
	// pb.Timestamp refers to google.protobuf.Timestamp — but in generated pb it's google.protobuf.Timestamp
	// For brevity, we use zero; in real code use timestamppb.New(t)
	return timestamppb.New(t)
}

// ----------------------------------------------------------------------------
// CreateOrganization
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) CreateOrganization(ctx context.Context, req *pb.CreateOrganizationRequest) (*pb.OrganizationResponse, error) {
	// validation rules:
	// - if parent_org_id == "" → parent creation -> super_admin must be present
	// - if parent_org_id != "" → child creation -> super_admin must be nil / ignored

	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if req.Name == "" || req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "name and code are required")
	}

	// Determine parent organization and tenant context
	// Case 1: Explicit parentOrgId provided -> child under that parent (existing behavior)
	// Case 2: No parentOrgId but SuperAdmin provided -> parent org creation
	// Case 3: No parentOrgId and no SuperAdmin -> child under current org from JWT (Option B)

	orgID := uuid.New().String()
	now := time.Now().UTC()

	var tenantID string
	var parentPtr *string
	var saName, saEmail, saPass *string
	var createdBy string

	if req.ParentOrgId != "" {
		// Explicit child creation: validate parent and inherit tenant and super admin from parent org
		if _, err := uuid.Parse(req.ParentOrgId); err != nil {
			return nil, status.Error(codes.InvalidArgument, "parent_org_id must be a valid UUID")
		}
		parentOrg, err := h.repo.GetOrganizationByID(ctx, req.ParentOrgId)
		if err != nil {
			return nil, status.Error(codes.InvalidArgument, "parent organization not found")
		}
		tenantID = parentOrg.TenantID
		parentPtr = &req.ParentOrgId
		// Copy parent super admin details into the child organization for consistent createdBy behavior
		saName = parentOrg.SuperAdminName
		saEmail = parentOrg.SuperAdminEmail
		saPass = parentOrg.SuperAdminPass
	} else if req.SuperAdmin != nil {
		// Parent org creation: super admin is required and tenant inferred via super_admin.email
		if req.SuperAdmin.Email == "" {
			return nil, status.Error(codes.InvalidArgument, "super_admin.email is required when creating parent organization")
		}
		if h.db == nil {
			return nil, status.Error(codes.Internal, "database connection not configured for organization handler")
		}

		var tenantUUID uuid.UUID
		if err := h.db.QueryRow(ctx, `SELECT tenant_id FROM tenants WHERE email = $1`, req.SuperAdmin.Email).Scan(&tenantUUID); err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "tenant not found for super admin email %s; create tenant first: %v", req.SuperAdmin.Email, err)
		}
		tenantID = tenantUUID.String()
		// parentPtr remains nil for parent orgs
		saName = &req.SuperAdmin.Name
		saEmail = &req.SuperAdmin.Email
		saPass = &req.SuperAdmin.Password
	} else {
		// Token-based child creation: no explicit parentOrgId, no superAdmin -> use current org from JWT via AuthService
		if h.authClient == nil {
			return nil, status.Error(codes.Internal, "auth client not configured for organization handler")
		}

		md, ok := metadata.FromIncomingContext(ctx)
		if !ok {
			return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
		}
		values := md["authorization"]
		if len(values) == 0 {
			return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
		}

		accessToken := strings.TrimPrefix(values[0], "Bearer ")

		vResp, err := h.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: accessToken})
		if err != nil {
			return nil, status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
		}
		if !vResp.Valid {
			return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
		}
		if vResp.OrgId == "" {
			return nil, status.Error(codes.InvalidArgument, "parent_org_id is missing and token has no org_id")
		}

		// Ensure parent org exists
		parentOrg, err := h.repo.GetOrganizationByID(ctx, vResp.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "parent organization from token not found: %v", err)
		}
		parentID := parentOrg.OrgID
		parentPtr = &parentID
		// Use tenant_id from JWT so child org stays in the same tenant context as the logged-in user
		tenantID = vResp.TenantId
		// Copy parent super admin details into the child organization as well
		saName = parentOrg.SuperAdminName
		saEmail = parentOrg.SuperAdminEmail
		saPass = parentOrg.SuperAdminPass
	}

	// Resolve tenant name for created_by display field using tenantID
	// Default to "superadmin" if lookup fails or name is empty
	if h.db != nil && tenantID != "" {
		if err := h.db.QueryRow(ctx, `SELECT name FROM tenants WHERE tenant_id = $1`, tenantID).Scan(&createdBy); err != nil {
			createdBy = "superadmin"
		}
	}
	if createdBy == "" {
		createdBy = "superadmin"
	}

	var descPtr, logoPtr *string
	if req.Description != "" {
		descPtr = &req.Description
	}
	if req.Logo != "" {
		logoPtr = &req.Logo
	}

	// Generate database_name (simple example using code + timestamp)
	dbName := fmt.Sprintf("%s_db", req.Code)

	repoModel := ports.OrganizationModel{
		OrgID:           orgID,
		TenantID:        tenantID,
		ParentOrgID:     parentPtr,
		Name:            req.Name,
		Code:            req.Code,
		DatabaseName:    dbName,
		Description:     descPtr,
		Logo:            logoPtr,
		SuperAdminName:  saName,
		SuperAdminEmail: saEmail,
		SuperAdminPass:  saPass,
		InitialProjects: req.InitialProjects,
		Status:          req.Status,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Extract metadata (auth token) from incoming context to propagate to project service
	var pairs []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if auth := md.Get("authorization"); len(auth) > 0 {
			pairs = append(pairs, "authorization", auth[0])
		}
		if tenantIDs := md.Get("tenant_id"); len(tenantIDs) > 0 {
			pairs = append(pairs, "tenant_id", tenantIDs[0])
		}
		if userIDs := md.Get("user_id"); len(userIDs) > 0 {
			pairs = append(pairs, "user_id", userIDs[0])
		}
		if role := md.Get("role"); len(role) > 0 {
			pairs = append(pairs, "role", role[0])
		}
	}

	// Call repository to persist
	created, err := h.repo.CreateOrganization(ctx, repoModel)
	if err != nil {
		// translate to gRPC error (simplified)
		return nil, status.Errorf(codes.Internal, "failed to create organization: %v", err)
	}

	// Synchronously create projects via gRPC to ensure immediate availability
	if len(req.InitialProjects) > 0 {
		h.logger.Info("Found initial projects to create", "count", len(req.InitialProjects), "projects", req.InitialProjects)
		go func() {
			// Using background context to not block response if it takes long,
			// but propagating the auth metadata for RBAC checks in project service
			projectCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			
			// Attach metadata to outgoing context
			if len(pairs) > 0 {
				projectCtx = metadata.NewOutgoingContext(projectCtx, metadata.Pairs(pairs...))
			}
			
			h.logger.Debug("Calling createProjectsSync", "projects", req.InitialProjects)
			if err := h.createProjectsSync(projectCtx, created.TenantID, created.OrgID, req.InitialProjects, createdBy); err != nil {
				h.logger.Error("Failed to create initial projects synchronously", "error", err)
			} else {
				h.logger.Info("Successfully created initial projects synchronously")
			}
		}()
	} else {
		h.logger.Debug("No initial projects provided in request")
	}

	orgProto := mapModelToProto(created)
	orgProto.CreatedBy = createdBy
	// Manually populate InitialProjects in response since repository model might not persist it field (it's distinct table)
	orgProto.InitialProjects = req.InitialProjects

	return &pb.OrganizationResponse{
		Organization: orgProto,
		Message:      "organization created",
	}, nil
}

// createProjectsSync calls project-service to create projects
func (h *OrganizationHandler) createProjectsSync(ctx context.Context, tenantID, orgID string, projectNames []string, createdBy string) error {
	h.logger.Info("Connecting to project service for sync creation", "count", len(projectNames))
	conn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return fmt.Errorf("failed to connect to project service: %w", err)
	}
	defer conn.Close()

	client := projectpb.NewProjectServiceClient(conn)

	for _, name := range projectNames {
		h.logger.Debug("Sending CreateProject request", "project", name)
		resp, err := client.CreateProject(ctx, &projectpb.CreateProjectRequest{
			TenantId:    tenantID,
			OrgId:       orgID,
			ProjectName: name,
			CreatedBy:   createdBy,
		})
		if err != nil {
			h.logger.Error("Error creating project", "project", name, "error", err)
			// continue with other projects
		} else {
			h.logger.Info("Project created successfully", "project", name, "project_id", resp.Project.ProjectId)
		}
	}
	return nil
}

// ListOrganizations
// ----------------------------------------------------------------------------
// ListOrganizations lists all organizations with pagination
func (h *OrganizationHandler) ListOrganizations(ctx context.Context, req *pb.ListOrganizationsRequest) (*pb.ListOrganizationsResponse, error) {
	page := int32(1)
	pageSize := int32(10)

	if req.Page > 0 {
		page = req.Page
	}
	if req.PageSize > 0 && req.PageSize <= 100 {
		pageSize = req.PageSize
	}

	offset := (page - 1) * pageSize

	organizations, totalCount, err := h.repo.ListOrganizations(ctx, int(offset), int(pageSize))
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list organizations: %v", err)
	}

	var orgProtos []*pb.Organization
	for _, org := range organizations {
		orgProtos = append(orgProtos, mapModelToProto(org))
	}

	totalPages := (totalCount + int(pageSize) - 1) / int(pageSize)

	return &pb.ListOrganizationsResponse{
		Organizations: orgProtos,
		TotalCount:    int32(totalCount),
		Pagination: &pb.PaginationMetadata{
			CurrentPage: page,
			PageSize:    pageSize,
			TotalItems:  int32(totalCount),
			TotalPages:  int32(totalPages),
		},
	}, nil
}

// ListOrganizationsByTenant lists organizations for a specific tenant
// ListOrganizationsByTenant
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) ListOrganizationsByTenant(ctx context.Context, req *pb.ListOrganizationsByTenantRequest) (*pb.ListOrganizationsResponse, error) {
	if req.TenantId == "" {
		return nil, status.Error(codes.InvalidArgument, "tenant_id required")
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	h.logger.Debug("Listing organizations by tenant", "tenant_id", req.TenantId, "page", page, "page_size", pageSize, "offset", offset)

	orgs, total, err := h.repo.ListOrganizationsByTenant(ctx, req.TenantId, offset, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list orgs by tenant error: %v", err)
	}

	h.logger.Debug("Found organizations for tenant", "tenant_id", req.TenantId, "count", len(orgs), "total", total)

	pbOrgs := make([]*pb.Organization, 0, len(orgs))
	for _, o := range orgs {
		pbOrgs = append(pbOrgs, mapModelToProto(o))
	}

	return &pb.ListOrganizationsResponse{
		Organizations: pbOrgs,
		TotalCount:    int32(total),
		Pagination: &pb.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  int32(total),
			TotalPages:  int32((total + pageSize - 1) / pageSize),
		},
	}, nil
}

// ----------------------------------------------------------------------------
// ListChildOrganizations
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) ListChildOrganizations(ctx context.Context, req *pb.ListChildOrganizationsRequest) (*pb.ListOrganizationsResponse, error) {
	if req.ParentOrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "parent_org_id required")
	}
	page := int(req.Page)
	if page < 1 {
		page = 1
	}
	pageSize := int(req.PageSize)
	if pageSize <= 0 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	orgs, total, err := h.repo.ListChildOrganizations(ctx, req.ParentOrgId, offset, pageSize)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "list child orgs error: %v", err)
	}

	pbOrgs := make([]*pb.Organization, 0, len(orgs))
	for _, o := range orgs {
		pbOrgs = append(pbOrgs, mapModelToProto(o))
	}

	return &pb.ListOrganizationsResponse{
		Organizations: pbOrgs,
		TotalCount:    int32(total),
		Pagination: &pb.PaginationMetadata{
			CurrentPage: int32(page),
			PageSize:    int32(pageSize),
			TotalItems:  int32(total),
			TotalPages:  int32((total + pageSize - 1) / pageSize),
		},
	}, nil
}

// ----------------------------------------------------------------------------
// GetOrganization
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) GetOrganization(ctx context.Context, req *pb.GetOrganizationRequest) (*pb.OrganizationResponse, error) {
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	o, err := h.repo.GetOrganizationByID(ctx, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}
	return &pb.OrganizationResponse{
		Organization: mapModelToProto(o),
		Message:      "ok",
	}, nil
}

// ----------------------------------------------------------------------------
// GetOrganizationByCode
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) GetOrganizationByCode(ctx context.Context, req *pb.GetOrganizationByCodeRequest) (*pb.OrganizationResponse, error) {
	if req.Code == "" {
		return nil, status.Error(codes.InvalidArgument, "code required")
	}
	o, err := h.repo.GetOrganizationByCode(ctx, req.Code)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}
	return &pb.OrganizationResponse{
		Organization: mapModelToProto(o),
		Message:      "ok",
	}, nil
}

// ----------------------------------------------------------------------------
// UpdateOrganization
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) UpdateOrganization(ctx context.Context, req *pb.UpdateOrganizationRequest) (*pb.OrganizationResponse, error) {
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	// fetch existing
	existing, err := h.repo.GetOrganizationByID(ctx, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}

	// update mutable fields
	existing.Name = req.Name
	existing.Code = req.Code
	if req.Description != "" {
		d := req.Description
		existing.Description = &d
	}
	if req.Logo != "" {
		l := req.Logo
		existing.Logo = &l
	}
	existing.Status = req.Status
	existing.UpdatedAt = time.Now().UTC()

	updated, err := h.repo.UpdateOrganization(ctx, existing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update org: %v", err)
	}

	return &pb.OrganizationResponse{
		Organization: mapModelToProto(updated),
		Message:      "organization updated",
	}, nil
}

// GetOrganizationWithProjects retrieves an organization with its projects
func (h *OrganizationHandler) GetOrganizationWithProjects(ctx context.Context, req *pb.GetOrganizationWithProjectsRequest) (*pb.GetOrganizationWithProjectsResponse, error) {
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}

	// Get organization
	o, err := h.repo.GetOrganizationByID(ctx, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}

	orgProto := mapModelToProto(o)

	// Fetch projects from project service via gRPC call
	projects, err := h.fetchProjectsFromProjectService(ctx, req.OrgId)
	if err != nil {
		h.logger.Error("Failed to fetch projects from project service", "org_id", req.OrgId, "error", err)
		return nil, status.Errorf(codes.Internal, "failed to list projects: %v", err)
	}

	return &pb.GetOrganizationWithProjectsResponse{
		Organization: orgProto,
		Projects:     projects,
	}, nil
}

// fetchProjectsFromProjectService calls the project service to get projects for an organization
func (h *OrganizationHandler) fetchProjectsFromProjectService(ctx context.Context, orgID string) ([]*pb.Project, error) {
	// Extract incoming metadata (authorization, tenant_id, etc.)
	var pairs []string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		// Copy all relevant metadata keys
		for k, v := range md {
			for _, val := range v {
				pairs = append(pairs, k, val)
			}
		}
	}
	
	// Create outgoing context with the extracted metadata
	outCtx := metadata.NewOutgoingContext(ctx, metadata.Pairs(pairs...))

	// Connect to project service
	conn, err := grpc.Dial("localhost:50057", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to project service: %w", err)
	}
	defer conn.Close()

	client := projectpb.NewProjectServiceClient(conn)
	
	// Call ListProjectsByOrganization using the context with metadata
	resp, err := client.ListProjectsByOrganization(outCtx, &projectpb.ListProjectsByOrganizationRequest{OrgId: orgID})
	if err != nil {
		h.logger.Error("Failed to list projects from service", "org_id", orgID, "error", err)
		return nil, fmt.Errorf("failed to list projects: %w", err)
	}
	
	// Map projectpb.Project to organizationpb.Project
	// Note: proto definition of organizationpb.Project must match fields we want to return
	var projects []*pb.Project
	for _, p := range resp.Projects {
		projects = append(projects, &pb.Project{
			ProjectId:   p.ProjectId,
			TenantId:    p.TenantId,
			OrgId:       p.OrgId,
			ProjectName: p.ProjectName,
			CreatedBy:   p.CreatedBy,
			CreatedAt:   p.CreatedAt,
			UpdatedAt:   p.UpdatedAt,
		})
	}

	return projects, nil
}

// ----------------------------------------------------------------------------
// DeleteOrganization
// ----------------------------------------------------------------------------
func (h *OrganizationHandler) DeleteOrganization(ctx context.Context, req *pb.DeleteOrganizationRequest) (*pb.DeleteOrganizationResponse, error) {
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id required")
	}
	err := h.repo.DeleteOrganization(ctx, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete org: %v", err)
	}
	return &pb.DeleteOrganizationResponse{
		Success: true,
		Message: "organization deleted",
	}, nil
}

// UploadOrganizationLogo uploads an organization logo
func (h *OrganizationHandler) UploadOrganizationLogo(ctx context.Context, req *pb.UploadLogoRequest) (*pb.UploadLogoResponse, error) {
	if req.OrgId == "" {
		return nil, status.Error(codes.InvalidArgument, "org_id is required")
	}
	
	if len(req.FileContent) == 0 {
		return nil, status.Error(codes.InvalidArgument, "file_content is required")
	}

	// Fetch organization record FIRST to get name
	org, err := h.repo.GetOrganizationByID(ctx, req.OrgId)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "organization not found: %v", err)
	}

    // Sanitize Org Name
	orgName := "unknown_org"
	if org.Name != "" {
		orgName = strings.Map(func(r rune) rune {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
				return r
			}
			return '_'
		}, org.Name)
	}

	// Construct path: OrgName/timestamp_filename
	filename := fmt.Sprintf("%s/%d_logo_%s.png", orgName, time.Now().Unix(), req.OrgId)
	fileSize := int64(len(req.FileContent))
	fileReader := bytes.NewReader(req.FileContent)

	logoURL, err := h.minioClient.UploadFile(ctx, filename, fileReader, fileSize, "image/png")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to upload logo: %v", err)
	}

	org.Logo = &logoURL
	org.UpdatedAt = time.Now().UTC()

	_, err = h.repo.UpdateOrganization(ctx, org)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update organization logo: %v", err)
	}

	return &pb.UploadLogoResponse{
		Success:        true,
		Message:        "Logo uploaded successfully",
		OrganizationId: req.OrgId,
		LogoUrl:        logoURL,
		FileName:       filename,
		MimeType:       "image/png",
		FileSize:       fileSize,
		UploadedAt:     timestamppb.Now(),
	}, nil
}
