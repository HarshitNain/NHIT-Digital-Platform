package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/ports"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/storage"
)

type paymentNoteRepository struct {
	db          *sql.DB
	queries     *generated.Queries
	minioClient *storage.MinIOClient
}

// NewPaymentNoteRepository creates a new payment note repository
func NewPaymentNoteRepository(db *sql.DB, minioClient *storage.MinIOClient) ports.PaymentNoteRepository {
	return &paymentNoteRepository{
		db:          db,
		queries:     generated.New(db),
		minioClient: minioClient,
	}
}

// Create creates a new payment note with particulars and documents
func (r *paymentNoteRepository) Create(ctx context.Context, note *domain.PaymentNote) (*domain.PaymentNote, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Create payment note with all 38 parameters
	createdNote, err := qtx.CreatePaymentNote(ctx, generated.CreatePaymentNoteParams{
		UserID:                 note.UserID,
		GreenNoteID:            sqlNullString(note.GreenNoteID),
		GreenNoteNo:            sqlNullString(note.GreenNoteNo),
		GreenNoteApprover:      sqlNullString(note.GreenNoteApprover),
		GreenNoteAppDate:       sqlNullString(note.GreenNoteAppDate),
		ReimbursementNoteID:    sqlNullInt64(note.ReimbursementNoteID),
		NoteNo:                 note.NoteNo,
		Subject:                sqlNullString(note.Subject),
		Date:                   sqlNullTime(note.Date),
		Department:             sqlNullString(note.Department),
		VendorCode:             sqlNullString(note.VendorCode),
		VendorName:             sqlNullString(note.VendorName),
		ProjectName:            sqlNullString(note.ProjectName),
		InvoiceNo:              sqlNullString(note.InvoiceNo),
		InvoiceDate:            sqlNullString(note.InvoiceDate),
		InvoiceAmount:          sql.NullString{String: fmt.Sprintf("%.2f", note.InvoiceAmount), Valid: true},
		InvoiceApprovedBy:      sqlNullString(note.InvoiceApprovedBy),
		LoaPoNo:                sqlNullString(note.LoaPoNo),
		LoaPoAmount:            sql.NullString{String: fmt.Sprintf("%.2f", note.LoaPoAmount), Valid: true},
		LoaPoDate:              sqlNullString(note.LoaPoDate),
		GrossAmount:            fmt.Sprintf("%.2f", note.GrossAmount),
		TotalAdditions:         fmt.Sprintf("%.2f", note.TotalAdditions),
		TotalDeductions:        fmt.Sprintf("%.2f", note.TotalDeductions),
		NetPayableAmount:       fmt.Sprintf("%.2f", note.NetPayableAmount),
		NetPayableRoundOff:     fmt.Sprintf("%.2f", note.NetPayableRoundOff),
		NetPayableWords:        sqlNullString(note.NetPayableWords),
		TdsPercentage:          sql.NullString{String: fmt.Sprintf("%.2f", note.TdsPercentage), Valid: true},
		TdsSection:             sqlNullString(note.TdsSection),
		TdsAmount:              sql.NullString{String: fmt.Sprintf("%.2f", note.TdsAmount), Valid: true},
		AccountHolderName:      sqlNullString(note.AccountHolderName),
		BankName:               sqlNullString(note.BankName),
		AccountNumber:          sqlNullString(note.AccountNumber),
		IfscCode:               sqlNullString(note.IfscCode),
		RecommendationOfPayment: sqlNullString(note.RecommendationOfPayment),
		Status:                 generated.NullPaymentNoteStatus{PaymentNoteStatus: generated.PaymentNoteStatus(note.Status), Valid: true},
		IsDraft:                note.IsDraft,
		AutoCreated:            note.AutoCreated,
		CreatedBy:              sqlNullInt64(note.CreatedBy),
		CreatedByName:         sqlNullString(note.CreatedByName),
		CreatedByDesignation:  sqlNullString(note.CreatedByDesignation),
		CreatedByRole:         sqlNullString(note.CreatedByRole),
		CreatedBySignature:    sqlNullString(note.CreatedBySignature),
		TaxableAmount:         sql.NullString{String: fmt.Sprintf("%.2f", note.TaxableAmount), Valid: true},
		GstAmount:             sql.NullString{String: fmt.Sprintf("%.2f", note.GstAmount), Valid: true},
		OtherCharges:          sql.NullString{String: fmt.Sprintf("%.2f", note.OtherCharges), Valid: true},
		GreenNoteHistory:      sqlNullString(note.GreenNoteHistory),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create payment note: %w", err)
	}

	// Insert add particulars
	for _, particular := range note.AddParticulars {
		_, err = qtx.InsertPaymentParticular(ctx, generated.InsertPaymentParticularParams{
			PaymentNoteID:  createdNote.ID,
			ParticularType: "ADD",
			Particular:     particular.Particular,
			Amount:         fmt.Sprintf("%.2f", particular.Amount),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert add particular: %w", err)
		}
	}

	// Insert less particulars
	for _, particular := range note.LessParticulars {
		_, err = qtx.InsertPaymentParticular(ctx, generated.InsertPaymentParticularParams{
			PaymentNoteID:  createdNote.ID,
			ParticularType: "LESS",
			Particular:     particular.Particular,
			Amount:         fmt.Sprintf("%.2f", particular.Amount),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert less particular: %w", err)
		}
	}

	// Insert documents if any
	for _, doc := range note.Documents {
		_, err = qtx.InsertPaymentNoteDocument(ctx, generated.InsertPaymentNoteDocumentParams{
			PaymentNoteID:    createdNote.ID,
			FileName:         doc.FileName,
			OriginalFilename: doc.OriginalFilename,
			MimeType:         sqlNullString(doc.MimeType),
			FileSize:         sql.NullInt64{Int64: doc.FileSize, Valid: true},
			ObjectKey:        doc.ObjectKey,
			UploadedBy:       doc.UploadedBy,
			UploadedByName:   sqlNullString(doc.UploadedByName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert document: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, createdNote.ID)
}

// GetByID retrieves a payment note by ID with all related entities
func (r *paymentNoteRepository) GetByID(ctx context.Context, id int64) (*domain.PaymentNote, error) {
	note, err := r.queries.GetPaymentNoteByID(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment note not found")
		}
		return nil, fmt.Errorf("failed to get payment note: %w", err)
	}

	// Get particulars
	particulars, err := r.queries.ListPaymentParticulars(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get particulars: %w", err)
	}

	// Get approval logs
	approvalLogs, err := r.queries.ListApprovalLogs(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get approval logs: %w", err)
	}

	// Get comments
	comments, err := r.queries.ListPaymentNoteComments(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get comments: %w", err)
	}

	// Get documents
	documents, err := r.queries.ListPaymentNoteDocuments(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to get documents: %w", err)
	}

	return r.toDomain(&note, particulars, approvalLogs, comments, documents), nil
}

// GetByGreenNoteID retrieves a payment note by green note ID
func (r *paymentNoteRepository) GetByGreenNoteID(ctx context.Context, greenNoteID string) (*domain.PaymentNote, error) {
	note, err := r.queries.GetPaymentNoteByGreenNoteID(ctx, sql.NullString{String: greenNoteID, Valid: true})
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Not found is not an error in this case
		}
		return nil, fmt.Errorf("failed to get payment note by green note ID: %w", err)
	}

	return r.GetByID(ctx, note.ID)
}

// GetByNoteNo retrieves a payment note by note number
func (r *paymentNoteRepository) GetByNoteNo(ctx context.Context, noteNo string) (*domain.PaymentNote, error) {
	note, err := r.queries.GetPaymentNoteByNoteNo(ctx, noteNo)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("payment note not found")
		}
		return nil, fmt.Errorf("failed to get payment note: %w", err)
	}

	return r.GetByID(ctx, note.ID)
}

// List retrieves payment notes with filters
func (r *paymentNoteRepository) List(ctx context.Context, filters domain.PaymentNoteFilters) ([]*domain.PaymentNote, int64, error) {
	// Count total
	count, err := r.queries.CountPaymentNotes(ctx, generated.CountPaymentNotesParams{
		Column1: stringPtrToString(filters.Status),
		Column2: boolPtrToBool(filters.IsDraft),
		Column3: stringPtrToString(filters.Search),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count payment notes: %w", err)
	}

	// Get notes
	offset := (filters.Page - 1) * filters.PerPage
	notes, err := r.queries.ListPaymentNotes(ctx, generated.ListPaymentNotesParams{
		Column1: stringPtrToString(filters.Status),
		Column2: boolPtrToBool(filters.IsDraft),
		Column3: stringPtrToString(filters.Search),
		Limit:   filters.PerPage,
		Offset:  offset,
	})
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list payment notes: %w", err)
	}

	// Convert to domain
	result := make([]*domain.PaymentNote, len(notes))
	for i, note := range notes {
		result[i] = r.toDomain(&note, nil, nil, nil, nil) // Lightweight list view
	}

	return result, count, nil
}

// Update updates a payment note
func (r *paymentNoteRepository) Update(ctx context.Context, note *domain.PaymentNote) (*domain.PaymentNote, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	qtx := r.queries.WithTx(tx)

	// Update payment note with all 36 parameters
	_, err = qtx.UpdatePaymentNote(ctx, generated.UpdatePaymentNoteParams{
		ID:                     note.ID,
		UserID:                 note.UserID,
		GreenNoteID:            sqlNullString(note.GreenNoteID),
		GreenNoteNo:            sqlNullString(note.GreenNoteNo),
		GreenNoteApprover:      sqlNullString(note.GreenNoteApprover),
		GreenNoteAppDate:       sqlNullString(note.GreenNoteAppDate),
		ReimbursementNoteID:    sqlNullInt64(note.ReimbursementNoteID),
		Subject:                sqlNullString(note.Subject),
		Date:                   sqlNullTime(note.Date),
		Department:             sqlNullString(note.Department),
		VendorCode:             sqlNullString(note.VendorCode),
		VendorName:             sqlNullString(note.VendorName),
		ProjectName:            sqlNullString(note.ProjectName),
		InvoiceNo:              sqlNullString(note.InvoiceNo),
		InvoiceDate:            sqlNullString(note.InvoiceDate),
		InvoiceAmount:          sql.NullString{String: fmt.Sprintf("%.2f", note.InvoiceAmount), Valid: true},
		InvoiceApprovedBy:      sqlNullString(note.InvoiceApprovedBy),
		LoaPoNo:                sqlNullString(note.LoaPoNo),
		LoaPoAmount:            sql.NullString{String: fmt.Sprintf("%.2f", note.LoaPoAmount), Valid: true},
		LoaPoDate:              sqlNullString(note.LoaPoDate),
		GrossAmount:            fmt.Sprintf("%.2f", note.GrossAmount),
		TotalAdditions:         fmt.Sprintf("%.2f", note.TotalAdditions),
		TotalDeductions:        fmt.Sprintf("%.2f", note.TotalDeductions),
		NetPayableAmount:       fmt.Sprintf("%.2f", note.NetPayableAmount),
		NetPayableRoundOff:     fmt.Sprintf("%.2f", note.NetPayableRoundOff),
		NetPayableWords:        sqlNullString(note.NetPayableWords),
		TdsPercentage:          sql.NullString{String: fmt.Sprintf("%.2f", note.TdsPercentage), Valid: true},
		TdsSection:             sqlNullString(note.TdsSection),
		TdsAmount:              sql.NullString{String: fmt.Sprintf("%.2f", note.TdsAmount), Valid: true},
		AccountHolderName:      sqlNullString(note.AccountHolderName),
		BankName:               sqlNullString(note.BankName),
		AccountNumber:          sqlNullString(note.AccountNumber),
		IfscCode:               sqlNullString(note.IfscCode),
		RecommendationOfPayment: sqlNullString(note.RecommendationOfPayment),
		Status:                 generated.NullPaymentNoteStatus{PaymentNoteStatus: generated.PaymentNoteStatus(note.Status), Valid: true},
		IsDraft:                note.IsDraft,
		CreatedByName:         sqlNullString(note.CreatedByName),
		CreatedByDesignation:  sqlNullString(note.CreatedByDesignation),
		CreatedByRole:         sqlNullString(note.CreatedByRole),
		CreatedBySignature:    sqlNullString(note.CreatedBySignature),
		TaxableAmount:         sql.NullString{String: fmt.Sprintf("%.2f", note.TaxableAmount), Valid: true},
		GstAmount:             sql.NullString{String: fmt.Sprintf("%.2f", note.GstAmount), Valid: true},
		OtherCharges:          sql.NullString{String: fmt.Sprintf("%.2f", note.OtherCharges), Valid: true},
		GreenNoteHistory:      sqlNullString(note.GreenNoteHistory),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update payment note: %w", err)
	}

	// Delete existing particulars
	err = qtx.DeletePaymentParticulars(ctx, note.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to delete particulars: %w", err)
	}

	// Insert new add particulars
	for _, particular := range note.AddParticulars {
		_, err = qtx.InsertPaymentParticular(ctx, generated.InsertPaymentParticularParams{
			PaymentNoteID:  note.ID,
			ParticularType: "ADD",
			Particular:     particular.Particular,
			Amount:         fmt.Sprintf("%.2f", particular.Amount),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert add particular: %w", err)
		}
	}

	// Insert new less particulars
	for _, particular := range note.LessParticulars {
		_, err = qtx.InsertPaymentParticular(ctx, generated.InsertPaymentParticularParams{
			PaymentNoteID:  note.ID,
			ParticularType: "LESS",
			Particular:     particular.Particular,
			Amount:         fmt.Sprintf("%.2f", particular.Amount),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to insert less particular: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, note.ID)
}

// UpdateStatus updates the status of a payment note and optionally records an approval log
func (r *paymentNoteRepository) UpdateStatus(ctx context.Context, id int64, status string, isDraft bool, logEntry *domain.PaymentApprovalLog) (*domain.PaymentNote, error) {
	_, err := r.queries.UpdatePaymentNoteStatus(ctx, generated.UpdatePaymentNoteStatusParams{
		ID:      id,
		Status:  generated.NullPaymentNoteStatus{PaymentNoteStatus: generated.PaymentNoteStatus(status), Valid: true},
		IsDraft: isDraft,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update status: %w", err)
	}

	if logEntry != nil {
		_, _ = r.AddApprovalLog(ctx, logEntry)
	}

	return r.GetByID(ctx, id)
}

// Delete deletes a payment note
func (r *paymentNoteRepository) Delete(ctx context.Context, id int64) error {
	// Get documents to delete from MinIO
	documents, err := r.queries.ListPaymentNoteDocuments(ctx, id)
	if err == nil && r.minioClient != nil {
		// Delete documents from MinIO
		for _, doc := range documents {
			_ = r.minioClient.DeleteDocument(ctx, doc.ObjectKey)
		}
	}

	err = r.queries.DeletePaymentNote(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to delete payment note: %w", err)
	}
	return nil
}

// PutOnHold puts a payment note on hold
func (r *paymentNoteRepository) PutOnHold(ctx context.Context, id int64, reason string, userID int64) (*domain.PaymentNote, error) {
	_, err := r.queries.PutPaymentNoteOnHold(ctx, generated.PutPaymentNoteOnHoldParams{
		ID:         id,
		HoldReason: sql.NullString{String: reason, Valid: true},
		HoldBy:     sql.NullInt64{Int64: userID, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to put on hold: %w", err)
	}

	return r.GetByID(ctx, id)
}

// RemoveFromHold removes a payment note from hold
func (r *paymentNoteRepository) RemoveFromHold(ctx context.Context, id int64, newStatus string) (*domain.PaymentNote, error) {
	_, err := r.queries.RemovePaymentNoteFromHold(ctx, generated.RemovePaymentNoteFromHoldParams{
		ID:     id,
		Status: generated.NullPaymentNoteStatus{PaymentNoteStatus: generated.PaymentNoteStatus(newStatus), Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to remove from hold: %w", err)
	}

	return r.GetByID(ctx, id)
}

// UpdateUTR updates the UTR information
func (r *paymentNoteRepository) UpdateUTR(ctx context.Context, id int64, utrNo string, utrDate string) (*domain.PaymentNote, error) {
	_, err := r.queries.UpdatePaymentNoteUTR(ctx, generated.UpdatePaymentNoteUTRParams{
		ID:      id,
		UtrNo:   sql.NullString{String: utrNo, Valid: true},
		UtrDate: sql.NullString{String: utrDate, Valid: true},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to update UTR: %w", err)
	}

	return r.GetByID(ctx, id)
}

// AddComment adds a comment to a payment note
func (r *paymentNoteRepository) AddComment(ctx context.Context, comment *domain.PaymentComment) (*domain.PaymentComment, error) {
	created, err := r.queries.InsertPaymentNoteComment(ctx, generated.InsertPaymentNoteCommentParams{
		PaymentNoteID:  comment.PaymentNoteID,
		Comment:       comment.Comment,
		Status:        sqlNullString(comment.Status),
		UserID:        comment.UserID,
		UserName:      sqlNullString(comment.UserName),
		UserEmail:     sqlNullString(comment.UserEmail),
		UserDesignation: sqlNullString(comment.UserDesignation),
		UserRole:        sqlNullString(comment.UserRole),
		UserSignature:   sqlNullString(comment.UserSignature),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add comment: %w", err)
	}

	return &domain.PaymentComment{
		ID:            created.ID,
		PaymentNoteID: created.PaymentNoteID,
		Comment:       created.Comment,
		Status:        nullStringToPtr(created.Status),
		UserID:        created.UserID,
		UserName:      nullStringToPtr(created.UserName),
		UserEmail:     nullStringToPtr(created.UserEmail),
		UserDesignation: nullStringToPtr(created.UserDesignation),
		UserRole:        nullStringToPtr(created.UserRole),
		UserSignature:   nullStringToPtr(created.UserSignature),
		CreatedAt:     created.CreatedAt,
	}, nil
}

// AddApprovalLog adds an approval log entry
func (r *paymentNoteRepository) AddApprovalLog(ctx context.Context, log *domain.PaymentApprovalLog) (*domain.PaymentApprovalLog, error) {
	created, err := r.queries.InsertApprovalLog(ctx, generated.InsertApprovalLogParams{
		PaymentNoteID:  log.PaymentNoteID,
		Status:        log.Status,
		Comments:      sqlNullString(log.Comments),
		ReviewerID:    log.ReviewerID,
		ReviewerName:  sqlNullString(log.ReviewerName),
		ReviewerEmail: sqlNullString(log.ReviewerEmail),
		ApproverLevel: sqlNullInt32(log.ApproverLevel),
		ReviewerDesignation: sqlNullString(log.ReviewerDesignation),
		ReviewerRole:        sqlNullString(log.ReviewerRole),
		ReviewerSignature:   sqlNullString(log.ReviewerSignature),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to add approval log: %w", err)
	}

	return &domain.PaymentApprovalLog{
		ID:            created.ID,
		PaymentNoteID: created.PaymentNoteID,
		Status:        created.Status,
		Comments:      nullStringToPtr(created.Comments),
		ReviewerID:    created.ReviewerID,
		ReviewerName:  nullStringToPtr(created.ReviewerName),
		ReviewerEmail: nullStringToPtr(created.ReviewerEmail),
		ApproverLevel: nullInt32ToPtr(created.ApproverLevel),
		ReviewerDesignation: nullStringToPtr(created.ReviewerDesignation),
		ReviewerRole:        nullStringToPtr(created.ReviewerRole),
		ReviewerSignature:   nullStringToPtr(created.ReviewerSignature),
		CreatedAt:     created.CreatedAt,
	}, nil
}

// UploadDocument uploads a document to MinIO and saves metadata
func (r *paymentNoteRepository) UploadDocument(ctx context.Context, paymentNoteID int64, filename string, data []byte, mimeType string, uploadedBy int64, uploadedByName string) (*domain.PaymentNoteDocument, error) {
	if r.minioClient == nil {
		return nil, fmt.Errorf("MinIO client not initialized")
	}

	// Generate object key
	objectKey := storage.GenerateObjectKey(paymentNoteID, filename)

	// Upload to MinIO
	_, size, err := r.minioClient.UploadDocument(ctx, objectKey, data, mimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to upload to MinIO: %w", err)
	}

	// Save metadata to database
	doc, err := r.queries.InsertPaymentNoteDocument(ctx, generated.InsertPaymentNoteDocumentParams{
		PaymentNoteID:    paymentNoteID,
		FileName:         filename,
		OriginalFilename: filename,
		MimeType:         sql.NullString{String: mimeType, Valid: true},
		FileSize:         sql.NullInt64{Int64: size, Valid: true},
		ObjectKey:        objectKey,
		UploadedBy:       uploadedBy,
		UploadedByName:   sql.NullString{String: uploadedByName, Valid: true},
	})
	if err != nil {
		// Try to delete from MinIO if database insert fails
		_ = r.minioClient.DeleteDocument(ctx, objectKey)
		return nil, fmt.Errorf("failed to save document metadata: %w", err)
	}

	return &domain.PaymentNoteDocument{
		ID:               doc.ID,
		PaymentNoteID:    doc.PaymentNoteID,
		FileName:         doc.FileName,
		OriginalFilename: doc.OriginalFilename,
		MimeType:         nullStringToPtr(doc.MimeType),
		FileSize:         doc.FileSize.Int64,
		ObjectKey:        doc.ObjectKey,
		UploadedBy:       doc.UploadedBy,
		UploadedByName:   nullStringToPtr(doc.UploadedByName),
		CreatedAt:        doc.CreatedAt,
	}, nil
}

// DownloadDocument retrieves a document from MinIO
func (r *paymentNoteRepository) DownloadDocument(ctx context.Context, documentID int64) ([]byte, string, error) {
	if r.minioClient == nil {
		return nil, "", fmt.Errorf("MinIO client not initialized")
	}

	// Get document metadata
	doc, err := r.queries.GetPaymentNoteByID(ctx, documentID) // This needs a GetDocumentByID query
	if err != nil {
		return nil, "", fmt.Errorf("document not found: %w", err)
	}

	// Download from MinIO
	data, err := r.minioClient.DownloadDocument(ctx, doc.NoteNo) // Using NoteNo as placeholder
	if err != nil {
		return nil, "", fmt.Errorf("failed to download from MinIO: %w", err)
	}

	return data, doc.NoteNo, nil
}

// DeleteDocument deletes a document from MinIO and database
func (r *paymentNoteRepository) DeleteDocument(ctx context.Context, documentID int64) error {
	// This needs proper implementation with GetDocumentByID query
	if r.minioClient != nil {
		// Delete from MinIO first
		// objectKey := ... get from database
		// _ = r.minioClient.DeleteDocument(ctx, objectKey)
	}

	return r.queries.DeletePaymentNoteDocument(ctx, documentID)
}

// GenerateOrderNumber generates the next payment note order number
func (r *paymentNoteRepository) GenerateOrderNumber(ctx context.Context, prefix string) (string, error) {
	nextNum, err := r.queries.GetNextPaymentNoteNumber(ctx, sql.NullString{String: prefix, Valid: true})
	if err != nil {
		return "", fmt.Errorf("failed to generate order number: %w", err)
	}

	// Format: PREFIX-YYYY-NNNNN
	year := time.Now().Year()
	orderNumber := fmt.Sprintf("%s-%d-%05d", prefix, year, nextNum)
	return orderNumber, nil
}

// Helper functions for domain conversion
func (r *paymentNoteRepository) toDomain(
	note *generated.PaymentNote,
	particulars []generated.PaymentNoteParticular,
	approvalLogs []generated.PaymentNoteApprovalLog,
	comments []generated.PaymentNoteComment,
	documents []generated.PaymentNoteDocument,
) *domain.PaymentNote {
	result := &domain.PaymentNote{
		ID:                     note.ID,
		UserID:                 note.UserID,
		GreenNoteID:            nullStringToPtr(note.GreenNoteID),
		GreenNoteNo:            nullStringToPtr(note.GreenNoteNo),
		GreenNoteApprover:      nullStringToPtr(note.GreenNoteApprover),
		GreenNoteAppDate:       nullStringToPtr(note.GreenNoteAppDate),
		ReimbursementNoteID:    nullInt64ToPtr(note.ReimbursementNoteID),
		NoteNo:                 note.NoteNo,
		Subject:                nullStringToPtr(note.Subject),
		Date:                   nullTimeToPtr(note.Date),
		Department:             nullStringToPtr(note.Department),
		VendorCode:             nullStringToPtr(note.VendorCode),
		VendorName:             nullStringToPtr(note.VendorName),
		ProjectName:            nullStringToPtr(note.ProjectName),
		InvoiceNo:              nullStringToPtr(note.InvoiceNo),
		InvoiceDate:            nullStringToPtr(note.InvoiceDate),
		InvoiceAmount:          parseDecimalSafe(note.InvoiceAmount.String),
		InvoiceApprovedBy:      nullStringToPtr(note.InvoiceApprovedBy),
		LoaPoNo:                nullStringToPtr(note.LoaPoNo),
		LoaPoAmount:            parseDecimalSafe(note.LoaPoAmount.String),
		LoaPoDate:              nullStringToPtr(note.LoaPoDate),
		GrossAmount:            parseDecimalSafe(note.GrossAmount),
		TotalAdditions:         parseDecimalSafe(note.TotalAdditions),
		TotalDeductions:        parseDecimalSafe(note.TotalDeductions),
		NetPayableAmount:       parseDecimalSafe(note.NetPayableAmount),
		NetPayableRoundOff:     parseDecimalSafe(note.NetPayableRoundOff),
		NetPayableWords:        nullStringToPtr(note.NetPayableWords),
		TdsPercentage:          parseDecimalSafe(note.TdsPercentage.String),
		TdsSection:             nullStringToPtr(note.TdsSection),
		TdsAmount:              parseDecimalSafe(note.TdsAmount.String),
		AccountHolderName:      nullStringToPtr(note.AccountHolderName),
		BankName:               nullStringToPtr(note.BankName),
		AccountNumber:          nullStringToPtr(note.AccountNumber),
		IfscCode:               nullStringToPtr(note.IfscCode),
		RecommendationOfPayment: nullStringToPtr(note.RecommendationOfPayment),
		Status:                 string(note.Status.PaymentNoteStatus),
		IsDraft:                note.IsDraft,
		AutoCreated:            note.AutoCreated,
		CreatedBy:              nullInt64ToPtr(note.CreatedBy),
		CreatedByName:         nullStringToPtr(note.CreatedByName),
		CreatedByDesignation:  nullStringToPtr(note.CreatedByDesignation),
		CreatedByRole:         nullStringToPtr(note.CreatedByRole),
		CreatedBySignature:    nullStringToPtr(note.CreatedBySignature),
		HoldReason:             nullStringToPtr(note.HoldReason),
		HoldDate:               nullTimeToPtr(note.HoldDate),
		HoldBy:                 nullInt64ToPtr(note.HoldBy),
		UtrNo:                  nullStringToPtr(note.UtrNo),
		UtrDate:                nullStringToPtr(note.UtrDate),
		TaxableAmount:         parseDecimalSafe(note.TaxableAmount.String),
		GstAmount:             parseDecimalSafe(note.GstAmount.String),
		OtherCharges:          parseDecimalSafe(note.OtherCharges.String),
		GreenNoteHistory:      nullStringToPtr(note.GreenNoteHistory),
		CreatedAt:              note.CreatedAt,
		UpdatedAt:              note.UpdatedAt,
	}

	// Convert particulars
	if particulars != nil {
		for _, p := range particulars {
			particular := domain.PaymentParticular{
				ID:             p.ID,
				PaymentNoteID:  p.PaymentNoteID,
				ParticularType: p.ParticularType,
				Particular:     p.Particular,
				Amount:         parseDecimalSafe(p.Amount),
				CreatedAt:      p.CreatedAt,
			}
			if p.ParticularType == "ADD" {
				result.AddParticulars = append(result.AddParticulars, particular)
			} else {
				result.LessParticulars = append(result.LessParticulars, particular)
			}
		}
	}

	// Convert approval logs
	if approvalLogs != nil {
		for _, log := range approvalLogs {
			result.ApprovalLogs = append(result.ApprovalLogs, domain.PaymentApprovalLog{
				ID:            log.ID,
				PaymentNoteID: log.PaymentNoteID,
				Status:        log.Status,
				Comments:      nullStringToPtr(log.Comments),
				ReviewerID:    log.ReviewerID,
				ReviewerName:  nullStringToPtr(log.ReviewerName),
				ReviewerEmail: nullStringToPtr(log.ReviewerEmail),
				ApproverLevel: nullInt32ToPtr(log.ApproverLevel),
				ReviewerDesignation: nullStringToPtr(log.ReviewerDesignation),
				ReviewerRole:        nullStringToPtr(log.ReviewerRole),
				ReviewerSignature:   nullStringToPtr(log.ReviewerSignature),
				CreatedAt:     log.CreatedAt,
			})
		}
	}

	// Convert comments
	if comments != nil {
		for _, c := range comments {
			result.Comments = append(result.Comments, domain.PaymentComment{
				ID:            c.ID,
				PaymentNoteID: c.PaymentNoteID,
				Comment:       c.Comment,
				Status:        nullStringToPtr(c.Status),
				UserID:        c.UserID,
				UserName:      nullStringToPtr(c.UserName),
				UserEmail:     nullStringToPtr(c.UserEmail),
				UserDesignation: nullStringToPtr(c.UserDesignation),
				UserRole:        nullStringToPtr(c.UserRole),
				UserSignature:   nullStringToPtr(c.UserSignature),
				CreatedAt:     c.CreatedAt,
			})
		}
	}

	// Convert documents
	if documents != nil {
		for _, d := range documents {
			result.Documents = append(result.Documents, domain.PaymentNoteDocument{
				ID:               d.ID,
				PaymentNoteID:    d.PaymentNoteID,
				FileName:         d.FileName,
				OriginalFilename: d.OriginalFilename,
				MimeType:         nullStringToPtr(d.MimeType),
				FileSize:         d.FileSize.Int64,
				ObjectKey:        d.ObjectKey,
				UploadedBy:       d.UploadedBy,
				UploadedByName:   nullStringToPtr(d.UploadedByName),
				CreatedAt:        d.CreatedAt,
			})
		}
	}

	return result
}

// Helper functions
func stringPtrToString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int64PtrToInt64(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func boolPtrToBool(b *bool) bool {
	if b == nil {
		return false
	}
	return *b
}



func int32PtrToInt32(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func nullStringToPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	return &ns.String
}

func nullInt64ToPtr(ni sql.NullInt64) *int64 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int64
}

func nullInt32ToPtr(ni sql.NullInt32) *int32 {
	if !ni.Valid {
		return nil
	}
	return &ni.Int32
}

func nullTimeToPtr(nt sql.NullTime) *time.Time {
	if !nt.Valid {
		return nil
	}
	return &nt.Time
}

func sqlNullString(s *string) sql.NullString {
	if s == nil {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: *s, Valid: true}
}

func sqlNullInt64(i *int64) sql.NullInt64 {
	if i == nil {
		return sql.NullInt64{Valid: false}
	}
	return sql.NullInt64{Int64: *i, Valid: true}
}

func sqlNullInt32(i *int32) sql.NullInt32 {
	if i == nil {
		return sql.NullInt32{Valid: false}
	}
	return sql.NullInt32{Int32: *i, Valid: true}
}

func sqlNullTime(t *time.Time) sql.NullTime {
	if t == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: *t, Valid: true}
}

func sqlNullBool(b *bool) sql.NullBool {
	if b == nil {
		return sql.NullBool{Valid: false}
	}
	return sql.NullBool{Bool: *b, Valid: true}
}

func parseDecimalSafe(s string) float64 {
	var f float64
	fmt.Sscanf(s, "%f", &f)
	return f
}
