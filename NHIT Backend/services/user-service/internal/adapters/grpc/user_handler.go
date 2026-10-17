package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	authpb "github.com/ShristiRnr/NHIT_Backend/api/pb/authpb"
	deptpb "github.com/ShristiRnr/NHIT_Backend/api/pb/departmentpb"
	desigpb "github.com/ShristiRnr/NHIT_Backend/api/pb/designationpb"
	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/storage"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"log/slog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"github.com/ShristiRnr/NHIT_Backend/pkg/middleware"
	"bytes"
)

type UserHandler struct {
	userpb.UnimplementedUserManagementServer
	userService ports.UserService
	db          *pgxpool.Pool
	authClient  authpb.AuthServiceClient
	deptConn    *grpc.ClientConn
	desigConn   *grpc.ClientConn
	minioClient *storage.MinIOClient
	logger      *slog.Logger
}

// NewUserHandler creates a new gRPC user handler
func NewUserHandler(userService ports.UserService, db *pgxpool.Pool, authClient authpb.AuthServiceClient, deptConn *grpc.ClientConn, desigConn *grpc.ClientConn, minioClient *storage.MinIOClient, logger *slog.Logger) *UserHandler {
	return &UserHandler{
		userService: userService,
		db:          db,
		authClient:  authClient,
		deptConn:    deptConn,
		desigConn:   desigConn,
		minioClient: minioClient,
		logger:      logger,
	}
}

type authContext struct {
	token *authpb.ValidateTokenResponse
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

func (h *UserHandler) requireAuthWithPermissions(ctx context.Context, requiredPerms ...string) (*authContext, error) {
	if h.authClient == nil {
		return nil, status.Error(codes.Internal, "auth client not configured for user handler")
	}

	md, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil, status.Error(codes.Unauthenticated, "metadata is not provided")
	}

	accessToken := firstMetadataValue(md, "authorization")
	if accessToken == "" {
		return nil, status.Error(codes.Unauthenticated, "authorization token is not provided")
	}
	if strings.HasPrefix(accessToken, "Bearer ") {
		accessToken = strings.TrimPrefix(accessToken, "Bearer ")
	}

	vResp, err := h.authClient.ValidateToken(ctx, &authpb.ValidateTokenRequest{Token: accessToken})
	if err != nil {
		return nil, status.Errorf(codes.Unauthenticated, "failed to validate token: %v", err)
	}
	if !vResp.Valid {
		return nil, status.Error(codes.Unauthenticated, "invalid or expired token")
	}

	isSuperAdmin := false
	for _, r := range vResp.Roles {
		if r == "SUPER_ADMIN" {
			isSuperAdmin = true
			break
		}
	}
	if isSuperAdmin || len(requiredPerms) == 0 {
		return &authContext{token: vResp}, nil
	}

	permSet := make(map[string]struct{}, len(vResp.Permissions))
	for _, p := range vResp.Permissions {
		permSet[p] = struct{}{}
	}
	for _, req := range requiredPerms {
		if _, ok := permSet[req]; ok {
			return &authContext{token: vResp}, nil
		}
	}

	return nil, status.Error(codes.PermissionDenied, "insufficient permissions")
}

// helper to build a simple role description based on permission keys
func buildRoleDescription(perms []string) string {
	if len(perms) == 0 {
		return ""
	}
	return "Permissions: " + strings.Join(perms, ", ")
}

func toPBRole(role *domain.Role) *userpb.RoleResponse {
	return &userpb.RoleResponse{
		RoleId:      role.RoleID.String(),
		TenantId:    role.TenantID.String(),
		Name:        role.Name,
		Permissions: role.Permissions,
	}
}

func toPBPermission(p *domain.Permission) *userpb.PermissionResponse {
	return &userpb.PermissionResponse{
		PermissionId:       p.PermissionID.String(),
		Name:               p.Name,
		Description:        p.Description,
		Module:             p.Module,
		Action:             p.Action,
		IsSystemPermission: p.IsSystemPermission,
	}
}

