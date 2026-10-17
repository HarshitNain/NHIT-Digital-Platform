package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/google/uuid"
)

// VendorRepository defines the contract for vendor data persistence (Hexagonal Architecture - Port)
type VendorRepository interface {
	// Vendor CRUD operations
	CreateVendor(ctx context.Context, vendor *domain.Vendor) error
	GetVendorByID(ctx context.Context, orgID, vendorID uuid.UUID) (*domain.Vendor, error)
	GetVendorByCode(ctx context.Context, orgID uuid.UUID, vendorCode string) (*domain.Vendor, error)
	GetVendorByEmail(ctx context.Context, orgID uuid.UUID, email string) (*domain.Vendor, error)
	UpdateVendor(ctx context.Context, vendor *domain.Vendor) error
	DeleteVendor(ctx context.Context, orgID, vendorID uuid.UUID) error
	ListVendors(ctx context.Context, tenantID uuid.UUID, filters VendorListFilters) ([]*domain.Vendor, int64, error)
	
	// Vendor validation operations
	IsVendorCodeExists(ctx context.Context, orgID uuid.UUID, code string, excludeID *uuid.UUID) (bool, error)
	IsVendorEmailExists(ctx context.Context, orgID uuid.UUID, email string, excludeID *uuid.UUID) (bool, error)
	
	// Vendor account CRUD operations
	CreateVendorAccount(ctx context.Context, account *domain.VendorAccount) error
	GetVendorAccountByID(ctx context.Context, accountID uuid.UUID) (*domain.VendorAccount, error)
	GetVendorAccountsByVendorID(ctx context.Context, vendorID uuid.UUID) ([]*domain.VendorAccount, error)
	GetPrimaryVendorAccount(ctx context.Context, vendorID uuid.UUID) (*domain.VendorAccount, error)
	UpdateVendorAccount(ctx context.Context, account *domain.VendorAccount) error
	DeleteVendorAccount(ctx context.Context, accountID uuid.UUID) error
	UnsetPrimaryVendorAccounts(ctx context.Context, vendorID uuid.UUID, excludeAccountID *uuid.UUID) error
	
	// Transaction support for business operations
	WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

// VendorListFilters represents filters for listing vendors
type VendorListFilters struct {
	IsActive       *bool
	VendorType     *string
	OrgID          *uuid.UUID // Filter by organization (meta-filter)
	ProjectIDs     []string   // List of project IDs belonging to the organization
	Project        *string   // Specific project filter
	Search         *string
	Limit          int
	Offset         int
}

// DatabaseRepository defines database-specific operations
type DatabaseRepository interface {
	// Health check
	Ping(ctx context.Context) error
	
	// Connection management
	Close() error
}
