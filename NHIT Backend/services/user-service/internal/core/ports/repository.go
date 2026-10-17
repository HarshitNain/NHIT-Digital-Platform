package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/user-service/internal/core/domain"
	"github.com/google/uuid"
)

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *domain.User) (*domain.User, error)
	GetByID(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) (*domain.User, error)
	GetByEmail(ctx context.Context, tenantID uuid.UUID, email string, orgID *uuid.UUID) (*domain.User, error)
	Update(ctx context.Context, user *domain.User) (*domain.User, error)
	Delete(ctx context.Context, userID uuid.UUID, orgID *uuid.UUID) error
	ListByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	ListAll(ctx context.Context, limit, offset int32) ([]*domain.User, int64, error)
	CountByTenant(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID) (int64, error)
	ListByDepartment(ctx context.Context, departmentID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	ListByDesignation(ctx context.Context, designationID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	SearchByName(ctx context.Context, name string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	SearchByEmail(ctx context.Context, email string, orgID *uuid.UUID, limit, offset int32) ([]*domain.User, int64, error)
	UpdateLastLogin(ctx context.Context, userID uuid.UUID, lastLoginIP, userAgent string) error
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// TenantRepository defines the interface for tenant data operations
type TenantRepository interface {
	Create(ctx context.Context, tenant *domain.Tenant) (*domain.Tenant, error)
	GetByID(ctx context.Context, tenantID uuid.UUID) (*domain.Tenant, error)
}

// UserRoleRepository defines the interface for user-role operations
type UserRoleRepository interface {
	AssignRole(ctx context.Context, userID, roleID uuid.UUID) error
	RemoveRole(ctx context.Context, userID, roleID uuid.UUID) error
	ListRolesByUser(ctx context.Context, userID uuid.UUID) ([]*domain.Role, error)
}

// RoleRepository defines the interface for role data operations
type RoleRepository interface {
	Create(ctx context.Context, role *domain.Role) (*domain.Role, error)
	GetByID(ctx context.Context, roleID uuid.UUID) (*domain.Role, error)
	ListByTenant(ctx context.Context, tenantID uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error)
	ListByTenantAndOrgIncludingSystem(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, limit, offset int32) ([]*domain.Role, int64, error)
	Update(ctx context.Context, role *domain.Role) (*domain.Role, error)
	Delete(ctx context.Context, roleID uuid.UUID) error
	Exists(ctx context.Context, tenantID uuid.UUID, orgID *uuid.UUID, name string) (bool, error)
	ExistsSystem(ctx context.Context, tenantID uuid.UUID, name string) (bool, error)
}

// PermissionRepository defines the interface for permission catalog operations
type PermissionRepository interface {
	ListAll(ctx context.Context) ([]*domain.Permission, error)
	ListByModule(ctx context.Context, module *string) ([]*domain.Permission, error)
}

// LoginHistoryRepository defines the interface for login history data operations
type LoginHistoryRepository interface {
	Create(ctx context.Context, history *domain.UserLoginHistory) (*domain.UserLoginHistory, error)
	ListByUser(ctx context.Context, userID uuid.UUID, limit, offset int32) ([]*domain.UserLoginHistory, int64, error)
}

// ActivityLogRepository defines the interface for activity log data operations
type ActivityLogRepository interface {
	Create(ctx context.Context, log *domain.ActivityLog) (*domain.ActivityLog, error)
	List(ctx context.Context, limit, offset int32) ([]*domain.ActivityLog, int64, error)
	Count(ctx context.Context) (int64, error)
}