func (h *UserHandler) CreateUser(ctx context.Context, req *userpb.CreateUserRequest) (*userpb.UserResponse, error) {
	tenantIDStr, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantID, err := uuid.Parse(tenantIDStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid tenant_id in context: %v", err)
	}

	// Parse and VALIDATE mandatory department_id
	var departmentID *uuid.UUID
	if req.DepartmentId == "" {
		return nil, status.Error(codes.InvalidArgument, "department_id is required")
	}
	deptID, err := uuid.Parse(req.DepartmentId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid department_id: %v", err)
	}
	departmentID = &deptID

	// Parse and VALIDATE mandatory designation_id
	var designationID *uuid.UUID
	if req.DesignationId == "" {
		return nil, status.Error(codes.InvalidArgument, "designation_id is required")
	}
	desigID, err := uuid.Parse(req.DesignationId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid designation_id: %v", err)
	}
	designationID = &desigID

	// VALIDATE mandatory role_id
	if req.RoleId == "" {
		return nil, status.Error(codes.InvalidArgument, "role_id is required")
	}
	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	// Parse banking information
	var accountHolderName, bankName, bankAccountNumber, ifscCode *string
	if req.AccountHolderName != "" {
		accountHolderName = &req.AccountHolderName
	}
	if req.BankName != "" {
		bankName = &req.BankName
	}
	if req.BankAccountNumber != "" {
		bankAccountNumber = &req.BankAccountNumber
	}
	if req.IfscCode != "" {
		ifscCode = &req.IfscCode
	}

	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		oid, err := uuid.Parse(orgIDStr)
		if err == nil {
			orgIDPtr = &oid
		}
	} else if req.OrgId != "" {
		// Allow specifying org_id in request if not in token (e.g. for super admins)
		oid, err := uuid.Parse(req.OrgId)
		if err == nil {
			orgIDPtr = &oid
		}
	}

	user := &domain.User{
		TenantID:          tenantID,
		OrgID:             orgIDPtr,
		Name:              req.Name,
		Email:             req.Email,
		Password:          req.Password,
		DepartmentID:      departmentID,
		DesignationID:     designationID,
		AccountHolderName: accountHolderName,
		BankName:          bankName,
		BankAccountNumber: bankAccountNumber,
		IFSCCode:          ifscCode,
	}

	createdUser, err := h.userService.CreateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create user: %v", err)
	}

	// Assign VALIDATED role to user
	if err := h.userService.AssignRoleToUser(ctx, createdUser.UserID, roleID); err != nil {
		// Note: User is created but role assignment failed. Ideally this should be a transaction.
		// For now, we return error so client knows something went wrong.
		return nil, status.Errorf(codes.Internal, "failed to assign role to user: %v", err)
	}

	// Link User to Organization
	// Priority: 1. req.OrgId 2. Token OrgId
	var targetOrgID uuid.UUID
	var hasOrgID bool

	if req.OrgId != "" {
		if parsed, err := uuid.Parse(req.OrgId); err == nil {
			targetOrgID = parsed
			hasOrgID = true
		}
	}

	// If not in request, try to extract from token
	if !hasOrgID {
		if orgIDPtr != nil {
			targetOrgID = *orgIDPtr
			hasOrgID = true
		}
	}

	if hasOrgID {
		// Insert into user_organizations
		// We use the same roleID as assigned above
		now := time.Now()
		_, err := h.db.Exec(ctx, `
			INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context, joined_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, org_id) DO NOTHING`, // Prevent error if already exists
			createdUser.UserID, targetOrgID, roleID, true, now, now,
		)
		if err != nil {
			h.logger.Warn("Failed to link user to organization", "user_id", createdUser.UserID, "org_id", targetOrgID, "error", err)
		}
	}

	// Fetch roles/permissions to return in response
	var roleNames []string
	var permissions []string
	
	roles, err := h.userService.GetUserRoles(ctx, createdUser.UserID)
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(createdUser, roleNames, permissions), nil
}

func (h *UserHandler) GetUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.UserResponse, error) {
	// Extract org_id from JWT metadata via context
	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user, err := h.userService.GetUser(ctx, userID, orgIDPtr)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	// Fetch roles/permissions
	roles, err := h.userService.GetUserRoles(ctx, user.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(user, roleNames, permissions), nil
}

func (h *UserHandler) UpdateUser(ctx context.Context, req *userpb.UpdateUserRequest) (*userpb.UserResponse, error) {
	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	user := &domain.User{
		UserID:   userID,
		OrgID:    orgIDPtr, // Pass org_id to ensure scoping
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		IsActive: req.IsActive,
	}

	if req.DepartmentId != "" {
		id, err := uuid.Parse(req.DepartmentId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid department_id: %v", err)
		}
		user.DepartmentID = &id
	}

	if req.DesignationId != "" {
		id, err := uuid.Parse(req.DesignationId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid designation_id: %v", err)
		}
		user.DesignationID = &id
	}

	// Handle role updates if provided
	if len(req.Roles) > 0 {
		var roleIDs []uuid.UUID
		for _, roleIDStr := range req.Roles {
			roleID, err := uuid.Parse(roleIDStr)
			if err != nil {
				return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
			}
			roleIDs = append(roleIDs, roleID)
		}
		user.Roles = roleIDs
	}

	updatedUser, err := h.userService.UpdateUser(ctx, user)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update user: %v", err)
	}

	// Fetch roles/permissions for response
	roles, err := h.userService.GetUserRoles(ctx, updatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(updatedUser, roleNames, permissions), nil
}

// Helper function to convert domain user to protobuf user with full details
func domainUserToProto(user *domain.User, roleNames []string, permissions []string) *userpb.UserResponse {
	resp := &userpb.UserResponse{
		UserId:      user.UserID.String(),
		Name:        user.Name,
		Email:       user.Email,
		Roles:       roleNames,
		Permissions: permissions,
		IsActive:    user.IsActive,
	}
	if user.DepartmentID != nil {
		resp.DepartmentId = user.DepartmentID.String()
	}
	if user.DesignationID != nil {
		resp.DesignationId = user.DesignationID.String()
	}
	if user.SignatureURL != nil {
		resp.SignatureUrl = *user.SignatureURL
	}
	return resp
}

func (h *UserHandler) DeleteUser(ctx context.Context, req *userpb.DeleteUserRequest) (*emptypb.Empty, error) {
	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	if err := h.userService.DeleteUser(ctx, userID, orgIDPtr); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete user: %v", err)
	}

	return &emptypb.Empty{}, nil
}

