package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	"github.com/google/uuid"
)

type registrationService struct {
	userRepo       ports.UserRepository
	orgClient      ports.OrganizationServiceClient // gRPC client for organization service
	roleRepo       ports.RoleRepository
	jwtManager     *utils.JWTManager
	emailService   utils.EmailService
	kafkaPublisher ports.KafkaPublisher
}

// NewRegistrationService creates a new registration service
func NewRegistrationService(
	userRepo ports.UserRepository,
	orgClient ports.OrganizationServiceClient,
	roleRepo ports.RoleRepository,
	jwtManager *utils.JWTManager,
	emailService utils.EmailService,
	kafkaPublisher ports.KafkaPublisher,
) *registrationService {
	return &registrationService{
		userRepo:       userRepo,
		orgClient:      orgClient,
		roleRepo:       roleRepo,
		jwtManager:     jwtManager,
		emailService:   emailService,
		kafkaPublisher: kafkaPublisher,
	}
}

// RegisterUserWithOrganization handles complete registration flow:
// Step 1: Create Organization
// Step 2: Create Super Admin user
// Step 3: Assign Super Admin role
// Step 4: Send welcome email
// Step 5: Publish Kafka events
func (s *registrationService) RegisterUserWithOrganization(ctx context.Context, req *domain.RegistrationRequest) (*domain.RegistrationResponse, error) {
	// Validate request
	if err := s.validateRegistrationRequest(req); err != nil {
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, uuid.Nil, req.Email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", req.Email)
	}

	// Generate tenant ID for new registration
	tenantID := uuid.New()

	// Step 1: Create Organization via gRPC
	orgID, err := s.orgClient.CreateOrganization(ctx, tenantID, req.OrganizationName, req.OrganizationCode, tenantID) // User ID will be updated after user creation
	if err != nil {
		return nil, fmt.Errorf("failed to create organization: %w", err)
	}

	// Step 2: Create Super Admin User
	hashedPassword, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	userID := uuid.New()
	now := time.Now()

	// Create user entity (simplified - should use User domain from user-service)
	user := &domain.User{
		UserID:    userID,
		TenantID:  tenantID,
		Name:      req.Name,
		Email:     req.Email,
		Password:  hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Convert domain.User to ports.UserData for repository
	userData := &ports.UserData{
		UserID:    user.UserID,
		TenantID:  user.TenantID,
		Email:     user.Email,
		Name:      user.Name,
		Password:  user.Password,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}

	// Save user to database
	_, err = s.userRepo.Create(ctx, userData)
	if err != nil {
		// Rollback: Delete organization if user creation fails
		_ = s.orgClient.DeleteOrganization(ctx, orgID)
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	// Step 3: Create Super Admin Role
	superAdminRole, err := s.createSuperAdminRole(ctx, tenantID, orgID, userID)
	if err != nil {
		// Rollback
		_ = s.userRepo.Delete(ctx, userID)
		_ = s.orgClient.DeleteOrganization(ctx, orgID)
		return nil, fmt.Errorf("failed to create super admin role: %w", err)
	}

	// Assign Super Admin role to user
	err = s.roleRepo.AssignRoleToUser(ctx, userID, superAdminRole.RoleID, orgID)
	if err != nil {
		return nil, fmt.Errorf("failed to assign super admin role: %w", err)
	}

	// Step 4: Update organization with super admin ID
	err = s.orgClient.SetSuperAdmin(ctx, orgID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to set super admin: %w", err)
	}

	// Step 5: Generate JWT tokens
	token, err := s.jwtManager.GenerateToken(userID, req.Email, req.Name, tenantID, &orgID, []string{"SUPER_ADMIN"}, superAdminRole.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	refreshToken, _, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String(), orgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Step 6: Send welcome email
	go func() {
		welcomeData := map[string]interface{}{
			"name":           req.Name,
			"email":          req.Email,
			"organization":   req.OrganizationName,
			"is_super_admin": true,
		}
		_ = s.emailService.SendWelcomeEmail(req.Email, req.Name, welcomeData)
	}()

	// Step 7: Publish Kafka events
	if s.kafkaPublisher != nil {
		// Publish User Registered event
		go func() {
			eventData := map[string]interface{}{
				"user_id":        userID.String(),
				"tenant_id":      tenantID.String(),
				"org_id":         orgID.String(),
				"email":          req.Email,
				"name":           req.Name,
				"is_super_admin": true,
				"timestamp":      time.Now(),
			}
			_ = s.kafkaPublisher.Publish(ctx, "user.registered", eventData)
		}()

		// Publish Organization Created event
		go func() {
			eventData := map[string]interface{}{
				"org_id":     orgID.String(),
				"tenant_id":  tenantID.String(),
				"org_name":   req.OrganizationName,
				"org_code":   req.OrganizationCode,
				"created_by": userID.String(),
				"timestamp":  time.Now(),
			}
			_ = s.kafkaPublisher.Publish(ctx, "organization.created", eventData)
		}()
	}

	// Step 8: Log activity
	activityLog := domain.NewActivityLog(
		userID,
		"REGISTER_WITH_ORGANIZATION",
		"USER",
		&[]string{userID.String()}[0],
		map[string]interface{}{
			"organization_id":   orgID.String(),
			"organization_name": req.OrganizationName,
			"is_super_admin":    true,
		},
		"", // IP address - should be passed from context
		"", // User agent - should be passed from context
		tenantID,
		&orgID,
	)
	// Save activity log (async)
	go func() {
		// TODO: Save to activity_logs table
		_ = activityLog
	}()

	return &domain.RegistrationResponse{
		UserID:           userID,
		Email:            req.Email,
		Name:             req.Name,
		OrganizationID:   orgID,
		OrganizationName: req.OrganizationName,
		IsSuperAdmin:     true,
		Token:            token,
		RefreshToken:     refreshToken,
		Message:          "Registration successful! You are now the Super Admin of " + req.OrganizationName,
	}, nil
}

// createSuperAdminRole creates the super admin role with all permissions
func (s *registrationService) createSuperAdminRole(ctx context.Context, tenantID, orgID, createdBy uuid.UUID) (*domain.SuperAdminRole, error) {
	// Define all permissions for super admin
	permissions := []string{
		"users.create", "users.read", "users.update", "users.delete",
		"roles.create", "roles.read", "roles.update", "roles.delete",
		"permissions.assign",
		"organizations.read", "organizations.update",
		"departments.create", "departments.read", "departments.update", "departments.delete",
		"designations.create", "designations.read", "designations.update", "designations.delete",
		"projects.create", "projects.read", "projects.update", "projects.delete",
		"vendors.create", "vendors.read", "vendors.update", "vendors.delete",
		"reports.view", "reports.export",
		"settings.manage",
		"activity_logs.view",
		"*", // Wildcard - full access
	}

	roleID := uuid.New()
	role := &domain.Role{
		RoleID:       roleID,
		TenantID:     tenantID,
		OrgID:        &orgID,
		Name:         "Super Admin",
		Description:  "Super Administrator with full access to all resources",
		Permissions:  permissions,
		IsSystemRole: true,
		CreatedBy:    &createdBy,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	// Save role to database
	_, err := s.roleRepo.Create(ctx, role)
	if err != nil {
		return nil, err
	}

	return &domain.SuperAdminRole{
		RoleID:      roleID,
		RoleName:    "Super Admin",
		Permissions: permissions,
	}, nil
}

// validateRegistrationRequest validates the registration request
func (s *registrationService) validateRegistrationRequest(req *domain.RegistrationRequest) error {
	// Validate name
	if req.Name == "" || len(req.Name) < 2 {
		return fmt.Errorf("name must be at least 2 characters")
	}

	// Validate email
	if !utils.IsValidEmail(req.Email) {
		return domain.ErrInvalidEmail
	}

	// Validate password
	if err := utils.ValidatePasswordStrength(req.Password); err != nil {
		return domain.ErrWeakPassword
	}

	// Validate organization name
	if len(req.OrganizationName) < 3 || len(req.OrganizationName) > 255 {
		return domain.ErrInvalidOrganizationName
	}

	// Generate org code if not provided
	if req.OrganizationCode == "" {
		req.OrganizationCode = s.generateOrgCode(req.OrganizationName)
	}

	return nil
}

// generateOrgCode generates organization code from name
func (s *registrationService) generateOrgCode(orgName string) string {
	// Take first 3 letters of each word, uppercase
	words := strings.Fields(orgName)
	code := ""
	for _, word := range words {
		if len(word) >= 3 {
			code += strings.ToUpper(word[:3])
		} else {
			code += strings.ToUpper(word)
		}
		if len(code) >= 6 {
			break
		}
	}

	// Add random suffix to ensure uniqueness
	suffix := fmt.Sprintf("%03d", time.Now().UnixNano()%1000)
	return code + suffix
}
