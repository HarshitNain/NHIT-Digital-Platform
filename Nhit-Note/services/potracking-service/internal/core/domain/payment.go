package domain

import (
	"time"
)

// Payment represents a payment domain model
type Payment struct {
	ID                  int64
	SlNo                string
	TemplateType        string
	Project             *string
	AccountFullName     *string
	FromAccountType     *string
	FullAccountNumber   *string
	ToAccount           *string
	ToAccountType       *string
	NameOfBeneficiary   *string
	AccountNumber       *string
	NameOfBank          *string
	IfscCode            *string
	Amount              float64
	Purpose             *string
	Status              string
	UserID              string
	PaymentNoteID       *int64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// PaymentGroup represents a group of payments with the same sl_no
type PaymentGroup struct {
	SlNo      string
	Status    string
	Payments  []Payment
	CreatedAt time.Time
}

// PaymentVendorAccount represents the many-to-many relationship
type PaymentVendorAccount struct {
	ID               int64
	PaymentID        int64
	VendorID         string
	VendorAccountID  *int64
	CreatedAt        time.Time
}

// PaymentShortcut represents a saved payment template
type PaymentShortcut struct {
	ID              int64
	SlNo            *string
	ShortcutName    string
	RequestDataJSON string
	UserID          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// BankLetterApprovalLog represents approval log for bank letters
type BankLetterApprovalLog struct {
	ID            int64
	SlNo          string
	Status        string
	Comments      *string
	ReviewerID    string
	ReviewerName  *string
	ReviewerEmail *string
	ApproverLevel *int32
	CreatedAt     time.Time
}

// PaymentFilters represents filter criteria for listing payments
type PaymentFilters struct {
	Status       *string
	OnlyAssigned bool
	UserID       string
	Search       *string
	Page         int32
	PerPage      int32
}
