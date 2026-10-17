package services

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/ports"
	"github.com/google/uuid"
	"github.com/mileusna/useragent"
	"golang.org/x/crypto/bcrypt"
	"log/slog"
)

type userService struct {
	userRepo           ports.UserRepository
	tenantRepo         ports.TenantRepository
	userRoleRepo       ports.UserRoleRepository
	roleRepo           ports.RoleRepository
	permissionRepo     ports.PermissionRepository
	loginHistoryRepo   ports.LoginHistoryRepository
	activityLogRepo    ports.ActivityLogRepository
	logger             *slog.Logger
}

// NewUserService creates a new user service instance

func NewUserService(userRepo ports.UserRepository, tenantRepo ports.TenantRepository, userRoleRepo ports.UserRoleRepository, roleRepo ports.RoleRepository, permissionRepo ports.PermissionRepository, loginHistoryRepo ports.LoginHistoryRepository, activityLogRepo ports.ActivityLogRepository, logger *slog.Logger) ports.UserService {
	return &userService{
		userRepo:         userRepo,
		tenantRepo:       tenantRepo,
		userRoleRepo:     userRoleRepo,
		roleRepo:         roleRepo,
		permissionRepo:   permissionRepo,
		loginHistoryRepo: loginHistoryRepo,
		activityLogRepo:  activityLogRepo,
		logger:           logger,
	}
}

