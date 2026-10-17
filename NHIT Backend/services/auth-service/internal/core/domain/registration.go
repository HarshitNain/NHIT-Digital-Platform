package domain

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrInvalidEmail          = errors.New("invalid email format")
	ErrWeakPassword          = errors.New("password must be at least 8 characters with uppercase, lowercase, number, and special character")
	ErrInvalidOrganizationName = errors.New("organization name must be between 3 and 255 characters")
	ErrRegistrationFailed    = errors.New("registration failed")
)

// User represents a user in the auth context (simplified version)
type User struct {
	UserID    uuid.UUID
	TenantID  uuid.UUID
	Name      string
	Email     string
	Password  string
	IsActive  bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Role represents a role with permissions
type Role struct {
	RoleID       uuid.UUID
	TenantID     uuid.UUID
	OrgID        *uuid.UUID
	Name         string
	Description  string
	Permissions  []string
	IsSystemRole bool
	CreatedBy    *uuid.UUID
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// RegistrationRequest represents a new user registration request
type RegistrationRequest struct {
	Name             string
	Email            string
	Password         string
	OrganizationName string
	OrganizationCode string
	PhoneNumber      *string
}

// RegistrationResponse represents the response after successful registration
type RegistrationResponse struct {
	UserID           uuid.UUID
	Email            string
	Name             string
	OrganizationID   uuid.UUID
	OrganizationName string
	IsSuperAdmin     bool
	Token            string
	RefreshToken     string
	Message          string
}

// SuperAdminRole represents the super admin role details
type SuperAdminRole struct {
	RoleID      uuid.UUID
	RoleName    string
	Permissions []string
}

// LoginHistory represents a login record
type LoginHistory struct {
	HistoryID     uuid.UUID
	UserID        uuid.UUID
	LoginTime     time.Time
	IPAddress     string
	UserAgent     string
	LoginMethod   string // JWT, SSO_GOOGLE, SSO_MICROSOFT
	IsSuccessful  bool
	FailureReason *string
	SessionID     *uuid.UUID
}

// ActivityLog represents a system activity log
type ActivityLog struct {
	LogID        uuid.UUID
	UserID       uuid.UUID
	Action       string // CREATE_USER, UPDATE_ROLE, SWITCH_ORG, etc.
	ResourceType string // USER, ORGANIZATION, PROJECT, etc.
	ResourceID   *string
	Details      map[string]interface{}
	IPAddress    string
	UserAgent    string
	TenantID     uuid.UUID
	OrgID        *uuid.UUID
	CreatedAt    time.Time
}

// OrganizationSwitch represents an organization context switch
type OrganizationSwitch struct {
	SwitchID       uuid.UUID
	UserID         uuid.UUID
	FromOrgID      *uuid.UUID
	ToOrgID        uuid.UUID
	SwitchTime     time.Time
	IPAddress      string
	UserAgent      string
}

// SSOProvider represents SSO provider types
type SSOProvider string

const (
	SSOProviderGoogle    SSOProvider = "GOOGLE"
	SSOProviderMicrosoft SSOProvider = "MICROSOFT"
)

// SSOLoginRequest represents SSO login request
type SSOLoginRequest struct {
	Provider     SSOProvider
	IDToken      string
	Email        string
	Name         string
	Picture      *string
	TenantID     *uuid.UUID
	OrgID        *uuid.UUID
}

// NewLoginHistory creates a new login history record
func NewLoginHistory(userID uuid.UUID, ipAddress, userAgent, loginMethod string, isSuccessful bool, sessionID *uuid.UUID) *LoginHistory {
	return &LoginHistory{
		HistoryID:    uuid.New(),
		UserID:       userID,
		LoginTime:    time.Now(),
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		LoginMethod:  loginMethod,
		IsSuccessful: isSuccessful,
		SessionID:    sessionID,
	}
}

// NewActivityLog creates a new activity log
func NewActivityLog(userID uuid.UUID, action, resourceType string, resourceID *string, details map[string]interface{}, ipAddress, userAgent string, tenantID uuid.UUID, orgID *uuid.UUID) *ActivityLog {
	return &ActivityLog{
		LogID:        uuid.New(),
		UserID:       userID,
		Action:       action,
		ResourceType: resourceType,
		ResourceID:   resourceID,
		Details:      details,
		IPAddress:    ipAddress,
		UserAgent:    userAgent,
		TenantID:     tenantID,
		OrgID:        orgID,
		CreatedAt:    time.Now(),
	}
}

// NewOrganizationSwitch creates a new organization switch record
func NewOrganizationSwitch(userID uuid.UUID, fromOrgID *uuid.UUID, toOrgID uuid.UUID, ipAddress, userAgent string) *OrganizationSwitch {
	return &OrganizationSwitch{
		SwitchID:   uuid.New(),
		UserID:     userID,
		FromOrgID:  fromOrgID,
		ToOrgID:    toOrgID,
		SwitchTime: time.Now(),
		IPAddress:  ipAddress,
		UserAgent:  userAgent,
	}
}
