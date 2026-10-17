package services

import (
	"context"
	"fmt"
	"strings"
	"time"

	userpb "github.com/ShristiRnr/NHIT_Backend/api/pb/userpb"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/config"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/domain"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/core/ports"
	"github.com/ShristiRnr/NHIT_Backend/services/auth-service/internal/utils"
	"github.com/google/uuid"
	"google.golang.org/grpc/metadata"
	"log/slog"
)

// authService implements the AuthService interface
type authService struct {
	userRepo              ports.UserRepository
	sessionRepo           ports.SessionRepository
	refreshTokenRepo      ports.RefreshTokenRepository
	passwordResetRepo     ports.PasswordResetRepository
	emailVerificationRepo ports.EmailVerificationRepository
	jwtManager            *utils.JWTManager
	emailService          utils.EmailService
	kafkaPublisher        ports.KafkaPublisher
	notificationClient    ports.NotificationClient
	userServiceClient     userpb.UserManagementClient     // gRPC client for User Service
	orgClient             ports.OrganizationServiceClient // gRPC client for Organization Service
	ssoConfig             *config.SSOConfig
	logger                *slog.Logger
}

// NewAuthService creates a new auth service
func NewAuthService(
	userRepo ports.UserRepository,
	sessionRepo ports.SessionRepository,
	refreshTokenRepo ports.RefreshTokenRepository,
	passwordResetRepo ports.PasswordResetRepository,
	emailVerificationRepo ports.EmailVerificationRepository,
	jwtManager *utils.JWTManager,
	emailService utils.EmailService,
	kafkaPublisher ports.KafkaPublisher,
	notificationClient ports.NotificationClient,
	userServiceClient userpb.UserManagementClient,
	orgClient ports.OrganizationServiceClient,
	ssoConfig *config.SSOConfig,
	logger *slog.Logger,
) ports.AuthService {
	return &authService{
		userRepo:              userRepo,
		sessionRepo:           sessionRepo,
		refreshTokenRepo:      refreshTokenRepo,
		passwordResetRepo:     passwordResetRepo,
		emailVerificationRepo: emailVerificationRepo,
		jwtManager:            jwtManager,
		emailService:          emailService,
		kafkaPublisher:        kafkaPublisher,
		notificationClient:    notificationClient,
		userServiceClient:     userServiceClient,
		orgClient:             orgClient,
		ssoConfig:             ssoConfig,
		logger:                logger,
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

// Register registers a new user
func (s *authService) Register(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name, email, password string, roles []string) (*domain.LoginResponse, error) {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(password); err != nil {
		return nil, fmt.Errorf("password validation failed: %w", err)
	}

	// Hash password
	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err == nil && existingUser != nil {
		return nil, fmt.Errorf("user with email %s already exists", email)
	}

	// ✅ OPTION 2: Create user via User Service gRPC call (Industry Standard)
	// Match the ACTUAL generated proto structure
	createUserReq := &userpb.CreateUserRequest{
		TenantId: tenantID.String(), // Field 1
		Email:    email,             // Field 2
		Name:     name,              // Field 3
		Password: hashedPassword,    // Field 4 - hashed password
		// Roles will be assigned separately after user creation
	}

	ctx = s.mergeMetadata(ctx, metadata.Pairs("tenant_id", tenantID.String()))
	userResp, err := s.userServiceClient.CreateUser(ctx, createUserReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create user via User Service: %w", err)
	}

	// Parse user ID from response
	userID, err := uuid.Parse(userResp.UserId)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID returned from User Service: %w", err)
	}

	s.logger.Info("User created via User Service", "user_id", userResp.UserId, "email", email)

	// Send verification email
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		// If email fails, notify user to update email
		s.logger.Warn("Failed to send verification email", "email", email, "error", err)
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			s.logger.Warn("Failed to send email update notification", "error", err)
		}
	}

	// Generate tokens
	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		userID.String(), email, name, tenantID.String(), "", roles, []string{},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	// Handle orgID for token generation
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String(), orgIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       userID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	if _, err := s.sessionRepo.Create(ctx, session); err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           userID,
		Email:            email,
		Name:             name,
		Roles:            roles,
		Permissions:      []string{},
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         tenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
	}, nil
}