// DeactivateUser deactivates a user (soft delete)
func (h *UserHandler) DeactivateUser(ctx context.Context, req *userpb.DeactivateUserRequest) (*userpb.UserResponse, error) {
	// Parse user_id to deactivate
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Extract deactivated_by from context (the logged-in admin)
	deactivatedByStr, ok := ctx.Value("user_id").(string)
	if !ok || deactivatedByStr == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	deactivatedBy, err := uuid.Parse(deactivatedByStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid user_id in context: %v", err)
	}

	// Call service to deactivate user
	deactivatedUser, err := h.userService.DeactivateUser(ctx, userID, deactivatedBy, req.Reason)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to deactivate user: %v", err)
	}

	// Fetch keys for response
	roles, err := h.userService.GetUserRoles(ctx, deactivatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(deactivatedUser, roleNames, permissions), nil
}

// ReactivateUser reactivates a previously deactivated user
func (h *UserHandler) ReactivateUser(ctx context.Context, req *userpb.ReactivateUserRequest) (*userpb.UserResponse, error) {
	// Parse user_id to reactivate
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	// Extract reactivated_by from context (the logged-in admin)
	reactivatedByStr, ok := ctx.Value("user_id").(string)
	if !ok || reactivatedByStr == "" {
		return nil, status.Error(codes.Unauthenticated, "user_id not found in context")
	}

	reactivatedBy, err := uuid.Parse(reactivatedByStr)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid user_id in context: %v", err)
	}

	// Call service to reactivate user
	reactivatedUser, err := h.userService.ReactivateUser(ctx, userID, reactivatedBy)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to reactivate user: %v", err)
	}

	// Fetch keys for response
	roles, err := h.userService.GetUserRoles(ctx, reactivatedUser.UserID)
	var roleNames []string
	var permissions []string
	if err == nil {
		permMap := make(map[string]bool)
		for _, r := range roles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
	}

	return domainUserToProto(reactivatedUser, roleNames, permissions), nil
}

func (h *UserHandler) ListUsers(ctx context.Context, req *userpb.ListUsersRequest) (*userpb.ListUsersResponse, error) {
	tenantIDStr, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantID, _ := uuid.Parse(tenantIDStr)

	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	}

	// Extract metadata for forwarding to other services
	md, _ := metadata.FromIncomingContext(ctx)

	// Extract pagination from flat fields
	var limit int32 = 10
	if req.PageSize > 0 {
		limit = req.PageSize
	}
	
	var offset int32 = 0
	if req.Page > 0 {
		offset = (req.Page - 1) * limit
	}
	
	users, total, err := h.userService.ListUsersByTenant(ctx, tenantID, orgIDPtr, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list users: %v", err)
	}
	
	pbUsers := make([]*userpb.User, len(users))
	for i, user := range users {
		// Fetch Roles
		roles, err := h.userService.GetUserRoles(ctx, user.UserID)
		var roleNames []string
		var permissions []string
		permMap := make(map[string]bool)

		if err == nil {
			for _, r := range roles {
				roleNames = append(roleNames, r.Name)
				for _, p := range r.Permissions {
					if !permMap[p] {
						permissions = append(permissions, p)
						permMap[p] = true
					}
				}
			}
		}


		// Fetch Department Name
		var deptName string
		if user.DepartmentID != nil {
			client := deptpb.NewDepartmentServiceClient(h.deptConn)
			// Forward metadata for auth
			outCtx := metadata.NewOutgoingContext(ctx, md)
			resp, err := client.GetDepartment(outCtx, &deptpb.GetDepartmentRequest{Id: user.DepartmentID.String()})
			if err == nil && resp.Department != nil {
				deptName = resp.Department.Name
			}
		}

		// Fetch Designation Name
		var desigName string
		if user.DesignationID != nil {
			client := desigpb.NewDesignationServiceClient(h.desigConn)
			// Forward metadata for auth
			outCtx := metadata.NewOutgoingContext(ctx, md)
			resp, err := client.GetDesignation(outCtx, &desigpb.GetDesignationRequest{Id: user.DesignationID.String()})
			if err == nil && resp.Designation != nil {
				desigName = resp.Designation.Name
			}
		}

		pbUsers[i] = toPBUserMessage(user, deptName, desigName, roleNames, permissions)
	}

	totalPages := (int32(total) + limit - 1) / limit

	return &userpb.ListUsersResponse{
		Users: pbUsers,
		Pagination: &userpb.PaginationMetadata{
			Page:       req.Page,
			PageSize:   limit,
			TotalItems: int32(total),
			TotalPages: totalPages,
		},
	}, nil
}

