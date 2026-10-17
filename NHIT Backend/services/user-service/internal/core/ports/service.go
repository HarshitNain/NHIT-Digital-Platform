package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/google/uuid"
)

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	GetUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*domain.User, error)
	UpdateUser(ctx context.Context, user *domain.User) (*domain.User, error)
	DeleteUser(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error
	ListUsersByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	ListUsersByOrganization(ctx context.Context, tenantID, orgID uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	ListUsersByDepartment(ctx context.Context, departmentID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	ListUsersByDesignation(ctx context.Context, designationID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	SearchUsersByName(ctx context.Context, name string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	SearchUsersByEmail(ctx context.Context, email string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	AssignRoleToUser(ctx context.Context, userID, roleID uuid.UUID) error
	GetUserRoles(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)

	// Role management
	CreateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetRole(ctx context.Context, roleID uuid.UUID) (*domain.Role, error)
	ListRolesByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error)
	ListRolesByOrganization(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, includeSystem bool, limit, offset int32) ([]*domain.Role, int64, error)
	UpdateRole(ctx context.Context, role *domain.Role) (*domain.Role, error)
	DeleteRole(ctx context.Context, roleID uuid.UUID) error

	// Permission catalog (fixed list)
	ListPermissions(ctx context.Context, module *string) ([]*domain.Permission, error)

	// Soft delete operations
	DeactivateUser(ctx context.Context, userID, deactivatedBy uuid.UUID, reason string) (*domain.User, error)
	ReactivateUser(ctx context.Context, userID, reactivatedBy uuid.UUID) (*domain.User, error)

	// Activity logging
	CreateActivityLog(ctx context.Context, log *domain.ActivityLog) (*domain.ActivityLog, error)
	ListActivityLogs(ctx context.Context, limit, offset int32) ([]*domain.ActivityLog, int64, error)



	// Login history
	CreateLoginHistory(ctx context.Context, history *domain.UserLoginHistory) (*domain.UserLoginHistory, error)
	ListLoginHistory(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.UserLoginHistory, int64, error)

	// Tenant management
	CreateTenant(ctx context.Context, name, email, password, role string) (*domain.Tenant, error)
	GetTenant(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
}
