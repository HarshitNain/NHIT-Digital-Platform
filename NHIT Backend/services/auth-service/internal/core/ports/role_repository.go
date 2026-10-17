package ports

import (
	"context"

	"github.com/google/uuid"
)

// Role represents a role entity (simplified version for auth service)
type Role struct {
	RoleID       uuid.UUID
	TenantID     uuid.UUID
	OrgID        *uuid.UUID
	Name         string
	Description  string
	Permissions  []string
	IsSystemRole bool
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role interface{}) (*Role, error)
	GetByID(ctx context.Context, roleID uuid.UUID) (*Role, error)
	GetByName(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name string) (*Role, error)
	Update(ctx context.Context, role interface{}) (*Role, error)
	Delete(ctx context.Context, roleID uuid.UUID) error
	ListByOrganization(ctx context.Context, orgID uuid.UUID, limit, offset int32) ([]*Role, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*Role, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID, orgID uuid.UUID) error
	RemoveRoleFromUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) ([]*Role, error)
	GetRolePermissions(ctx context.Context, roleID uuid.UUID) ([]string, error)
}