// ===== Role & Permission Management =====

func (h *UserHandler) CreateRole(ctx context.Context, req *userpb.CreateRoleRequest) (*userpb.RoleResponse, error) {
	authCtx, err := h.requireAuthWithPermissions(ctx, "create-role")
	if err != nil {
		return nil, err
	}

	tenantID, err := uuid.Parse(authCtx.token.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid tenant_id in token: %v", err)
	}

	// Determine organization context: prefer explicit header (current dashboard), fallback to token org_id
	var orgID *uuid.UUID
	md, _ := metadata.FromIncomingContext(ctx)
	if orgHeader := firstMetadataValue(md, "x-org-id", "org-id"); orgHeader != "" {
		id, err := uuid.Parse(orgHeader)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid org_id header: %v", err)
		}
		orgID = &id
	} else if authCtx.token.OrgId != "" {
		id, err := uuid.Parse(authCtx.token.OrgId)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "invalid org_id in token: %v", err)
		}
		orgID = &id
	}

	desc := buildRoleDescription(req.Permissions)

	role := &domain.Role{
		TenantID:    tenantID,
		OrgID:       orgID,
		Name:        req.Name,
		Description: desc,
		Permissions: req.Permissions,
		CreatedBy:   authCtx.token.Name,
	}

	created, err := h.userService.CreateRole(ctx, role)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create role: %v", err)
	}

	return toPBRole(created), nil
}

func (h *UserHandler) GetRole(ctx context.Context, req *userpb.GetRoleRequest) (*userpb.RoleResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "view-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	role, err := h.userService.GetRole(ctx, roleID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	return toPBRole(role), nil
}

func (h *UserHandler) ListRoles(ctx context.Context, req *userpb.ListRolesRequest) (*userpb.ListRolesResponse, error) {
	authCtx, err := h.requireAuthWithPermissions(ctx, "view-role")
	if err != nil {
		return nil, err
	}

	if authCtx.token.TenantId == "" {
		return nil, status.Error(codes.Internal, "tenant_id missing in auth token")
	}

	tenantID, err := uuid.Parse(authCtx.token.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "invalid tenant_id in token: %v", err)
	}

	// Extract pagination
	var limit int32 = 10
	if req.PageSize > 0 {
		limit = req.PageSize
	}
	var offset int32 = 0
	if req.Page > 0 {
		offset = (req.Page - 1) * limit
	}

	// Extract OrgID from request or token
	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	} else if req.OrgId != nil && *req.OrgId != "" {
		if parsed, err := uuid.Parse(*req.OrgId); err == nil {
			orgIDPtr = &parsed
		}
	}

	var roles []*domain.Role
	var total int64
	var listErr error

	if orgIDPtr != nil {
		// List roles for organization, including system roles
		roles, total, listErr = h.userService.ListRolesByOrganization(ctx, tenantID, orgIDPtr, true, limit, offset)
	} else {
		roles, total, listErr = h.userService.ListRolesByTenant(ctx, tenantID, limit, offset)
	}

	if listErr != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", listErr)
	}

	totalPages := (int32(total) + limit - 1) / limit

	resp := &userpb.ListRolesResponse{
		Roles: make([]*userpb.RoleResponse, len(roles)),
		Pagination: &userpb.PaginationMetadata{
			Page:       req.Page,
			PageSize:   limit,
			TotalItems: int32(total),
			TotalPages: totalPages,
		},
	}
	for i, r := range roles {
		resp.Roles[i] = toPBRole(r)
	}
	return resp, nil
}