// LoginGlobal authenticates a user without requiring tenant_id
func (s *authService) LoginGlobal(ctx context.Context, email, password string, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Get user by email globally (across all tenants)
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		s.logger.Warn("LoginGlobal: User not found", "email", email, "error", err)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		s.logger.Warn("LoginGlobal: Password mismatch", "email", email, "error", err)
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Check if email is verified unless user is super admin
	if user.EmailVerifiedAt == nil {
		roles, err := s.userServiceClient.ListRolesOfUser(s.mergeMetadata(ctx, nil), &userpb.GetUserRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user roles: %w", err)
		}

		isSuperAdmin := false
		for _, role := range roles.Roles {
			if role.Name == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			return nil, fmt.Errorf("email not verified. Please verify your email before logging in")
		}
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		s.logger.Warn("Failed to update last login", "user_id", user.UserID, "error", err)
	}

	// Auto-detect organization if not provided
	if orgID == nil {
		orgsResp, err := s.userServiceClient.ListUserOrganizations(s.mergeMetadata(ctx, nil), &userpb.ListUserOrganizationsRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			s.logger.Warn("Failed to list user organizations", "user_id", user.UserID, "error", err)
		} else if len(orgsResp.Organizations) > 0 {
			var selected *userpb.UserOrganizationInfo
			for _, o := range orgsResp.Organizations {
				if o.IsCurrentContext {
					selected = o
					break
				}
			}
			if selected == nil {
				selected = orgsResp.Organizations[0]
			}
			if selected.OrgId != "" {
				if parsedOrgID, err := uuid.Parse(selected.OrgId); err != nil {
					s.logger.Warn("Invalid org_id for user", "org_id", selected.OrgId, "user_id", user.UserID, "error", err)
				} else {
					orgID = &parsedOrgID
				}
			}
		}
	}

	if orgID == nil {
		return nil, fmt.Errorf("no organization found for this user; please create an organization before logging in")
	}

	orgIDStr := orgID.String()

	// Fetch roles and permissions for JWT claims
	rolesResp, err := s.userServiceClient.ListRolesOfUser(s.mergeMetadata(ctx, nil), &userpb.GetUserRequest{
		UserId: user.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String(), orgIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session creation
	s.logger.Info("Global login: Created new session",
		"session_id", createdSession.SessionID,
		"user_id", user.UserID,
		"tenant_id", user.TenantID,
		"expires_at", createdSession.ExpiresAt)

	// Create login history entry
	if s.userServiceClient != nil {
		// Extract IP and User-Agent from context metadata
		md, _ := metadata.FromIncomingContext(ctx)
		ipAddress := firstMetadataValue(md, "x-forwarded-for", "x-real-ip", "remote-addr")
		userAgent := firstMetadataValue(md, "grpcgateway-user-agent", "x-user-agent", "user-agent")
		_, err = s.userServiceClient.CreateUserLoginHistory(s.mergeMetadata(ctx, nil), &userpb.CreateUserLoginHistoryRequest{
			UserId:    user.UserID.String(),
			IpAddress: ipAddress,
			UserAgent: userAgent,
		})
		if err != nil {
			s.logger.Warn("Failed to create login history", "user_id", user.UserID, "error", err)
			// Don't fail login if history creation fails
		}

		// Create activity log for login
		_, err = s.userServiceClient.CreateActivityLog(s.mergeMetadata(ctx, nil), &userpb.CreateActivityLogRequest{
			Name:        fmt.Sprintf("User Logged in with email-ID [%s]", user.Email),
			Description: "User successfully logged in",
		})
		if err != nil {
			s.logger.Warn("Failed to create activity log", "user_id", user.UserID, "error", err)
			// Don't fail login if activity log creation fails
		}
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// Login authenticates a user and returns tokens (tenant-specific)
func (s *authService) Login(ctx context.Context, email, password string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, tenantID, email)
	if err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Verify password
	if err := utils.VerifyPassword(user.Password, password); err != nil {
		return nil, fmt.Errorf("invalid email or password")
	}

	// Check if user is active
	if !user.IsActive {
		return nil, fmt.Errorf("user account is deactivated")
	}

	// Check if email is verified unless user is super admin
	if user.EmailVerifiedAt == nil {
		roles, err := s.userServiceClient.ListRolesOfUser(s.mergeMetadata(ctx, nil), &userpb.GetUserRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to fetch user roles: %w", err)
		}

		isSuperAdmin := false
		for _, role := range roles.Roles {
			if role.Name == "SUPER_ADMIN" {
				isSuperAdmin = true
				break
			}
		}

		if !isSuperAdmin {
			return nil, fmt.Errorf("email not verified. Please verify your email before logging in")
		}
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		s.logger.Warn("Failed to update last login", "user_id", user.UserID, "error", err)
	}

	// Auto-detect organization if not provided
	if orgID == nil {
		orgsResp, err := s.userServiceClient.ListUserOrganizations(s.mergeMetadata(ctx, nil), &userpb.ListUserOrganizationsRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			s.logger.Warn("Failed to list user organizations", "user_id", user.UserID, "error", err)
		} else if len(orgsResp.Organizations) > 0 {
			var selected *userpb.UserOrganizationInfo
			for _, o := range orgsResp.Organizations {
				if o.IsCurrentContext {
					selected = o
					break
				}
			}
			if selected == nil {
				selected = orgsResp.Organizations[0]
			}
			if selected.OrgId != "" {
				if parsedOrgID, err := uuid.Parse(selected.OrgId); err != nil {
					s.logger.Warn("Invalid org_id for user", "org_id", selected.OrgId, "user_id", user.UserID, "error", err)
				} else {
					orgID = &parsedOrgID
				}
			}
		}
	}

	if orgID == nil {
		return nil, fmt.Errorf("no organization found for this user; please create an organization before logging in")
	}

	orgIDStr := orgID.String()

	// Fetch roles and permissions for JWT claims
	rolesResp, err := s.userServiceClient.ListRolesOfUser(s.mergeMetadata(ctx, nil), &userpb.GetUserRequest{
		UserId: user.UserID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String(), orgIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	// Log session creation
	s.logger.Info("Created new session",
		"session_id", createdSession.SessionID,
		"user_id", user.UserID,
		"expires_at", createdSession.ExpiresAt)

	// Create login history entry
	if s.userServiceClient != nil {
		// Extract IP and User-Agent from context metadata
		md, _ := metadata.FromIncomingContext(ctx)
		ipAddress := firstMetadataValue(md, "x-forwarded-for", "x-real-ip", "remote-addr")
		userAgent := firstMetadataValue(md, "grpcgateway-user-agent", "x-user-agent", "user-agent")
		_, err = s.userServiceClient.CreateUserLoginHistory(s.mergeMetadata(ctx, nil), &userpb.CreateUserLoginHistoryRequest{
			UserId:    user.UserID.String(),
			IpAddress: ipAddress,
			UserAgent: userAgent,
		})
		if err != nil {
			s.logger.Warn("Failed to create login history", "user_id", user.UserID, "error", err)
			// Don't fail login if history creation fails
		}

		// Create activity log for login
		_, err = s.userServiceClient.CreateActivityLog(s.mergeMetadata(ctx, nil), &userpb.CreateActivityLogRequest{
			Name:        fmt.Sprintf("User Logged in with email-ID [%s]", user.Email),
			Description: "User successfully logged in",
		})
		if err != nil {
			s.logger.Warn("Failed to create activity log", "user_id", user.UserID, "error", err)
			// Don't fail login if activity log creation fails
		}
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// Logout logs out a user by invalidating their refresh token and session
func (s *authService) Logout(ctx context.Context, userID uuid.UUID, refreshToken string) error {
	// Delete refresh token
	if err := s.refreshTokenRepo.Delete(ctx, refreshToken); err != nil {
		s.logger.Warn("Failed to delete refresh token", "error", err)
	}

	// Find token associated with session to invalidate the specific session
	// Default to invalidating all sessions for safety
	invalidatedSessions := 0

	// Get all sessions for the user
	sessions, err := s.sessionRepo.GetByUserID(ctx, userID)
	if err != nil {
		s.logger.Warn("Failed to get sessions for user", "user_id", userID, "error", err)
		// Continue despite error - we'll try to delete sessions anyway
	}

	// Try to find a session with this refresh token
	sessionFound := false
	for _, session := range sessions {
		// Since we don't store the refresh token in the session, we delete all sessions
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			s.logger.Warn("Failed to delete session", "session_id", session.SessionID, "error", err)
		} else {
			invalidatedSessions++
			sessionFound = true
		}
	}

	if !sessionFound && len(sessions) > 0 {
		s.logger.Warn("No sessions found to invalidate for user", "user_id", userID)
	}

	s.logger.Info("Logout complete", "user_id", userID, "invalidated_sessions", invalidatedSessions)
	
	// Update last logout time
	if err := s.userRepo.UpdateLastLogout(ctx, userID); err != nil {
		s.logger.Warn("Failed to update last logout", "user_id", userID, "error", err)
	}

	// Create activity log for logout
	if s.userServiceClient != nil {
		user, err := s.userRepo.GetByID(ctx, userID)
		email := "unknown"
		if err == nil && user != nil {
			email = user.Email
		}
		_, err = s.userServiceClient.CreateActivityLog(s.mergeMetadata(ctx, nil), &userpb.CreateActivityLogRequest{
			Name:        fmt.Sprintf("User Logged out with email-ID [%s]", email),
			Description: "User successfully logged out",
		})
		if err != nil {
			s.logger.Warn("Failed to create activity log for logout", "user_id", userID, "error", err)
		}
	}

	return nil
}

// RefreshToken refreshes an access token using a refresh token
func (s *authService) RefreshToken(ctx context.Context, refreshToken string, tenantID uuid.UUID, orgID *uuid.UUID) (*domain.LoginResponse, error) {
	start := time.Now()
	// Validate refresh token
	claims, err := s.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	userIDStr := claims.Subject
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user ID in token: %w", err)
	}

	// If tenant/org ID were not provided in arguments (e.g. from handler), try to use what's in the token
	if tenantID == uuid.Nil && claims.TenantID != "" {
		if parsed, err := uuid.Parse(claims.TenantID); err == nil {
			tenantID = parsed
		}
	}
	
	if orgID == nil && claims.OrgID != "" {
		if parsed, err := uuid.Parse(claims.OrgID); err == nil {
			orgID = &parsed
		}
	}

	// 1. Concurrent check: refresh token + user existence + roles
	type result struct {
		user      *ports.UserData
		rolesResp *userpb.ListRolesResponse
		err       error
		step      string
	}
	resChan := make(chan result, 2)

	// A. Check refresh token in parallel
	go func() {
		storedUserID, err := s.refreshTokenRepo.GetUserIDByToken(ctx, refreshToken)
		if err != nil {
			resChan <- result{err: fmt.Errorf("refresh token not found or expired"), step: "GetUserIDByToken"}
			return
		}
		if storedUserID != userID {
			resChan <- result{err: fmt.Errorf("refresh token user ID mismatch"), step: "TokenUserMismatch"}
			return
		}
		resChan <- result{step: "TokenValid"}
	}()

	// B. Get user details and roles in parallel
	go func() {
		user, err := s.userRepo.GetByID(ctx, userID)
		if err != nil {
			resChan <- result{err: fmt.Errorf("failed to get user: %w", err), step: "GetByID"}
			return
		}
		
		orgIDStr := ""
		if orgID != nil {
			orgIDStr = orgID.String()
		}
		rolesCtx := s.mergeMetadata(ctx, metadata.Pairs(
			"tenant_id", user.TenantID.String(),
			"org_id", orgIDStr,
		))
		rolesResp, err := s.userServiceClient.ListRolesOfUser(rolesCtx, &userpb.GetUserRequest{
			UserId: user.UserID.String(),
		})
		if err != nil {
			resChan <- result{err: fmt.Errorf("failed to fetch user roles for token: %w", err), step: "ListRoles"}
			return
		}
		resChan <- result{user: user, rolesResp: rolesResp, step: "UserRolesDone"}
	}()

	var user *ports.UserData
	var rolesResp *userpb.ListRolesResponse
	for i := 0; i < 2; i++ {
		res := <-resChan
		if res.err != nil {
			return nil, res.err
		}
		if res.step == "UserRolesDone" {
			user = res.user
			rolesResp = res.rolesResp
		}
	}

	// 2. Generate new tokens
	orgIDStr := ""
	if orgID != nil {
		orgIDStr = orgID.String()
	}

	roleNames := make([]string, 0, len(rolesResp.Roles))
	permSet := make(map[string]struct{})
	for _, r := range rolesResp.Roles {
		if r.Name != "" {
			roleNames = append(roleNames, r.Name)
		}
		for _, p := range r.Permissions {
			permSet[p] = struct{}{}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	newRefreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String(), orgIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// 3. Parallel updates/creation
	updateResChan := make(chan error, 2)

	// A. Update old token expiry & Create new token in background
	go func() {
		gracePeriod := 5 * time.Minute
		if err := s.refreshTokenRepo.UpdateExpiresAt(ctx, refreshToken, time.Now().Add(gracePeriod)); err != nil {
			s.logger.Warn("Failed to update old refresh token expiry", "error", err)
		}
		
		err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
			Token:     newRefreshToken,
			UserID:    user.UserID,
			ExpiresAt: time.Unix(refreshExpiresAt, 0),
			CreatedAt: time.Now(),
		})
		updateResChan <- err
	}()

	// B. Create New Session
	var createdSession *domain.Session
	go func() {
		session := &domain.Session{
			SessionID:    uuid.New(),
			UserID:       user.UserID,
			SessionToken: accessToken,
			CreatedAt:    time.Now().UTC(),
			ExpiresAt:    time.Unix(accessExpiresAt, 0).UTC(),
		}
		var err error
		createdSession, err = s.sessionRepo.Create(ctx, session)
		updateResChan <- err
	}()

	// C. Update Last Login (Async) - Don't wait for it
	go func() {
		if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
			s.logger.Warn("Failed to update last login during refresh", "user_id", user.UserID, "error", err)
		}
	}()

	// Wait for critical background tasks
	for i := 0; i < 2; i++ {
		if err := <-updateResChan; err != nil {
			return nil, err
		}
	}

	s.logger.Info("RefreshToken: Completed", "user_id", user.UserID, "duration", time.Since(start))

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     newRefreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// ValidateToken validates an access token
func (s *authService) ValidateToken(ctx context.Context, token string) (*domain.TokenValidation, error) {
	claims, err := s.jwtManager.ValidateToken(token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("invalid token: %w", err)
	}

	// Check if session exists and is valid
	session, err := s.sessionRepo.GetByToken(ctx, token)
	if err != nil {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session not found")
	}

	if session.ExpiresAt.Before(time.Now().UTC()) {
		return &domain.TokenValidation{Valid: false}, fmt.Errorf("session expired")
	}

	userID, _ := uuid.Parse(claims.UserID)
	tenantID, _ := uuid.Parse(claims.TenantID)
	var orgID *uuid.UUID
	if claims.OrgID != "" {
		oid, _ := uuid.Parse(claims.OrgID)
		orgID = &oid
	}

	expiresAt := time.Now()
	if claims.RegisteredClaims.ExpiresAt != nil {
		expiresAt = claims.RegisteredClaims.ExpiresAt.Time
	}

	return &domain.TokenValidation{
		Valid:       true,
		UserID:      userID,
		Email:       claims.Email,
		Name:        claims.Name,
		TenantID:    tenantID,
		OrgID:       orgID,
		Roles:       claims.Roles,
		Permissions: claims.Permissions,
		ExpiresAt:   expiresAt,
	}, nil
}

// SendVerificationEmail sends a verification email to the user
func (s *authService) SendVerificationEmail(ctx context.Context, userID uuid.UUID) error {
	// Get user details (in production, call User Service)
	// For now, we'll use mock data
	email := "user@example.com"
	name := "User"

	// Create verification token
	verificationToken, err := s.emailVerificationRepo.Create(ctx, userID, time.Now().Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("failed to create verification token: %w", err)
	}

	// Send email with error handling
	if err := s.emailService.SendVerificationEmail(email, name, verificationToken.Token.String()); err != nil {
		s.logger.Warn("Failed to send verification email", "error", err)
		// Send notification about email failure
		if err := s.emailService.SendEmailUpdateNotification(email, name); err != nil {
			s.logger.Warn("Failed to send email update notification", "error", err)
		}
		return fmt.Errorf("failed to send verification email. Please update your email address")
	}

	return nil
}

// VerifyEmail verifies a user's email address
func (s *authService) VerifyEmail(ctx context.Context, userID uuid.UUID, token string) error {
	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid verification token format")
	}

	// Verify token
	valid, err := s.emailVerificationRepo.Verify(ctx, userID, tokenUUID)
	if err != nil {
		return fmt.Errorf("failed to verify token: %w", err)
	}

	if !valid {
		return fmt.Errorf("invalid or expired verification token")
	}

	// Update user email verified status
	if err := s.userRepo.VerifyEmail(ctx, userID); err != nil {
		return fmt.Errorf("failed to update email verification status: %w", err)
	}

	// Delete verification token
	if err := s.emailVerificationRepo.Delete(ctx, userID); err != nil {
		s.logger.Warn("Failed to delete verification token", "user_id", userID, "error", err)
	}

	return nil
}

// ForgotPassword initiates a password reset flow
func (s *authService) ForgotPassword(ctx context.Context, email string) error {
	// Get user by email globally to find their tenant_id
	user, err := s.userRepo.GetByEmailGlobal(ctx, email)
	if err != nil {
		// Don't reveal if user exists or not for security
		s.logger.Warn("User not found during forgot password request", "email", email, "error", err)
		return nil
	}

	// Create password reset token
	resetToken := uuid.New()
	expiresAt := time.Now().Add(1 * time.Hour)
	_, err = s.passwordResetRepo.Create(ctx, user.UserID, resetToken, expiresAt)
	if err != nil {
		return fmt.Errorf("failed to create password reset token: %w", err)
	}

	// Use notification service for sending emails
	if s.notificationClient != nil {
		if err := s.notificationClient.SendPasswordResetEmail(ctx, email, user.Name, resetToken.String()); err != nil {
			s.logger.Warn("Failed to send password reset email via notification service", "error", err)
			// Fallback to legacy email service
			return s.sendPasswordResetEmailLegacy(email, user.Name, resetToken.String())
		}
	} else {
		// Fallback to legacy email service if notification client is not configured
		return s.sendPasswordResetEmailLegacy(email, user.Name, resetToken.String())
	}

	if s.kafkaPublisher != nil {
		eventTimestamp := time.Now()
		eventData := map[string]interface{}{
			"user_id":      user.UserID.String(),
			"tenant_id":    user.TenantID.String(),
			"email":        user.Email,
			"name":         user.Name,
			"reset_token":  resetToken.String(),
			"expires_at":   expiresAt.Format(time.RFC3339),
			"requested_at": eventTimestamp.Format(time.RFC3339),
		}
		if err := s.kafkaPublisher.Publish(ctx, "user.password_reset_requested", eventData); err != nil {
			s.logger.Warn("Failed to publish password reset event", "error", err)
		}
	}

	return nil
}

// ResetPasswordByToken resets a user's password using a reset token
func (s *authService) ResetPasswordByToken(ctx context.Context, token, newPassword string) error {
	// Validate password strength
	if err := utils.ValidatePasswordStrength(newPassword); err != nil {
		return fmt.Errorf("password validation failed: %w", err)
	}

	tokenUUID, err := uuid.Parse(token)
	if err != nil {
		return fmt.Errorf("invalid reset token format")
	}

	// Get password reset record
	resetRecord, err := s.passwordResetRepo.GetByToken(ctx, tokenUUID)
	if err != nil {
		return fmt.Errorf("invalid or expired reset token")
	}

	// Check if token is expired
	if resetRecord.ExpiresAt.Before(time.Now()) {
		return fmt.Errorf("reset token has expired")
	}

	// Hash new password
	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	// Fetch user details to get email for cross-table sync
	user, err := s.userRepo.GetByID(ctx, resetRecord.UserID)
	if err != nil {
		return fmt.Errorf("failed to fetch user for sync: %w", err)
	}

	// Update password in users table
	if err := s.userRepo.UpdatePassword(ctx, resetRecord.UserID, hashedPassword); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	// TRIGGER SYNC: Update password in tenants and organizations tables if email matches
	s.logger.Info("Syncing password across platforms", "email", user.Email)
	
	if err := s.userRepo.UpdateTenantPassword(ctx, user.Email, hashedPassword); err != nil {
		s.logger.Warn("Failed to sync tenant password", "email", user.Email, "error", err)
	}

	if err := s.userRepo.UpdateOrganizationSuperAdminPassword(ctx, user.Email, hashedPassword); err != nil {
		s.logger.Warn("Failed to sync organization super admin password", "email", user.Email, "error", err)
	}
	// Delete reset token
	if err := s.passwordResetRepo.Delete(ctx, tokenUUID); err != nil {
		s.logger.Warn("Failed to delete reset token", "error", err)
	}

	// Invalidate all sessions for security
	if err := s.InvalidateAllSessions(ctx, resetRecord.UserID); err != nil {
		s.logger.Warn("Failed to invalidate sessions after password reset", "user_id", resetRecord.UserID, "error", err)
	}

	return nil
}

// InvalidateAllSessions invalidates all sessions for a user
func (s *authService) InvalidateAllSessions(ctx context.Context, userID uuid.UUID) error {
	// Get all sessions for user
	sessions, err := s.GetActiveSessions(ctx, userID)
	if err != nil {
		return fmt.Errorf("failed to get sessions: %w", err)
	}

	// Delete each session
	for _, session := range sessions {
		if err := s.sessionRepo.Delete(ctx, session.SessionID); err != nil {
			s.logger.Warn("Failed to delete session during mass invalidation", "session_id", session.SessionID, "error", err)
		}
	}

	return nil
}

// GetActiveSessions gets all active sessions for a user
func (s *authService) GetActiveSessions(ctx context.Context, userID uuid.UUID) ([]*domain.Session, error) {
	// Use sessionRepo's GetByUserID method
	return s.sessionRepo.GetByUserID(ctx, userID)
}

// SwitchOrganization switches a user to a different organization within the same tenant
func (s *authService) SwitchOrganization(ctx context.Context, userID uuid.UUID, newOrgID uuid.UUID, tenantID uuid.UUID) (*domain.LoginResponse, error) {
	// Verify user exists and belongs to the tenant
	user, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}

	if user.TenantID != tenantID {
		return nil, fmt.Errorf("user does not belong to the specified tenant")
	}

	// Verify the organization exists and belongs to the SAME tenant (critical validation)
	org, err := s.orgClient.GetOrganization(s.mergeMetadata(ctx, nil), newOrgID)
	if err != nil {
		return nil, fmt.Errorf("organization not found: %w", err)
	}

	if org.TenantID != tenantID {
		return nil, fmt.Errorf("organization does not belong to the specified tenant - cross-tenant switching not allowed")
	}

	// Additional validation: ensure user has access to organizations within the same tenant
	// Get user's current organizations to validate switching rights
	userOrgs, err := s.orgClient.ListUserOrganizations(s.mergeMetadata(ctx, nil), userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user organizations: %w", err)
	}

	// Check if user has access to the target organization
	// User can switch to any organization they belong to within the same tenant
	hasAccess := false
	for _, userOrg := range userOrgs {
		if userOrg.OrgID == newOrgID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		return nil, fmt.Errorf("user does not have access to switch to this organization")
	}

	// Invalidate all existing sessions for the user (AGGRESSIVE - DISABLE FOR DEBUG)
	/*
	if err := s.InvalidateAllSessions(ctx, userID); err != nil {
		s.logger.Warn("Failed to invalidate sessions during org switch", "user_id", userID, "error", err)
	}
	*/

	// Fetch roles and permissions for the new organization context
	// Ensure we forward incoming Authorization header for this protected operation
	ctx = s.mergeMetadata(ctx, metadata.Pairs(
		"tenant_id", tenantID.String(),
		"org_id", newOrgID.String(),
	))
	
	rolesResp, err := s.userServiceClient.ListRolesOfUser(ctx, &userpb.GetUserRequest{
		UserId: userID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user roles for token: %w", err)
	}

	roleNames := make([]string, 0)
	permSet := make(map[string]struct{})
	newOrgIDStr := newOrgID.String()

	for _, r := range rolesResp.Roles {
		// Include roles that belong to this org OR are tenant-level (empty org_id)
		if r.OrgId == "" || r.OrgId == newOrgIDStr {
			if r.Name != "" {
				roleNames = append(roleNames, r.Name)
			}
			for _, p := range r.Permissions {
				permSet[p] = struct{}{}
			}
		}
	}
	permissions := make([]string, 0, len(permSet))
	for p := range permSet {
		permissions = append(permissions, p)
	}

	// Generate new tokens for the new organization context
	// We must do this BEFORE creating the session so we can store the token in the session
	token, err := s.jwtManager.GenerateToken(userID, user.Email, user.Name, tenantID, &newOrgID, roleNames, permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Create new session with the new organization context
	sessionID := uuid.New()
	expiresAt := time.Now().Add(24 * time.Hour) // 24 hour session

	session := &domain.Session{
		SessionID:    sessionID,
		UserID:       userID,
		SessionToken: token, // Store the actual token!
		CreatedAt:    time.Now(),
		ExpiresAt:    expiresAt,
	}

	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(userID.String(), tenantID.String(), newOrgID.String())
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	refreshTokenRecord := &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    userID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}

	if err := s.refreshTokenRepo.Create(ctx, refreshTokenRecord); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	s.logger.Info("User switched organization", "user_id", userID, "new_org_id", newOrgID, "tenant_id", tenantID)

	return &domain.LoginResponse{
		UserID:           userID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		TenantID:         tenantID,
		OrgID:            &newOrgID,
		Token:            token,
		RefreshToken:     refreshToken,
		TokenExpiresAt:   expiresAt.Unix(),
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "", // Will be set by middleware
	}, nil
}

// InitiateSSO initiates the SSO flow by returning the provider's auth URL
func (s *authService) InitiateSSO(ctx context.Context, provider string) (string, error) {
	var ssoProvider SSOProvider
	state := uuid.New().String() // Simple state generation, should optionally be stored/validated

	switch strings.ToLower(provider) {
	case "google":
		ssoProvider = NewGoogleProvider(s.ssoConfig, s.logger)
	case "microsoft", "azure":
		ssoProvider = NewMicrosoftProvider(s.ssoConfig, s.logger)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}

	return ssoProvider.GetAuthURL(state), nil
}

// CompleteSSO handles the callback from the SSO provider
func (s *authService) CompleteSSO(ctx context.Context, provider, code string) (*domain.LoginResponse, error) {
	var ssoProvider SSOProvider

	switch strings.ToLower(provider) {
	case "google":
		ssoProvider = NewGoogleProvider(s.ssoConfig, s.logger)
	case "microsoft", "azure":
		ssoProvider = NewMicrosoftProvider(s.ssoConfig, s.logger)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", provider)
	}

	// 1. Exchange code for token
	token, err := ssoProvider.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}

	// 2. Get User Info
	userInfo, err := ssoProvider.GetUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}

	s.logger.Info("SSO Success", "provider", provider, "name", userInfo.Name, "email", userInfo.Email)

	// 3. Login or Register
	// Logic: Try to login if user exists. If not, fail (or auto-register if enabled).
	// For production strictness, we search Globally.
	
	// We use LoginGlobal logic here but bypass password check
	// Check if user exists by email globally
	user, err := s.userRepo.GetByEmailGlobal(ctx, userInfo.Email)
	if err != nil {
		// User not found
		// For now, we return error. To enable auto-registration, we would call CreateUser here.
		return nil, fmt.Errorf("user with email %s not found in the system. Please register first", userInfo.Email)
	}

	// User found - Proceed to Login
	
	// Ensure email is verified (implicitly yes since it came from SSO, but let's be safe)
	if user.EmailVerifiedAt == nil {
		// Auto-verify email since identity provider verified it
		if err := s.userRepo.VerifyEmail(ctx, user.UserID); err != nil {
			s.logger.Warn("Failed to auto-verify email for SSO user", "user_id", user.UserID, "error", err)
		}
	}

	// Update last login
	if err := s.userRepo.UpdateLastLogin(ctx, user.UserID, "", ""); err != nil {
		s.logger.Warn("Failed to update last login for SSO user", "user_id", user.UserID, "error", err)
	}

	// Auto-detect organization logic (copied from LoginGlobal)
	var orgID *uuid.UUID
	orgsResp, err := s.userServiceClient.ListUserOrganizations(s.mergeMetadata(ctx, nil), &userpb.ListUserOrganizationsRequest{
		UserId: user.UserID.String(),
	})
	if err == nil && len(orgsResp.Organizations) > 0 {
		var selected *userpb.UserOrganizationInfo
		for _, o := range orgsResp.Organizations {
			if o.IsCurrentContext {
				selected = o
				break
			}
		}
		if selected == nil {
			selected = orgsResp.Organizations[0]
		}
		if selected.OrgId != "" {
			if parsedOrgID, err := uuid.Parse(selected.OrgId); err == nil {
				orgID = &parsedOrgID
			}
		}
	}

	if orgID == nil {
		return nil, fmt.Errorf("no organization found for this user")
	}

	orgIDStr := orgID.String()

	// Generate Tokens
	rolesResp, err := s.userServiceClient.ListRolesOfUser(s.mergeMetadata(ctx, nil), &userpb.GetUserRequest{UserId: user.UserID.String()})
	var roleNames []string
	var permissions []string
	if err == nil {
		permSet := make(map[string]struct{})
		for _, r := range rolesResp.Roles {
			if r.Name != "" { roleNames = append(roleNames, r.Name) }
			for _, p := range r.Permissions { permSet[p] = struct{}{} }
		}
		for p := range permSet { permissions = append(permissions, p) }
	}

	accessToken, accessExpiresAt, err := s.jwtManager.GenerateAccessToken(
		user.UserID.String(), user.Email, user.Name, user.TenantID.String(), orgIDStr, roleNames, permissions,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}

	refreshToken, refreshExpiresAt, err := s.jwtManager.GenerateRefreshToken(user.UserID.String(), user.TenantID.String(), orgIDStr)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}

	// Store refresh token
	if err := s.refreshTokenRepo.Create(ctx, &domain.RefreshToken{
		Token:     refreshToken,
		UserID:    user.UserID,
		ExpiresAt: time.Unix(refreshExpiresAt, 0),
		CreatedAt: time.Now(),
	}); err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Create session
	session := &domain.Session{
		SessionID:    uuid.New(),
		UserID:       user.UserID,
		SessionToken: accessToken,
		CreatedAt:    time.Now(),
		ExpiresAt:    time.Unix(accessExpiresAt, 0),
	}
	createdSession, err := s.sessionRepo.Create(ctx, session)
	if err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &domain.LoginResponse{
		Token:            accessToken,
		RefreshToken:     refreshToken,
		UserID:           user.UserID,
		Email:            user.Email,
		Name:             user.Name,
		Roles:            roleNames,
		Permissions:      permissions,
		LastLoginAt:      time.Now(),
		LastLoginIP:      "",
		TenantID:         user.TenantID,
		OrgID:            orgID,
		TokenExpiresAt:   accessExpiresAt,
		RefreshExpiresAt: refreshExpiresAt,
		SessionID:        createdSession.SessionID,
	}, nil
}

// mergeMetadata merges incoming metadata with new keys for outgoing calls
func (s *authService) mergeMetadata(ctx context.Context, additional metadata.MD) context.Context {
	incoming, _ := metadata.FromIncomingContext(ctx)
	
	// Create map to hold merged metadata
	merged := make(metadata.MD)
	
	// Copy all incoming metadata
	for k, v := range incoming {
		merged[k] = v
	}
	
	// Overwrite/Add additional metadata
	for k, v := range additional {
		merged[k] = v
	}
	
	return metadata.NewOutgoingContext(ctx, merged)
}