func (s *userService) CreateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	// Hash password before storing
	if user.Password != "" {
		hashedPassword, err := s.hashPassword(user.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	// Default to active if not specified
	user.IsActive = true
	
	// Auto-verify email for admin-created users
	now := time.Now()
	if user.EmailVerifiedAt == nil {
		user.EmailVerifiedAt = &now
	}

	// Create user in repository
	createdUser, err := s.userRepo.Create(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return createdUser, nil
}

func (s *userService) GetUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*domain.User, error) {
	user, err := s.userRepo.GetByID(ctx, userID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	return user, nil
}

func (s *userService) UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error) {
	existingUser, err := s.userRepo.GetByID(ctx, user.UserID, user.OrgID)
	if err != nil {
		return nil, fmt.Errorf("failed to get existing user: %w", err)
	}

	// --- Partial Update Checks ---
	if user.Name == "" {
		user.Name = existingUser.Name
	}
	if user.Email == "" {
		user.Email = existingUser.Email
	}
	// Password: If empty, keep existing. If provided, it will be hashed below.
	if user.Password == "" {
		user.Password = existingUser.Password
	}
	// Department: If nil, keep existing.
	if user.DepartmentID == nil {
		user.DepartmentID = existingUser.DepartmentID
	}
	// Designation: If nil, keep existing.
	if user.DesignationID == nil {
		user.DesignationID = existingUser.DesignationID
	}
	// Note: Account fields are not strictly checked here for nil as they are pointers in domain, 
	// but if the handler sends nil, we might blank them out if we strictly follow "replace" semantics.
	// However, user asked for "jab tak update na kiya jaye tab tak purana rahega".
	// Assuming the handler passes nil if not provided.
	if user.AccountHolderName == nil {
		user.AccountHolderName = existingUser.AccountHolderName
	}
	if user.BankName == nil {
		user.BankName = existingUser.BankName
	}
	if user.BankAccountNumber == nil {
		user.BankAccountNumber = existingUser.BankAccountNumber
	}
	if user.IFSCCode == nil {
		user.IFSCCode = existingUser.IFSCCode
	}
	if user.SignatureURL == nil {
		user.SignatureURL = existingUser.SignatureURL
	}

	// Handle Reactivation / Deactivation logic
	if user.IsActive {
		// Case: Active or Reactivating
		user.DeactivatedAt = nil
		user.DeactivatedBy = nil
		user.DeactivatedByName = nil
	} else {
		// Case: Inactive or Deactivating
		if existingUser.IsActive {
			now := time.Now()
			user.DeactivatedAt = &now
		} else {
			user.DeactivatedAt = existingUser.DeactivatedAt
			user.DeactivatedBy = existingUser.DeactivatedBy
			user.DeactivatedByName = existingUser.DeactivatedByName
		}
		// Force clear password if inactive? 
		// User said "password humesha same hi rhega", but deactivation usually implies blocking access.
		// Leaving as is for now, but usually we don't clear password on deactivation to allow reactivation.
		// Reverting the "Force clear password" logic from previous code to respect "password humesha same hi rhega".
		if user.Password == "" {
             user.Password = existingUser.Password
        }
	}

	// Hash password if it's being updated (differs from existing)
	if user.Password != "" && user.Password != existingUser.Password {
		hashedPassword, err := s.hashPassword(user.Password)
		if err != nil {
			return nil, fmt.Errorf("failed to hash password: %w", err)
		}
		user.Password = hashedPassword
	}

	var updatedUser *domain.User
	err = s.userRepo.WithTransaction(ctx, func(txCtx context.Context) error {
		// --- Role Management (Internal) ---
		if user.Roles != nil {
			// Fetch existing roles
			existingRoles, err := s.userRoleRepo.ListRolesByUser(txCtx, user.UserID)
			if err != nil {
				return fmt.Errorf("failed to fetch existing roles: %w", err)
			}

			existingRoleMap := make(map[uuid.UUID]bool)
			for _, r := range existingRoles {
				existingRoleMap[r.RoleID] = true
			}

			newRoleMap := make(map[uuid.UUID]bool)
			for _, rID := range user.Roles {
				newRoleMap[rID] = true
			}

			// Roles to Remove: Exist in DB but not in new list
			for rID := range existingRoleMap {
				if !newRoleMap[rID] {
					if err := s.userRoleRepo.RemoveRole(txCtx, user.UserID, rID); err != nil {
						return fmt.Errorf("failed to remove role %s: %w", rID, err)
					}
				}
			}

			// Roles to Add: Exist in new list but not in DB
			for rID := range newRoleMap {
				if !existingRoleMap[rID] {
					if err := s.userRoleRepo.AssignRole(txCtx, user.UserID, rID); err != nil {
						return fmt.Errorf("failed to assign role %s: %w", rID, err)
					}
				}
			}
		}

		var updateErr error
		updatedUser, updateErr = s.userRepo.Update(txCtx, user)
		if updateErr != nil {
			return fmt.Errorf("failed to update user: %w", updateErr)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return updatedUser, nil
}

func (s *userService) DeleteUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error {
	return s.userRepo.Delete(ctx, userID, orgID)
}

func (s *userService) ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.ListByTenant(ctx, tenantID, orgID, limit, offset)
}

func (s *userService) ListUsersByOrganization(ctx context.Context, tenantID, orgID uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.ListByTenant(ctx, tenantID, &orgID, limit, offset)
}

func (s *userService) ListUsersByDepartment(ctx context.Context, departmentID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.ListByDepartment(ctx, departmentID, orgID, limit, offset)
}

func (s *userService) ListUsersByDesignation(ctx context.Context, designationID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.ListByDesignation(ctx, designationID, orgID, limit, offset)
}

func (s *userService) SearchUsersByName(ctx context.Context, name string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.SearchByName(ctx, name, orgID, limit, offset)
}

func (s *userService) SearchUsersByEmail(ctx context.Context, email string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error) {
	return s.userRepo.SearchByEmail(ctx, email, orgID, limit, offset)
}

func (s *userService) AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error {
	return s.userRoleRepo.AssignRole(ctx, userID, roleID)
}

func (s *userService) GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error) {
	return s.userRoleRepo.ListRolesByUser(ctx, userID)
}

// ensureSuperAdminRoleForTenant ensures that a SUPER_ADMIN system role with full permissions
// exists for the given tenant. If it already exists, it is returned. Otherwise it is created
// using the global permission catalog. This helper is internal and does not expose any new
// endpoints; it is used by CreateTenant and can be reused by other internal flows if needed.
func (s *userService) ensureSuperAdminRoleForTenant(ctx context.Context, tenantID uuid.UUID, createdBy string) (*domain.Role, error) {
	// Check if exists
	roles, _, err := s.roleRepo.ListByTenant(ctx, tenantID, 100, 0) // Check first 100 roles
	if err != nil {
		return nil, err
	}
	for _, r := range roles {
		if r.Name == "SUPER_ADMIN" {
			return r, nil
		}
	}

	// Create if not exists
	// 1. Get all system permissions
	perms, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	var permKeys []string
	for _, p := range perms {
		permKeys = append(permKeys, p.Name) // Use Name as key? Or ID? Role uses keys (names)
	}

	// 2. Create Role
	role := &domain.Role{
		TenantID:     tenantID,
		Name:         "SUPER_ADMIN",
		Description:  "System Super Admin with full access",
		Permissions:  permKeys,
		IsSystemRole: true,
		CreatedBy:    createdBy,
	}
	return s.roleRepo.Create(ctx, role)
}

func (s *userService) CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	// Check if already exists
	var exists bool
	var err error
	if role.IsSystemRole {
		exists, err = s.roleRepo.ExistsSystem(ctx, role.TenantID, role.Name)
	} else {
		exists, err = s.roleRepo.Exists(ctx, role.TenantID, role.OrgID, role.Name)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to check role existence: %w", err)
	}
	if exists {
		return nil, fmt.Errorf("role with name '%s' already exists in this scope", role.Name)
	}

	return s.roleRepo.Create(ctx, role)
}

func (s *userService) GetRole(ctx context.Context, roleID uuid.UUID) (*domain.Role, error) {
	return s.roleRepo.GetByID(ctx, roleID)
}

func (s *userService) ListRolesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error) {
	return s.roleRepo.ListByTenant(ctx, tenantID, limit, offset)
}