func (h *UserHandler) ListRolesByOrganization(ctx context.Context, req *userpb.ListRolesByOrganizationRequest) (*userpb.ListRolesResponse, error) {
	tenantIDStr, err := middleware.GetTenantIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	tenantID, _ := uuid.Parse(tenantIDStr)

	var orgIDPtr *uuid.UUID
	if orgIDStr, ok := middleware.GetOrgIDFromContext(ctx); ok {
		if parsed, err := uuid.Parse(orgIDStr); err == nil {
			orgIDPtr = &parsed
		}
	} else if req.OrgId != "" {
		if parsed, err := uuid.Parse(req.OrgId); err == nil {
			orgIDPtr = &parsed
		}
	}

	// Extract pagination
	var limit int32 = 10
	if req.PageSize > 0 {
		limit = req.PageSize
	}
	var offset int32 = 0
	if req.Page > 0 {
		offset = (req.Page - 1) * limit
	}

	roles, total, err := h.userService.ListRolesByOrganization(ctx, tenantID, orgIDPtr, req.IncludeSystemRoles, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list roles: %v", err)
	}

	totalPages := (int32(total) + limit - 1) / limit

	resp := &userpb.ListRolesResponse{
		Roles: make([]*userpb.RoleResponse, len(roles)),
		Pagination: &userpb.PaginationMetadata{
			Page:       req.Page,
			PageSize:   limit,
			TotalItems: int32(total),
			TotalPages: totalPages,
		},
	}
	for i, r := range roles {
		resp.Roles[i] = toPBRole(r)
	}
	return resp, nil
}

func (h *UserHandler) UpdateRole(ctx context.Context, req *userpb.UpdateRoleRequest) (*userpb.RoleResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "edit-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	existing, err := h.userService.GetRole(ctx, roleID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "role not found: %v", err)
	}

	existing.Name = req.Name
	existing.Permissions = req.Permissions
	existing.Description = buildRoleDescription(req.Permissions)

	updated, err := h.userService.UpdateRole(ctx, existing)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to update role: %v", err)
	}

	return toPBRole(updated), nil
}

func (h *UserHandler) DeleteRole(ctx context.Context, req *userpb.DeleteRoleRequest) (*emptypb.Empty, error) {
	if _, err := h.requireAuthWithPermissions(ctx, "delete-role"); err != nil {
		return nil, err
	}

	roleID, err := uuid.Parse(req.RoleId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
	}

	if err := h.userService.DeleteRole(ctx, roleID); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to delete role: %v", err)
	}

	return &emptypb.Empty{}, nil
}

func (h *UserHandler) ListPermissions(ctx context.Context, req *userpb.ListPermissionsRequest) (*userpb.ListPermissionsResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx); err != nil {
		return nil, err
	}

	var module *string
	if req.Module != "" {
		m := req.Module
		module = &m
	}

	perms, err := h.userService.ListPermissions(ctx, module)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions: %v", err)
	}

	resp := &userpb.ListPermissionsResponse{Permissions: make([]*userpb.PermissionResponse, len(perms))}
	for i, p := range perms {
		resp.Permissions[i] = toPBPermission(p)
	}
	return resp, nil
}

func (h *UserHandler) GetPermissionsByModule(ctx context.Context, req *userpb.GetPermissionsByModuleRequest) (*userpb.ListPermissionsResponse, error) {
	if _, err := h.requireAuthWithPermissions(ctx); err != nil {
		return nil, err
	}

	module := req.Module
	perms, err := h.userService.ListPermissions(ctx, &module)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list permissions by module: %v", err)
	}

	resp := &userpb.ListPermissionsResponse{Permissions: make([]*userpb.PermissionResponse, len(perms))}
	for i, p := range perms {
		resp.Permissions[i] = toPBPermission(p)
	}
	return resp, nil
}

func (h *UserHandler) CreateCustomPermission(ctx context.Context, req *userpb.CreateCustomPermissionRequest) (*userpb.PermissionResponse, error) {
	return nil, status.Error(codes.Unimplemented, "custom permissions are not supported; use fixed catalog")
}

func (h *UserHandler) AssignRolesToUser(ctx context.Context, req *userpb.AssignRolesRequest) (*userpb.UserResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	for _, roleIDStr := range req.Roles {
		roleID, err := uuid.Parse(roleIDStr)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid role_id: %v", err)
		}

		if err := h.userService.AssignRoleToUser(ctx, userID, roleID); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to assign role: %v", err)
		}
	}

	user, err := h.userService.GetUser(ctx, userID, nil)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user: %v", err)
	}

	response := domainUserToProto(user, nil, nil)
	
	freshRoles, err := h.userService.GetUserRoles(ctx, userID)
	if err == nil {
		var roleNames []string
		var permissions []string
		permMap := make(map[string]bool)
		for _, r := range freshRoles {
			roleNames = append(roleNames, r.Name)
			for _, p := range r.Permissions {
				if !permMap[p] {
					permissions = append(permissions, p)
					permMap[p] = true
				}
			}
		}
		response.Roles = roleNames
		response.Permissions = permissions
	}

	return response, nil
}

