package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents the core user domain model
type User struct {
	UserID          uuid.UUID
	TenantID        uuid.UUID
	OrgID           *uuid.UUID // Organization assignment
	Name            string
	Email           string
	Password        string
	EmailVerifiedAt *time.Time
	LastLoginAt     *time.Time
	LastLogoutAt    *time.Time
	LastLoginIP     string
	UserAgent       string
	DepartmentID    *uuid.UUID // Department assignment
	DesignationID   *uuid.UUID // Designation assignment
	
	// Banking Information
	AccountHolderName  *string
	BankName           *string
	BankAccountNumber  *string
	IFSCCode           *string
	
	// Signature
	SignatureURL       *string
	
	// Roles (Optional for updates)
	Roles           []uuid.UUID

	IsActive        bool       // For soft delete
	DeactivatedAt   *time.Time // When user was deactivated
	DeactivatedBy   *uuid.UUID // Who deactivated the user
	DeactivatedByName *string  // Name of the user who deactivated
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// UserRole represents user-role association
type UserRole struct {
	UserID uuid.UUID
	RoleID uuid.UUID
}

// Role represents a role in the system (Dynamic & Organization-specific)
type Role struct {
	RoleID       uuid.UUID
	TenantID     uuid.UUID
	OrgID        *uuid.UUID // Organization-specific roles (nil for system roles)
	Name         string
	Description  string
	Permissions  []string
	IsSystemRole bool   // System vs custom role
	CreatedBy    string // Super admin name who created role
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Permission represents individual permissions
type Permission struct {
	PermissionID       uuid.UUID
	Name               string // e.g., "users.create", "projects.view"
	Description        string // Human readable description
	Module             string // e.g., "users", "projects", "reports"
	Action             string // e.g., "create", "read", "update", "delete"
	IsSystemPermission bool   // System vs custom permission
}

// UserOrganizationRole represents user's role within specific organization
type UserOrganizationRole struct {
	UserID           uuid.UUID
	OrgID            uuid.UUID
	RoleID           uuid.UUID
	IsCurrentContext bool      // Whether this is active organization
	AssignedBy       uuid.UUID // Super admin who assigned this role
	AssignedAt       time.Time
	UpdatedAt        time.Time
}




// UserLoginHistory represents user login tracking
type UserLoginHistory struct {
	HistoryID uuid.UUID
	UserID    uuid.UUID
	IPAddress *string
	UserAgent *string
	LoginTime time.Time
}

// ActivityLog represents user activity tracking for audit trail
type ActivityLog struct {
	ID          int32
	Name        string
	Description string
	CreatedAt   time.Time
}

// Tenant represents a tenant in the multi-tenant system
type Tenant struct {
	TenantID  uuid.UUID
	Name      string // Super admin name (for tenant creation)
	Email     string // Super admin email
	Password  string // Super admin password (hashed)
	CreatedAt time.Time
	UpdatedAt time.Time
}
