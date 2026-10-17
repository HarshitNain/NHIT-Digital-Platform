package domain

import (
	"time"
)

// PaymentNote represents a payment note domain model
type PaymentNote struct {
	ID                    int64
	UserID                int64
	
	// Green Note Reference
	GreenNoteID           *string
	GreenNoteNo           *string
	GreenNoteApprover     *string
	GreenNoteAppDate      *string
	
	// Reimbursement Reference
	ReimbursementNoteID   *int64
	
	// Payment Note Details
	NoteNo                string
	Subject               *string
	Date                  *time.Time
	Department            *string
	
	// Vendor Details
	VendorCode            *string
	VendorName            *string
	
	// Project Details
	ProjectName           *string
	
	// Invoice Details
	InvoiceNo             *string
	InvoiceDate           *string
	InvoiceAmount         float64
	InvoiceApprovedBy     *string
	
	// LOA/PO Details
	LoaPoNo               *string
	LoaPoAmount           float64
	LoaPoDate             *string
	
	// Financial Calculations
	GrossAmount           float64
	TotalAdditions        float64
	TotalDeductions       float64
	NetPayableAmount      float64
	NetPayableRoundOff    float64
	NetPayableWords       *string
	
	// TDS Details
	TdsPercentage         float64
	TdsSection            *string
	TdsAmount             float64
	
	// Bank Details
	AccountHolderName     *string
	BankName              *string
	AccountNumber         *string
	IfscCode              *string
	
	// Recommendation
	RecommendationOfPayment *string
	
	// Status and Flags
	Status                string
	IsDraft               bool
	AutoCreated           bool
	CreatedBy             *int64
	CreatedByName         *string
	CreatedByDesignation  *string
	CreatedByRole         *string
	CreatedBySignature    *string
	
	// Hold Information
	HoldReason            *string
	HoldDate              *time.Time
	HoldBy                *int64
	
	// UTR Information
	UtrNo                 *string
	UtrDate               *string
	
	// Financial Breakdown
	TaxableAmount         float64
	GstAmount             float64
	OtherCharges          float64
	GreenNoteHistory      *string
	GreenNoteURL          *string
	
	// Timestamps
	CreatedAt             time.Time
	UpdatedAt             time.Time
	
	// Related entities
	AddParticulars        []PaymentParticular
	LessParticulars       []PaymentParticular
	ApprovalLogs          []PaymentApprovalLog
	Comments              []PaymentComment
	Documents             []PaymentNoteDocument
}

// PaymentParticular represents an add/less particular item
type PaymentParticular struct {
	ID             int64
	PaymentNoteID  int64
	ParticularType string // 'ADD' or 'LESS'
	Particular     string
	Amount         float64
	CreatedAt      time.Time
}

// PaymentApprovalLog represents an approval log entry
type PaymentApprovalLog struct {
	ID            int64
	PaymentNoteID int64
	Status        string
	Comments      *string
	ReviewerID    int64
	ReviewerName  *string
	ReviewerEmail *string
	ApproverLevel *int32
	ReviewerDesignation *string
	ReviewerRole        *string
	ReviewerSignature   *string
	CreatedAt     time.Time
}

// PaymentComment represents a comment on a payment note
type PaymentComment struct {
	ID            int64
	PaymentNoteID int64
	Comment       string
	Status        *string
	UserID        int64
	UserName      *string
	UserEmail     *string
	UserDesignation *string
	UserRole        *string
	UserSignature   *string
	CreatedAt     time.Time
}

// PaymentNoteDocument represents a supporting document
type PaymentNoteDocument struct {
	ID               int64
	PaymentNoteID    int64
	FileName         string
	OriginalFilename string
	MimeType         *string
	FileSize         int64
	ObjectKey        string
	UploadedBy       int64
	UploadedByName   *string
	CreatedAt        time.Time
}

// PaymentNoteFilters represents filter criteria for listing payment notes
type PaymentNoteFilters struct {
	Status      *string
	IsDraft     *bool
	Search      *string
	Page        int32
	PerPage     int32
}
