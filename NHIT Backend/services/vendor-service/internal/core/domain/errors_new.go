package domain

import "errors"

// Domain errors for vendor business logic
var (
	// Vendor validation errors
	ErrInvalidTenantID        = errors.New("invalid tenant ID")
	ErrInvalidVendorID        = errors.New("invalid vendor ID")
	ErrInvalidVendorName      = errors.New("invalid vendor name")
	ErrInvalidVendorEmail     = errors.New("invalid vendor email")
	ErrInvalidVendorCode      = errors.New("invalid vendor code")
	ErrInvalidEmailFormat     = errors.New("invalid email format")
	ErrInvalidPAN             = errors.New("invalid PAN")
	ErrInvalidPANFormat       = errors.New("invalid PAN format")
	ErrInvalidBeneficiaryName = errors.New("invalid beneficiary name")
	ErrInvalidCreatedBy       = errors.New("invalid created by user ID")
	
	// Account validation errors
	ErrInvalidAccountID       = errors.New("invalid account ID")
	ErrInvalidAccountName     = errors.New("invalid account name")
	ErrInvalidAccountNumber   = errors.New("invalid account number")
	ErrInvalidBankName        = errors.New("invalid bank name")
	ErrInvalidIFSCCode        = errors.New("invalid IFSC code")
	ErrInvalidIFSCFormat      = errors.New("invalid IFSC code format")
	
	// Business logic errors
	ErrVendorNotFound         = errors.New("vendor not found")
	ErrVendorAccountNotFound  = errors.New("vendor account not found")
	ErrVendorCodeExists       = errors.New("vendor code already exists")
	ErrVendorEmailExists      = errors.New("vendor email already exists")
	ErrNoPrimaryAccount       = errors.New("no primary account found")
	ErrCannotDeletePrimaryAccount = errors.New("cannot delete primary account without reassignment")
	
	// Transaction errors
	ErrTransactionFailed      = errors.New("transaction failed")
	ErrDatabaseConnection     = errors.New("database connection error")
)
