package ports

import (
	"context"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/domain"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
)

// PaymentNoteRepository defines the interface for payment note data operations
type PaymentNoteRepository interface {
	// Create creates a new payment note with particulars
	Create(ctx context.Context, note *domain.PaymentNote) (*domain.PaymentNote, error)
	
	// GetByID retrieves a payment note by ID with all related entities
	GetByID(ctx context.Context, id int64) (*domain.PaymentNote, error)
	
	// GetByGreenNoteID retrieves a payment note by green note ID
	GetByGreenNoteID(ctx context.Context, greenNoteID string) (*domain.PaymentNote, error)
	
	// GetByNoteNo retrieves a payment note by note number
	GetByNoteNo(ctx context.Context, noteNo string) (*domain.PaymentNote, error)
	
	// List retrieves payment notes with filters
	List(ctx context.Context, filters domain.PaymentNoteFilters) ([]*domain.PaymentNote, int64, error)
	
	// Update updates a payment note
	Update(ctx context.Context, note *domain.PaymentNote) (*domain.PaymentNote, error)
	
	// UpdateStatus updates only the status of a payment note
	UpdateStatus(ctx context.Context, id int64, status string, isDraft bool, logEntry *domain.PaymentApprovalLog) (*domain.PaymentNote, error)
	
	// Delete deletes a payment note
	Delete(ctx context.Context, id int64) error
	
	// PutOnHold puts a payment note on hold
	PutOnHold(ctx context.Context, id int64, reason string, userID int64) (*domain.PaymentNote, error)
	
	// RemoveFromHold removes a payment note from hold
	RemoveFromHold(ctx context.Context, id int64, newStatus string) (*domain.PaymentNote, error)
	
	// UpdateUTR updates the UTR information
	UpdateUTR(ctx context.Context, id int64, utrNo string, utrDate string) (*domain.PaymentNote, error)
	
	// AddComment adds a comment to a payment note
	AddComment(ctx context.Context, comment *domain.PaymentComment) (*domain.PaymentComment, error)
	
	// AddApprovalLog adds an approval log entry
	AddApprovalLog(ctx context.Context, log *domain.PaymentApprovalLog) (*domain.PaymentApprovalLog, error)
	
	// GenerateOrderNumber generates the next payment note order number
	GenerateOrderNumber(ctx context.Context, prefix string) (string, error)
	
	// Document management
	// UploadDocument uploads a document to MinIO and saves metadata
	UploadDocument(ctx context.Context, paymentNoteID int64, filename string, data []byte, mimeType string, uploadedBy int64, uploadedByName string) (*domain.PaymentNoteDocument, error)
	
	// DownloadDocument retrieves a document from MinIO
	DownloadDocument(ctx context.Context, documentID int64) ([]byte, string, error)
	
	// DeleteDocument deletes a document from MinIO and database
	DeleteDocument(ctx context.Context, documentID int64) error
}

// PaymentNoteService defines the business logic interface
type PaymentNoteService interface {
	CreatePaymentNote(ctx context.Context, payload *domain.PaymentNote) (*domain.PaymentNote, error)
	GetPaymentNoteByID(ctx context.Context, id int64) (*domain.PaymentNote, error)
	GetPaymentNoteByNoteNo(ctx context.Context, noteNo string) (*domain.PaymentNote, error)
	ListPaymentNotes(ctx context.Context, filters domain.PaymentNoteFilters) ([]*domain.PaymentNote, int64, error)
	UpdatePaymentNote(ctx context.Context, id int64, note *domain.PaymentNote) (*domain.PaymentNote, error)
	DeletePaymentNote(ctx context.Context, id int64) error
	CreateDraftFromGreenNote(ctx context.Context, greenNoteID string, greenNoteData map[string]interface{}) (*domain.PaymentNote, error)
	ConvertDraftToActive(ctx context.Context, id int64) (*domain.PaymentNote, error)
	PutOnHold(ctx context.Context, id int64, reason string, userID int64) (*domain.PaymentNote, error)
	RemoveFromHold(ctx context.Context, id int64, newStatus string) (*domain.PaymentNote, error)
	UpdateUTR(ctx context.Context, id int64, utrNo string, utrDate string) (*domain.PaymentNote, error)
	AddComment(ctx context.Context, paymentNoteID int64, comment string, status string, userID int64, userName string, userEmail string) (*domain.PaymentComment, error)
	GeneratePaymentNoteOrderNumber(ctx context.Context) (string, error)
	UpdateStatus(ctx context.Context, id int64, status string, isDraft bool, logEntry *domain.PaymentApprovalLog) (*domain.PaymentNote, error)
	DomainToProto(note *domain.PaymentNote) *paymentnotepb.PaymentNote
	DomainToSummaryProto(note *domain.PaymentNote) *paymentnotepb.PaymentNoteSummary
	ProtoToDomain(proto *paymentnotepb.PaymentNotePayload) *domain.PaymentNote
}