func (h *UserHandler) ListRolesOfUser(ctx context.Context, req *userpb.GetUserRequest) (*userpb.ListRolesResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	roles, err := h.userService.GetUserRoles(ctx, userID)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user roles: %v", err)
	}

	pbRoles := make([]*userpb.RoleResponse, len(roles))
	for i, role := range roles {
		pbRoles[i] = &userpb.RoleResponse{
			RoleId:      role.RoleID.String(),
			TenantId:    role.TenantID.String(),
			Name:        role.Name,
			Permissions: role.Permissions,
		}
	}

	return &userpb.ListRolesResponse{Roles: pbRoles}, nil
}

func (h *UserHandler) CreateTenant(ctx context.Context, req *userpb.CreateTenantRequest) (*userpb.TenantResponse, error) {
	tenant, err := h.userService.CreateTenant(ctx, req.Name, req.Email, req.Password, "SUPER_ADMIN")
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create tenant: %v", err)
	}

	return &userpb.TenantResponse{
		TenantId: tenant.TenantID.String(),
		Name:     tenant.Name,
		Email:    tenant.Email,
		Password: "",
	}, nil
}

func (h *UserHandler) GetTenant(ctx context.Context, req *userpb.GetTenantRequest) (*userpb.TenantResponse, error) {
	tenantID, err := uuid.Parse(req.TenantId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid tenant_id: %v", err)
	}

	tenant, err := h.userService.GetTenant(ctx, tenantID)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "tenant not found: %v", err)
	}

	return &userpb.TenantResponse{
		TenantId: tenant.TenantID.String(),
		Name:     tenant.Name,
		Email:    tenant.Email,
		Password: "",
	}, nil
}

func (h *UserHandler) ListUserOrganizations(ctx context.Context, req *userpb.ListUserOrganizationsRequest) (*userpb.ListUserOrganizationsResponse, error) {
	if h.db == nil {
		return nil, status.Error(codes.Internal, "database not configured for user organizations")
	}

	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	fetch := func() ([]*userpb.UserOrganizationInfo, error) {
		rows, err := h.db.Query(ctx, `
			SELECT uo.org_id, o.name, uo.is_current_context, uo.joined_at
			FROM user_organizations uo
			JOIN organizations o ON o.org_id = uo.org_id
			WHERE uo.user_id = $1
			ORDER BY uo.joined_at ASC`, userID)
		if err != nil {
			return nil, err
		}
		defer rows.Close()

		var result []*userpb.UserOrganizationInfo
		for rows.Next() {
			var (
				orgID     uuid.UUID
				orgName   string
				isCurrent bool
				joinedAt  time.Time
			)
			if err := rows.Scan(&orgID, &orgName, &isCurrent, &joinedAt); err != nil {
				return nil, err
			}
			result = append(result, &userpb.UserOrganizationInfo{
				OrgId:            orgID.String(),
				OrgName:          orgName,
				IsCurrentContext: isCurrent,
				JoinedAt:         timestamppb.New(joinedAt),
			})
		}
		return result, nil
	}

	orgInfos, err := fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user organizations: %v", err)
	}
	if len(orgInfos) > 0 {
		return &userpb.ListUserOrganizationsResponse{Organizations: orgInfos}, nil
	}

	var email string
	if err := h.db.QueryRow(ctx, `SELECT email FROM users WHERE user_id = $1`, userID).Scan(&email); err != nil {
		return nil, status.Errorf(codes.Internal, "failed to get user for organization lookup: %v", err)
	}

	orgRows, err := h.db.Query(ctx, `
		SELECT org_id, name
		FROM organizations
		WHERE super_admin_email = $1
		ORDER BY created_at ASC`, email)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to find organizations for user: %v", err)
	}
	defer orgRows.Close()

	var orgs []struct{id uuid.UUID; name string}
	for orgRows.Next() {
		var o struct{id uuid.UUID; name string}
		if err := orgRows.Scan(&o.id, &o.name); err != nil {
			return nil, status.Errorf(codes.Internal, "failed to scan organization: %v", err)
		}
		orgs = append(orgs, o)
	}

	if len(orgs) == 0 {
		return &userpb.ListUserOrganizationsResponse{Organizations: nil}, nil
	}

	now := time.Now()
	for i, o := range orgs {
		isCurrent := i == 0
		_, err := h.db.Exec(ctx, `
			INSERT INTO user_organizations (user_id, org_id, role_id, is_current_context, joined_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6)
			ON CONFLICT (user_id, org_id) DO NOTHING`,
			userID, o.id, uuid.Nil, isCurrent, now, now,
		)
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to link user to organization: %v", err)
		}
	}

	orgInfos, err = fetch()
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list user organizations after linking: %v", err)
	}

	return &userpb.ListUserOrganizationsResponse{Organizations: orgInfos}, nil
}