func (s *userService) ListRolesByOrganization(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, includeSystem bool, limit, offset int32) ([]*domain.Role, int64, error) {
	roles, total, err := s.roleRepo.ListByTenantAndOrgIncludingSystem(ctx, tenantID, orgID, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	
	if includeSystem {
		return roles, total, nil
	}
	
	// Filter out system roles. Note: Pagination might be slightly off if we filter system roles here.
	// But usually, system roles are few. For strict pagination, the SQL query should handle the filtering.
	var filtered []*domain.Role
	for _, r := range roles {
		if !r.IsSystemRole {
			filtered = append(filtered, r)
		}
	}
	return filtered, total, nil
}

// UpdateRole updates role name/description/permissions
func (s *userService) UpdateRole(ctx context.Context, role *domain.Role) (*domain.Role, error) {
	updated, err := s.roleRepo.Update(ctx, role)
	if err != nil {
		return nil, fmt.Errorf("failed to update role: %w", err)
	}
	return updated, nil
}

// DeleteRole deletes a role by ID
func (s *userService) DeleteRole(ctx context.Context, roleID uuid.UUID) error {
	// Load role first to enforce domain rules around system roles
	role, err := s.roleRepo.GetByID(ctx, roleID)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}

	// Prevent deletion of system roles, especially SUPER_ADMIN
	if role.IsSystemRole || role.Name == "SUPER_ADMIN" {
		return fmt.Errorf("cannot delete system role %s", role.Name)
	}

	if err := s.roleRepo.Delete(ctx, roleID); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}
	return nil
}

// ListPermissions returns the fixed permission catalog, optionally filtered by module
func (s *userService) ListPermissions(ctx context.Context, module *string) ([]*domain.Permission, error) {
	if module != nil && *module != "" {
		perms, err := s.permissionRepo.ListByModule(ctx, module)
		if err != nil {
			return nil, fmt.Errorf("failed to list permissions by module: %w", err)
		}
		return perms, nil
	}

	perms, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list permissions: %w", err)
	}
	return perms, nil
}

// DeactivateUser implements soft delete for users
func (s *userService) DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID, reason string) (*domain.User, error) {
	// Get user - for deactivation we don't strictly NEED orgID if we trust userID
	// But let's assume it's global for now. Or should we pass OrgID?
	// The interface doesn't have it yet. Let's stick to what we have or update interface.
	user, err := s.userRepo.GetByID(ctx, userID, nil)

	// Fetch the admin user who is deactivating
	adminUser, err := s.userRepo.GetByID(ctx, deactivatedBy, nil)
	var adminName *string
	if err == nil && adminUser != nil {
		name := adminUser.Name
		adminName = &name
		// Log warning but proceed
		s.logger.Warn("Failed to fetch info for deactivator", "deactivated_by", deactivatedBy, "error", err)
	}

	// Mark as inactive
	user.IsActive = false
	now := time.Now()
	user.DeactivatedAt = &now
	user.DeactivatedBy = &deactivatedBy
	user.DeactivatedByName = adminName
	

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to deactivate user: %w", err)
	}

	// TODO: Create notification for super admin
	// This would be a call to notification service

	// TODO: Log activity
	// This would be a call to activity log service

	return updatedUser, nil
}

