package ports

import (
	"context"

	"github.com/ShristiRnr/NHIT_Backend/services/vendor-service/internal/core/domain"
	"github.com/google/uuid"
	"log/slog"
)

// VendorService defines the business logic contract (Hexagonal Architecture - Port)
type VendorService interface {
	// Vendor business operations (equivalent to PHP VendorService)
	CreateVendor(ctx context.Context, params domain.CreateVendorParams) (*domain.Vendor, error)
	GetVendorByID(ctx context.Context, orgID, vendorID uuid.UUID) (*domain.Vendor, error)
	GetVendorByCode(ctx context.Context, orgID uuid.UUID, vendorCode string) (*domain.Vendor, error)
	UpdateVendor(ctx context.Context, orgID, vendorID uuid.UUID, params domain.UpdateVendorParams) (*domain.Vendor, error)
	DeleteVendor(ctx context.Context, orgID, vendorID uuid.UUID) error
	ListVendors(ctx context.Context, tenantID uuid.UUID, filters VendorListFilters) ([]*domain.Vendor, int64, error)
	
	// Vendor code operations (equivalent to PHP VendorService code methods)
	GenerateVendorCode(ctx context.Context, vendorName string, vendorType *string) (string, error)
	UpdateVendorCode(ctx context.Context, orgID, vendorID uuid.UUID, newCode string) (*domain.Vendor, error)
	RegenerateVendorCode(ctx context.Context, orgID, vendorID uuid.UUID) (*domain.Vendor, error)
	
	// Vendor account business operations (equivalent to PHP VendorService account methods)
	CreateVendorAccount(ctx context.Context, params domain.CreateVendorAccountParams) (*domain.VendorAccount, error)
	GetVendorAccounts(ctx context.Context, vendorID uuid.UUID) ([]*domain.VendorAccount, error)
	GetVendorBankingDetails(ctx context.Context, vendorID uuid.UUID, accountID *uuid.UUID) (*BankingDetails, error)
	UpdateVendorAccount(ctx context.Context, accountID uuid.UUID, params domain.UpdateVendorAccountParams) (*domain.VendorAccount, error)
	DeleteVendorAccount(ctx context.Context, accountID uuid.UUID) error
	ToggleAccountStatus(ctx context.Context, accountID uuid.UUID) (*domain.VendorAccount, error)
	SetPrimaryAccount(ctx context.Context, accountID uuid.UUID) (*domain.VendorAccount, error)
}

// BankingDetails represents banking information for payment processing
type BankingDetails struct {
	AccountName   string  `json:"account_name"`
	AccountNumber string  `json:"account_number"`
	AccountType   *string `json:"account_type,omitempty"`
	NameOfBank    string  `json:"name_of_bank"`
	BranchName    *string `json:"branch_name,omitempty"`
	IFSCCode      string  `json:"ifsc_code"`
	SwiftCode     *string `json:"swift_code,omitempty"`
	Remarks       *string `json:"remarks,omitempty"`
}

// VendorServiceConfig holds configuration for vendor service
type VendorServiceConfig struct {
	EnableCodeGeneration bool
	DefaultVendorType    string
	MaxAccountsPerVendor int
}

// Logger defines logging contract (deprecated in favor of slog.Logger)
type Logger = *slog.Logger

// EventPublisher defines event publishing contract
type EventPublisher interface {
	PublishVendorCreated(ctx context.Context, vendor *domain.Vendor) error
	PublishVendorUpdated(ctx context.Context, vendor *domain.Vendor) error
	PublishVendorDeleted(ctx context.Context, tenantID, vendorID uuid.UUID) error
	PublishAccountCreated(ctx context.Context, account *domain.VendorAccount) error
	PublishAccountUpdated(ctx context.Context, account *domain.VendorAccount) error
	PublishAccountDeleted(ctx context.Context, accountID uuid.UUID) error
}