func (h *UserHandler) CreateUserLoginHistory(ctx context.Context, req *userpb.CreateUserLoginHistoryRequest) (*userpb.UserLoginHistoryResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	history := &domain.UserLoginHistory{
		UserID:    userID,
		IPAddress: &req.IpAddress,
		UserAgent: &req.UserAgent,
		LoginTime: time.Now(),
	}

	created, err := h.userService.CreateLoginHistory(ctx, history)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create login history: %v", err)
	}

	return &userpb.UserLoginHistoryResponse{
		HistoryId: created.HistoryID.String(),
		UserId:    created.UserID.String(),
		IpAddress: *created.IPAddress,
		UserAgent: *created.UserAgent,
		LoginTime: timestamppb.New(created.LoginTime),
	}, nil
}

func (h *UserHandler) ListUserLoginHistories(ctx context.Context, req *userpb.ListUserLoginHistoriesRequest) (*userpb.ListUserLoginHistoriesResponse, error) {
	var userID uuid.UUID
	var err error

	if req.UserId == "" {
		userIDStr, err := middleware.GetUserIDFromContext(ctx)
		if err != nil {
			return nil, err
		}
		userID, _ = uuid.Parse(userIDStr)
	} else {
		userID, err = uuid.Parse(req.UserId)
		if err != nil {
			return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
		}
	}

	var limit, offset int32 = 10, 0
	if req.Page != nil {
		limit = req.Page.PageSize
		offset = (req.Page.Page - 1) * req.Page.PageSize
	}

	histories, total, err := h.userService.ListLoginHistory(ctx, userID, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list login histories: %v", err)
	}

	pbHistories := make([]*userpb.UserLoginHistoryResponse, len(histories))
	for i, hist := range histories {
		pbHistories[i] = &userpb.UserLoginHistoryResponse{
			HistoryId: hist.HistoryID.String(),
			UserId:    hist.UserID.String(),
			IpAddress: *hist.IPAddress,
			UserAgent: *hist.UserAgent,
			LoginTime: timestamppb.New(hist.LoginTime),
		}
	}

	page := int32(1)
	if req.Page != nil {
		page = req.Page.Page
	}
	totalPages := (int32(total) + limit - 1) / limit

	return &userpb.ListUserLoginHistoriesResponse{
		Histories: pbHistories,
		Pagination: &userpb.PaginationMetadata{
			Page:       page,
			PageSize:   limit,
			TotalItems: int32(total),
			TotalPages: totalPages,
		},
	}, nil
}

func toPBUserMessage(user *domain.User, deptName, desigName string, roleNames, permissions []string) *userpb.User {
	return &userpb.User{
		UserId:          user.UserID.String(),
		Name:            user.Name,
		Email:           user.Email,
		DepartmentName:  deptName,
		DesignationName: desigName,
		Roles:           roleNames,
		Permissions:     permissions,
		IsActive:          user.IsActive,
		DeactivatedAt:     timestamppb.New(safeTime(user.DeactivatedAt)),
		DeactivatedBy:     safeUUIDStr(user.DeactivatedBy),
		DeactivatedByName: stringPtrValue(user.DeactivatedByName),
		CreatedAt:         timestamppb.New(user.CreatedAt),
		UpdatedAt:         timestamppb.New(user.UpdatedAt),
		EmailVerifiedAt:   timestamppb.New(safeTime(user.EmailVerifiedAt)),
		LastLoginAt:       timestamppb.New(safeTime(user.LastLoginAt)),
		LastLogoutAt:      timestamppb.New(safeTime(user.LastLogoutAt)),
		LastLoginIp:       user.LastLoginIP,
		UserAgent:         user.UserAgent,
		SignatureUrl:      stringPtrValue(user.SignatureURL),
	}
}