// ReactivateUser reactivates a deactivated user
func (s *userService) ReactivateUser(ctx context.Context, userID, reactivatedBy uuid.UUID) (*domain.User, error) {
	// Get user
	user, err := s.userRepo.GetByID(ctx, userID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Mark as active
	user.IsActive = true
	user.DeactivatedAt = nil
	user.DeactivatedBy = nil

	// Update user
	updatedUser, err := s.userRepo.Update(ctx, user)
	if err != nil {
		return nil, fmt.Errorf("failed to reactivate user: %w", err)
	}

	// TODO: Log activity

	return updatedUser, nil
}

// CreateActivityLog creates an activity log entry
func (s *userService) CreateActivityLog(ctx context.Context, log *domain.ActivityLog) (*domain.ActivityLog, error) {
	created, err := s.activityLogRepo.Create(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to create activity log: %w", err)
	}
	s.logger.Info("Activity Log created", "name", created.Name, "description", created.Description)
	return created, nil
}

// ListActivityLogs lists activity logs
func (s *userService) ListActivityLogs(ctx context.Context, limit, offset int32) ([]*domain.ActivityLog, int64, error) {
	logs, totalCount, err := s.activityLogRepo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list activity logs: %w", err)
	}
	return logs, totalCount, nil
}

// CreateLoginHistory creates a login history entry
func (s *userService) CreateLoginHistory(ctx context.Context, history *domain.UserLoginHistory) (*domain.UserLoginHistory, error) {
	ip := ""
	if history.IPAddress != nil {
		ip = *history.IPAddress
	}

	// Parse User-Agent using the more detailed mileusna library
	uaRaw := ""
	if history.UserAgent != nil {
		uaRaw = *history.UserAgent
	}
	ua := useragent.Parse(uaRaw)

	// Fetch detailed IP metadata (Location, ISP, VPN info)
	info := s.fetchIPInfo(ip)
	location := "Unknown"
	isp := "Unknown"
	isProxy := false

	if info != nil {
		location = fmt.Sprintf("%s, %s", info.City, info.Country)
		isp = info.ISP
		isProxy = info.Proxy
	}

	// Detect Firewall/Proxy: if IP contains comma or API detects proxy, it's likely a proxy chain
	firewallStatus := "Not Detected"
	if strings.Contains(ip, ",") || isProxy {
		firewallStatus = "Detected"
	}

	// [SECURITY BLOCK] If proxy/firewall is detected, block the login
	if firewallStatus == "Detected" {
		s.logger.Warn("SECURITY ALERT: Login blocked due to Proxy/VPN detection", "user_id", history.UserID, "ip", ip)
		return nil, fmt.Errorf("SECURITY ALERT: Access denied. Please disable VPN or Proxy to login")
	}

	// Refine OS display for Windows 10/11 ambiguity
	osDisplay := fmt.Sprintf("%s %s", ua.OS, ua.OSVersion)
	if ua.OS == "Windows" && (ua.OSVersion == "10" || ua.OSVersion == "10.0") {
		osDisplay = "Windows 10/11"
	}

	// Clean, professional, and detailed format (saaff saff words)
	// Example: OS: Windows 11, Browser: Chrome 131, Location: Delhi, India, ISP: Jio, Firewall: Detected
	formattedUA := fmt.Sprintf("OS: %s, Browser: %s %s, Location: %s, ISP: %s, Firewall: %s",
		osDisplay, ua.Name, ua.Version, location, isp, firewallStatus)

	// Update history object with the clean formatted string before saving to history table
	history.UserAgent = &formattedUA

	created, err := s.loginHistoryRepo.Create(ctx, history)
	if err != nil {
		return nil, fmt.Errorf("failed to create login history: %w", err)
	}

	// Also update the user's last_login_at, last_login_ip, and user_agent in the users table
	if err := s.userRepo.UpdateLastLogin(ctx, history.UserID, ip, formattedUA); err != nil {
		// Log error but don't fail the history creation
		s.logger.Warn("Failed to update user last_login", "error", err)
	}

	return created, nil
}

// ListLoginHistory lists login history for a user
func (s *userService) ListLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.UserLoginHistory, int64, error) {
	histories, totalCount, err := s.loginHistoryRepo.ListByUser(ctx, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list login history: %w", err)
	}
	return histories, totalCount, nil
}