func safeTime(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

func safeUUIDStr(u *uuid.UUID) string {
	if u == nil {
		return ""
	}
	return u.String()
}

func (h *UserHandler) CreateActivityLog(ctx context.Context, req *userpb.CreateActivityLogRequest) (*userpb.ActivityLogResponse, error) {
	log := &domain.ActivityLog{
		Name:        req.Name,
		Description: req.Description,
	}

	created, err := h.userService.CreateActivityLog(ctx, log)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to create activity log: %v", err)
	}

	return &userpb.ActivityLogResponse{
		Id:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		CreatedAt:   timestamppb.New(created.CreatedAt),
	}, nil
}

func (h *UserHandler) ListActivityLogs(ctx context.Context, req *userpb.ListActivityLogsRequest) (*userpb.ListActivityLogsResponse, error) {
	var limit, offset int32 = 10, 0
	page := int32(1)
	if req.Page != nil {
		limit = req.Page.PageSize
		page = req.Page.Page
		offset = (page - 1) * limit
	}

	logs, total, err := h.userService.ListActivityLogs(ctx, limit, offset)
	if err != nil {
		return nil, status.Errorf(codes.Internal, "failed to list activity logs: %v", err)
	}

	pbLogs := make([]*userpb.ActivityLog, len(logs))
	for i, log := range logs {
		pbLogs[i] = &userpb.ActivityLog{
			Id:          log.ID,
			Name:        log.Name,
			Description: log.Description,
			CreatedAt:   timestamppb.New(log.CreatedAt),
		}
	}

	totalPages := (int32(total) + limit - 1) / limit

	return &userpb.ListActivityLogsResponse{
		Logs: pbLogs,
		Pagination: &userpb.PaginationMetadata{
			Page:       page,
			PageSize:   limit,
			TotalItems: int32(total),
			TotalPages: totalPages,
		},
	}, nil
}

func stringPtrValue(ptr *string) string {
	if ptr != nil {
		return *ptr
	}
	return ""
}

func (h *UserHandler) UploadUserSignature(ctx context.Context, req *userpb.UploadSignatureRequest) (*userpb.UploadSignatureResponse, error) {
	userID, err := uuid.Parse(req.UserId)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid user_id: %v", err)
	}

	var signatureURL string

	// 1. Check if we already have a URL (from Gateway upload)
	if req.SignatureUrl != "" {
		signatureURL = req.SignatureUrl
		h.logger.Info("Signature URL received from Gateway", "url", signatureURL)
	} else {
		// 2. Fallback: Perform upload if file bytes are provided
		if h.minioClient == nil {
			return nil, status.Error(codes.FailedPrecondition, "MinIO client not initialized")
		}

		if len(req.SignatureFile) == 0 {
			return nil, status.Error(codes.InvalidArgument, "signature file or URL is required")
		}

		user, err := h.userService.GetUser(ctx, userID, nil)
		if err != nil {
			return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
		}

		filename := req.Filename
		if filename == "" {
			filename = "signature.jpg"
		}

		reader := bytes.NewReader(req.SignatureFile)
		size := int64(len(req.SignatureFile))

		// Sanitize User Name for path
		userName := "unknown_user"
		if user.Name != "" {
			userName = strings.Map(func(r rune) rune {
				if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
					return r
				}
				return '_'
			}, user.Name)
		}

		// Path: UserName/timestamp_filename
		objectName := fmt.Sprintf("%s/%d_%s", userName, time.Now().Unix(), filename)

		signatureURL, err = h.minioClient.UploadFile(ctx, objectName, reader, size, "image/jpeg")
		if err != nil {
			return nil, status.Errorf(codes.Internal, "failed to upload signature: %v", err)
		}
		h.logger.Info("Signature uploaded to MinIO", "url", signatureURL)
	}

	// 3. Update the user record with the signature URL
	h.logger.Debug("Fetching user to update signature", "user_id", userID)
	user, err := h.userService.GetUser(ctx, userID, nil)
	if err != nil {
		return nil, status.Errorf(codes.NotFound, "user not found: %v", err)
	}

	h.logger.Debug("Setting signature_url", "url", signatureURL)
	user.SignatureURL = &signatureURL
	
	h.logger.Debug("Calling UpdateUser with signature_url", "url", *user.SignatureURL)
	updatedUser, err := h.userService.UpdateUser(ctx, user)
	if err != nil {
		h.logger.Error("Failed to update user signature", "error", err)
		return nil, status.Errorf(codes.Internal, "failed to update user signature URL: %v", err)
	}

	if updatedUser.SignatureURL != nil {
		h.logger.Info("User signature URL updated successfully", "url", *updatedUser.SignatureURL)
	} else {
		h.logger.Warn("Updated user has nil signature_url")
	}

	return &userpb.UploadSignatureResponse{
		SignatureUrl: signatureURL,
		Message:      "Signature updated successfully",
	}, nil
}