// CreateTenant creates a new tenant with super admin user
func (s *userService) CreateTenant(ctx context.Context, name, email, password, role string) (*domain.Tenant, error) {
	// Generate tenant ID
	tenantID := uuid.New()
	now := time.Now()

	// Hash password for super admin
	hashedPassword, err := s.hashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create tenant
	tenant := &domain.Tenant{
		TenantID:  tenantID,
		Name:      name,
		Email:     email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if _, err := s.tenantRepo.Create(ctx, tenant); err != nil {
		return nil, fmt.Errorf("failed to create tenant record: %w", err)
	}

	// Create super admin user for this tenant
	superAdmin := &domain.User{
		UserID:          uuid.New(),
		TenantID:        tenantID,
		Name:            name,
		Email:           email,
		Password:        hashedPassword,
		EmailVerifiedAt: &now,
		IsActive:        true,
		CreatedAt:       now,
		UpdatedAt:       now,
	}

	// Store super admin user (tenant creation creates the first user)
	createdSuperAdmin, err := s.userRepo.Create(ctx, superAdmin)
	if err != nil {
		return nil, fmt.Errorf("failed to create super admin user: %w", err)
	}

	// Ensure SUPER_ADMIN role with full permissions exists for this tenant
	createdRole, err := s.ensureSuperAdminRoleForTenant(ctx, tenantID, name)
	if err != nil {
		return nil, err
	}

	// Assign SUPER_ADMIN role to the initial super admin user
	if err := s.userRoleRepo.AssignRole(ctx, createdSuperAdmin.UserID, createdRole.RoleID); err != nil {
		return nil, fmt.Errorf("failed to assign super admin role to user: %w", err)
	}

	s.logger.Info("Created tenant with super admin user and role", "tenant_id", tenantID, "user_id", createdSuperAdmin.UserID, "role_id", createdRole.RoleID)

	return tenant, nil
}

// GetTenant retrieves tenant information by tenant ID
func (s *userService) GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error) {
	tenant, err := s.tenantRepo.GetByID(ctx, tenantID)
	if err != nil {
		return nil, fmt.Errorf("failed to get tenant: %w", err)
	}

	return tenant, nil
}

// hashPassword hashes a password using bcrypt
func (s *userService) hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

type ipInfo struct {
	Status  string `json:"status"`
	Country string `json:"country"`
	City    string `json:"city"`
	ISP     string `json:"isp"`
	Proxy   bool   `json:"proxy"`
}

func (s *userService) fetchIPInfo(ip string) *ipInfo {
	// Use the first IP if it's a list (X-Forwarded-For)
	targetIP := strings.Split(ip, ",")[0]
	targetIP = strings.TrimSpace(targetIP)

	// RFC 1918 and Private Ranges Check
	// Immediately skip external API for local/private IPs to save seconds of timeout
	if targetIP == "" || targetIP == "127.0.0.1" || targetIP == "::1" || 
		strings.HasPrefix(targetIP, "192.168.") || 
		strings.HasPrefix(targetIP, "10.") || 
		strings.HasPrefix(targetIP, "172.16.") || 
		strings.HasPrefix(targetIP, "172.17.") || 
		strings.HasPrefix(targetIP, "172.18.") || 
		strings.HasPrefix(targetIP, "172.19.") || 
		strings.HasPrefix(targetIP, "172.20.") || 
		strings.HasPrefix(targetIP, "172.21.") || 
		strings.HasPrefix(targetIP, "172.22.") || 
		strings.HasPrefix(targetIP, "172.23.") || 
		strings.HasPrefix(targetIP, "172.24.") || 
		strings.HasPrefix(targetIP, "172.25.") || 
		strings.HasPrefix(targetIP, "172.26.") || 
		strings.HasPrefix(targetIP, "172.27.") || 
		strings.HasPrefix(targetIP, "172.28.") || 
		strings.HasPrefix(targetIP, "172.29.") || 
		strings.HasPrefix(targetIP, "172.30.") || 
		strings.HasPrefix(targetIP, "172.31.") {
		s.logger.Debug("Private IP detected, skipping external enrichment", "ip", targetIP)
		return nil
	}

	s.logger.Debug("Fetching IP metadata", "ip", targetIP)

	url := fmt.Sprintf("http://ip-api.com/json/%s?fields=status,country,city,isp,proxy", targetIP)
	// Lower timeout significantly. Login/Refresh should NOT wait 5 seconds for enrichment.
	client := &http.Client{Timeout: 800 * time.Millisecond} 
	resp, err := client.Get(url)
	if err != nil {
		s.logger.Debug("IP Info fetch failed or timed out", "error", err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil
	}

	var info ipInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil
	}
	if info.Status != "success" && info.Status != "" {
		return nil
	}
	return &info
}
