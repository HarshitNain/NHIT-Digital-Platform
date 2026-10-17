package services

import (
	"context"
	"fmt"
	"time"
	"log"
	"strings"

	
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/domain"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/core/ports"
	"github.com/ShristiRnr/Nhit-Note/services/paymentnote-service/internal/utils"
	paymentnotepb "github.com/ShristiRnr/Nhit-Note/api/pb/paymentnotepb"
	approvalpb "github.com/ShristiRnr/Nhit-Note/api/pb/approvalpb"
	potrackingpb "github.com/ShristiRnr/Nhit-Note/api/pb/potrackingpb"
	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	commonpb "github.com/ShristiRnr/Nhit-Note/api/pb/commonpb"
)

type paymentNoteService struct {
	repo             ports.PaymentNoteRepository
	approvalClient   approvalpb.ApprovalServiceClient
	potrackingClient potrackingpb.POTrackingServiceClient
	greenNoteClient  greennotepb.GreenNoteServiceClient
}

// NewPaymentNoteService creates a new payment note service
func NewPaymentNoteService(repo ports.PaymentNoteRepository, approvalClient approvalpb.ApprovalServiceClient, potrackingClient potrackingpb.POTrackingServiceClient, greenNoteClient greennotepb.GreenNoteServiceClient) ports.PaymentNoteService {
	return &paymentNoteService{
		repo:             repo,
		approvalClient:   approvalClient,
		potrackingClient: potrackingClient,
		greenNoteClient:  greenNoteClient,
	}
}

// CreatePaymentNote creates a new payment note with financial calculations
func (s *paymentNoteService) CreatePaymentNote(ctx context.Context, payload *domain.PaymentNote) (*domain.PaymentNote, error) {
	note := payload
	
	// Ensure at least one reference is provided for direct creation
	if (note.GreenNoteID == nil || *note.GreenNoteID == "") && (note.ReimbursementNoteID == nil || *note.ReimbursementNoteID == 0) {
		return nil, fmt.Errorf("must provide either Green Note ID or Reimbursement Note ID")
	}

	// Auto-fetch details if Green Note ID is provided
	if note.GreenNoteID != nil && *note.GreenNoteID != "" && s.greenNoteClient != nil {
		log.Printf("🔍 Auto-fetching GreenNote info for creation, ID: %s", *note.GreenNoteID)
		gnResp, err := s.greenNoteClient.GetGreenNote(ctx, &greennotepb.GetGreenNoteRequest{
			Id: *note.GreenNoteID,
		})
		if err == nil && gnResp.Data != nil {
			gn := gnResp.Data
			
			// Auto-populate fields
			if note.GreenNoteNo == nil || *note.GreenNoteNo == "" { note.GreenNoteNo = &gn.WorkOrderNo }
			if note.GreenNoteAppDate == nil || *note.GreenNoteAppDate == "" { 
				// GreenNote doesn't have a specific approved_at field, use updated_at as proxy
				if gn.UpdatedAt != nil {
					appDate := gn.UpdatedAt.AsTime().Format("2006-01-02")
					note.GreenNoteAppDate = &appDate
				}
			}
			if note.Department == nil || *note.Department == "" { note.Department = &gn.DepartmentName }
			if note.ProjectName == nil || *note.ProjectName == "" { note.ProjectName = &gn.ProjectName }
			
			// Extract approver if not present
			if note.GreenNoteApprover == nil || *note.GreenNoteApprover == "" {
				approver := extractApprover(gn.Remarks)
				if approver != "" {
					note.GreenNoteApprover = &approver
				}
			}
			
			// Financials
			if note.TaxableAmount == 0 { note.TaxableAmount = gn.BaseValue }
			if note.GstAmount == 0 { note.GstAmount = gn.Gst }
			if note.OtherCharges == 0 { note.OtherCharges = gn.OtherCharges }
			
			// Invoices
			if gn.Invoice != nil {
				if note.InvoiceNo == nil || *note.InvoiceNo == "" { note.InvoiceNo = &gn.Invoice.InvoiceNumber }
				if note.InvoiceDate == nil || *note.InvoiceDate == "" { note.InvoiceDate = &gn.Invoice.InvoiceDate }
				if note.InvoiceAmount == 0 { note.InvoiceAmount = gn.Invoice.InvoiceValue }
			}
			
			// Gross Amount (starting point for financials)
			if note.GrossAmount == 0 {
				note.GrossAmount = gn.TotalAmount
			}
		}
	}

	// Set status to Pending (P) if not explicitly set as draft
	if !note.IsDraft && (note.Status == "" || note.Status == "D") {
		note.Status = "P"
		note.IsDraft = false
	}
	
	// Calculate financial fields if not provided
	note = s.calculateFinancials(note)
	
	// Generate note number if not provided
	if note.NoteNo == "" {
		noteNo, err := s.repo.GenerateOrderNumber(ctx, "PN")
		if err != nil {
			return nil, fmt.Errorf("failed to generate note number: %w", err)
		}
		note.NoteNo = noteNo
	}
	
	// Create in database
	created, err := s.repo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment note: %w", err)
	}

	// If it's created as pending, initiate approval workflow
	if created.Status == "P" && !created.IsDraft {
		log.Printf("🚀 Initiating approval for directly created PaymentNote %d", created.ID)
		resp, err := s.ConvertDraftToActive(ctx, created.ID)
		if err == nil {
			created = resp
		} else {
			log.Printf("⚠️ Failed to initiate approval for directly created PN %d: %v", created.ID, err)
		}
	}
	
	return created, nil
}

// UpdatePaymentNote updates an existing payment note
func (s *paymentNoteService) UpdatePaymentNote(ctx context.Context, id int64, payload *domain.PaymentNote) (*domain.PaymentNote, error) {
	updated := payload
	existing, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("payment note not found: %w", err)
	}
	
	updated.ID = existing.ID
	updated.NoteNo = existing.NoteNo // Keep existing note number
	updated.CreatedAt = existing.CreatedAt
	
	// Recalculate financials
	updated = s.calculateFinancials(updated)
	
	// Update in database
	result, err := s.repo.Update(ctx, updated)
	if err != nil {
		return nil, fmt.Errorf("failed to update payment note: %w", err)
	}
	
	return result, nil
}

// GetPaymentNoteByID retrieves a payment note by ID and enriches it with PO details
func (s *paymentNoteService) GetPaymentNoteByID(ctx context.Context, id int64) (*domain.PaymentNote, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// If the note is in draft status, change it to pending as per user request (view transitions to pending)
	if note.Status == "D" || note.IsDraft {
		log.Printf("🔄 Transitioning PaymentNote %d from Draft to Pending on view", id)
		// Update status to Pending (P)
		updatedNote, err := s.repo.UpdateStatus(ctx, id, "P", false, nil)
		if err == nil {
			note = updatedNote
			// Initiate approval workflow if it's the first time it transitions
			// We can reuse ConvertDraftToActive logic or call InitiateApproval directly
			_, _ = s.ConvertDraftToActive(ctx, id)
			log.Printf("✅ Success: PaymentNote %d transitioned to Pending and approval initiated", id)
		} else {
			log.Printf("⚠️ Failed to update status on view for PaymentNote %d: %v", id, err)
		}
	}
	
	// Enrichment with GreenNote details
	if note.GreenNoteID != nil && *note.GreenNoteID != "" && s.greenNoteClient != nil {
		log.Printf("🔍 Fetching GreenNote info for ID: %s", *note.GreenNoteID)
		gnResp, err := s.greenNoteClient.GetGreenNote(ctx, &greennotepb.GetGreenNoteRequest{
			Id: *note.GreenNoteID,
		})
		if err == nil && gnResp.Data != nil {
			gn := gnResp.Data
			log.Printf("✅ Found GreenNote: %s", gn.WorkOrderNo)
			
			// Auto-populate fields from GreenNote
			note.GreenNoteNo = &gn.WorkOrderNo
			// GreenNote doesn't have a specific approved_at field, use updated_at as proxy
			if gn.UpdatedAt != nil {
				appDate := gn.UpdatedAt.AsTime().Format("2006-01-02")
				note.GreenNoteAppDate = &appDate
			}
			note.Department = &gn.DepartmentName
			note.ProjectName = &gn.ProjectName
			note.TaxableAmount = gn.BaseValue
			note.GstAmount = gn.Gst
			note.OtherCharges = gn.OtherCharges
			note.GreenNoteHistory = &gn.Remarks // GN approval logs are in remarks
			
			// Extract approver
			approver := extractApprover(gn.Remarks)
			if approver != "" {
				note.GreenNoteApprover = &approver
			}

			// If GN has primary invoice, use it
			if gn.Invoice != nil {
				note.InvoiceNo = &gn.Invoice.InvoiceNumber
				note.InvoiceDate = &gn.Invoice.InvoiceDate
				note.InvoiceAmount = gn.Invoice.InvoiceValue
			}
			
			// Construct view link
			gnUrl := fmt.Sprintf("/green-notes/view/%s", *note.GreenNoteID)
			note.GreenNoteURL = &gnUrl
		} else if err != nil {
			log.Printf("⚠️ Failed to fetch GreenNote details: %v", err)
		}
	}
	
	// Enrichment with PO details from potracking-service (existing logic)
	if note.LoaPoNo != nil && *note.LoaPoNo != "" && s.potrackingClient != nil {
		log.Printf("🔍 Fetching PO Tracking info for OrderNo: %s", *note.LoaPoNo)
		poResp, err := s.potrackingClient.GetPOByOrderNumber(ctx, &potrackingpb.GetPORequest{
			OrderNo: *note.LoaPoNo,
		})
		if err == nil && poResp.GetPoTracking() != nil {
			po := poResp.GetPoTracking()
			log.Printf("✅ Found PO Tracking: Amount=%.2f, Date=%s", po.OrderValue, po.OrderDate)
			note.LoaPoAmount = po.OrderValue
			poDate := po.OrderDate
			note.LoaPoDate = &poDate
		} else if err != nil {
			log.Printf("⚠️ Failed to fetch PO details from potracking-service: %v", err)
		}
	}
	
	return note, nil
}

// GetPaymentNoteByNoteNo retrieves a payment note by note number
func (s *paymentNoteService) GetPaymentNoteByNoteNo(ctx context.Context, noteNo string) (*domain.PaymentNote, error) {
	return s.repo.GetByNoteNo(ctx, noteNo)
}

// ListPaymentNotes retrieves payment notes with filters
func (s *paymentNoteService) ListPaymentNotes(ctx context.Context, filters domain.PaymentNoteFilters) ([]*domain.PaymentNote, int64, error) {
	// Apply defaults
	if filters.Page <= 0 {
		filters.Page = 1
	}
	if filters.PerPage <= 0 {
		filters.PerPage = 10
	}
	return s.repo.List(ctx, filters)
}

// UpdateStatus updates the status of a payment note and records an approval log
func (s *paymentNoteService) UpdateStatus(ctx context.Context, id int64, status string, isDraft bool, logEntry *domain.PaymentApprovalLog) (*domain.PaymentNote, error) {
	return s.repo.UpdateStatus(ctx, id, status, isDraft, logEntry)
}

// DeletePaymentNote deletes a payment note
func (s *paymentNoteService) DeletePaymentNote(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}

// CreateDraftFromGreenNote creates a draft payment note from an approved green note
func (s *paymentNoteService) CreateDraftFromGreenNote(ctx context.Context, greenNoteID string, greenNoteData map[string]interface{}) (*domain.PaymentNote, error) {
	// Check if draft already exists
	existing, err := s.repo.GetByGreenNoteID(ctx, greenNoteID)
	if err == nil && existing != nil {
		return existing, nil // Already exists, return it
	}
	
	orderNo := ""
	if v, ok := greenNoteData["order_no"].(string); ok {
		orderNo = v
	}

	poNo := ""
	if v, ok := greenNoteData["po_number"].(string); ok {
		poNo = v
	}
	
	netAmount := 0.0
	if v, ok := greenNoteData["net_amount"].(float64); ok {
		netAmount = v
	}

	approver := ""
	if v, ok := greenNoteData["approver"].(string); ok {
		approver = v
	}

	appDate := ""
	if v, ok := greenNoteData["approved_at"].(string); ok {
		appDate = v
	}

	// Create draft payment note
	note := &domain.PaymentNote{
		GreenNoteID:       &greenNoteID,
		GreenNoteNo:       &orderNo,
		LoaPoNo:           &poNo,
		GreenNoteApprover: &approver,
		GreenNoteAppDate:  &appDate,
		GrossAmount:       netAmount,
		Status:            "D", // Draft
		IsDraft:           true,
		AutoCreated:       true,
		CreatedAt:         time.Now(),
		UserID:            1, // Default System/Admin User to satisfy DB constraint
	}
	
	// Calculate initial financials (Net Payable, etc.)
	note = s.calculateFinancials(note)
	
	// Generate note number
	noteNo, err := s.repo.GenerateOrderNumber(ctx, "PN")
	if err != nil {
		return nil, fmt.Errorf("failed to generate note number: %w", err)
	}
	note.NoteNo = noteNo
	
	// Create in database
	created, err := s.repo.Create(ctx, note)
	if err != nil {
		return nil, fmt.Errorf("failed to create draft: %w", err)
	}

	// 1. Add automated log entry for draft creation
	logMsg := fmt.Sprintf("draft payment note created for greennote id : %s", greenNoteID)
	_, _ = s.repo.AddApprovalLog(ctx, &domain.PaymentApprovalLog{
		PaymentNoteID: created.ID,
		Status:        "D",
		Comments:      &logMsg,
		ReviewerID:    0, // System
		ReviewerName:  protoStringPtr("System"),
	})
	log.Printf("📜 %s", logMsg)

	// 2. Link Payment Note ID back to Green Note
	if s.greenNoteClient != nil {
		log.Printf("🔗 Linking PaymentNote %d to GreenNote %s", created.ID, greenNoteID)
		_, err := s.greenNoteClient.LinkPaymentNote(ctx, &greennotepb.LinkPaymentNoteRequest{
			Id:            greenNoteID,
			PaymentNoteId: fmt.Sprintf("%d", created.ID),
		})
		if err != nil {
			log.Printf("⚠️ Failed to link PaymentNote to GreenNote: %v", err)
		} else {
			log.Printf("✅ Successfully linked PaymentNote %d to GreenNote %s", created.ID, greenNoteID)
		}
	}
	
	return created, nil
}

// ConvertDraftToActive converts a draft payment note to active status and initiates approval
func (s *paymentNoteService) ConvertDraftToActive(ctx context.Context, id int64) (*domain.PaymentNote, error) {
	note, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	
	if !note.IsDraft {
		return nil, fmt.Errorf("payment note is not a draft")
	}
	
	// Initiate approval workflow
	// SourceType for PaymentNote should be "PAYMENTNOTE"
	resp, err := s.approvalClient.InitiateApproval(ctx, &approvalpb.InitiateApprovalRequest{
		SlNo:                fmt.Sprintf("%d", note.ID),
		Amount:              note.NetPayableAmount,
		CreatedByUserId:     fmt.Sprintf("%d", note.UserID),
		SourceType:          "PAYMENTNOTE",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to initiate approval: %w", err)
	}

	// Update status based on approval initiation
	return s.repo.UpdateStatus(ctx, id, resp.Status, false, nil)
}

// PutOnHold puts a payment note on hold
func (s *paymentNoteService) PutOnHold(ctx context.Context, id int64, reason string, userID int64) (*domain.PaymentNote, error) {
	return s.repo.PutOnHold(ctx, id, reason, userID)
}

// RemoveFromHold removes a payment note from hold
func (s *paymentNoteService) RemoveFromHold(ctx context.Context, id int64, newStatus string) (*domain.PaymentNote, error) {
	return s.repo.RemoveFromHold(ctx, id, newStatus)
}

// UpdateUTR updates the UTR information
func (s *paymentNoteService) UpdateUTR(ctx context.Context, id int64, utrNo string, utrDate string) (*domain.PaymentNote, error) {
	return s.repo.UpdateUTR(ctx, id, utrNo, utrDate)
}

// AddComment adds a comment to a payment note
func (s *paymentNoteService) AddComment(ctx context.Context, paymentNoteID int64, comment string, status string, userID int64, userName string, userEmail string) (*domain.PaymentComment, error) {
	c := &domain.PaymentComment{
		PaymentNoteID: paymentNoteID,
		Comment:       comment,
		Status:        stringPtr(status),
		UserID:        userID,
		UserName:      stringPtr(userName),
		UserEmail:     stringPtr(userEmail),
	}
	return s.repo.AddComment(ctx, c)
}

// GeneratePaymentNoteOrderNumber generates a payment note order number
func (s *paymentNoteService) GeneratePaymentNoteOrderNumber(ctx context.Context) (string, error) {
	return s.repo.GenerateOrderNumber(ctx, "PN")
}

// UploadDocument uploads a document to a payment note
func (s *paymentNoteService) UploadDocument(ctx context.Context, paymentNoteID int64, filename string, data []byte, mimeType string, uploadedBy int64, uploadedByName string) (*domain.PaymentNoteDocument, error) {
	return s.repo.UploadDocument(ctx, paymentNoteID, filename, data, mimeType, uploadedBy, uploadedByName)
}


// calculateFinancials calculates all financial fields for a payment note
func (s *paymentNoteService) calculateFinancials(note *domain.PaymentNote) *domain.PaymentNote {
	// 1. Calculate Gross Amount from breakdown if not manually set
	// Gross Amount (Invoice) = Taxable + GST + Other
	if note.TaxableAmount > 0 || note.GstAmount > 0 || note.OtherCharges > 0 {
		note.GrossAmount = note.TaxableAmount + note.GstAmount + note.OtherCharges
	}

	// 2. Calculate total custom additions
	totalAdd := 0.0
	if note.AddParticulars != nil {
		for _, p := range note.AddParticulars {
			totalAdd += p.Amount
		}
	}
	note.TotalAdditions = totalAdd
	
	// 3. Calculate total custom deductions (excluding TDS for now)
	totalDeduct := 0.0
	if note.LessParticulars != nil {
		for _, p := range note.LessParticulars {
			totalDeduct += p.Amount
		}
	}
	
	// 4. Calculate TDS if percentage is provided (on Taxable Amount as per standard rules)
	if note.TdsPercentage > 0 {
		// Usually TDS is on Taxable Amount (Base Value)
		targetAmount := note.TaxableAmount
		if targetAmount == 0 {
			targetAmount = note.GrossAmount // Fallback
		}
		note.TdsAmount = utils.CalculateTDS(targetAmount, note.TdsPercentage)
		totalDeduct += note.TdsAmount
	}
	note.TotalDeductions = totalDeduct
	
	// 5. Calculate net payable: Gross (Invoice) + Additions - Deductions
	log.Printf("DEBUG CALC: Taxable=%.2f, GST=%.2f, Other=%.2f -> Gross=%.2f", note.TaxableAmount, note.GstAmount, note.OtherCharges, note.GrossAmount)
	log.Printf("DEBUG CALC: Add=%.2f, Deduct=%.2f", note.TotalAdditions, note.TotalDeductions)
	note.NetPayableAmount = utils.CalculateNetPayable(note.GrossAmount, note.TotalAdditions, note.TotalDeductions)
	log.Printf("DEBUG CALC: Result Net=%.2f", note.NetPayableAmount)
	
	// 6. Round off
	note.NetPayableRoundOff = utils.RoundOff(note.NetPayableAmount)
	
	// 7. Convert to words
	words := utils.NumberToWords(note.NetPayableRoundOff)
	note.NetPayableWords = &words
	
	return note
}

// ProtoToDomain converts proto payload to domain model
func (s *paymentNoteService) ProtoToDomain(proto *paymentnotepb.PaymentNotePayload) *domain.PaymentNote {
	note := &domain.PaymentNote{
		UserID:                 protoToUserId(proto.UserId),
		GreenNoteID:            protoStringPtr(proto.GreenNoteId),
		GreenNoteNo:            protoStringPtr(proto.GreenNoteNo),
		GreenNoteApprover:      protoStringPtr(proto.GreenNoteApprover),
		GreenNoteAppDate:       protoStringPtr(proto.GreenNoteAppDate),
		ReimbursementNoteID:    protoInt64Ptr(proto.ReimbursementNoteId),
		NoteNo:                 proto.NoteNo,
		Subject:                protoStringPtr(proto.Subject),
		Department:             protoStringPtr(proto.Department),
		VendorCode:             protoStringPtr(proto.VendorCode),
		VendorName:             protoStringPtr(proto.VendorName),
		ProjectName:            protoStringPtr(proto.ProjectName),
		InvoiceNo:              protoStringPtr(proto.InvoiceNo),
		InvoiceDate:            protoStringPtr(proto.InvoiceDate),
		InvoiceAmount:          proto.InvoiceAmount,
		InvoiceApprovedBy:      protoStringPtr(proto.InvoiceApprovedBy),
		LoaPoNo:                protoStringPtr(proto.LoaPoNo),
		LoaPoAmount:            proto.LoaPoAmount,
		LoaPoDate:              protoStringPtr(proto.LoaPoDate),
		GrossAmount:            proto.GrossAmount,
		TdsPercentage:          proto.TdsPercentage,
		TdsSection:             protoStringPtr(proto.TdsSection),
		AccountHolderName:      protoStringPtr(proto.AccountHolderName),
		BankName:               protoStringPtr(proto.BankName),
		AccountNumber:          protoStringPtr(proto.AccountNumber),
		IfscCode:               protoStringPtr(proto.IfscCode),
		RecommendationOfPayment: protoStringPtr(proto.RecommendationOfPayment),
		Status:                 proto.Status,
		IsDraft:                proto.IsDraft,
		AutoCreated:            proto.AutoCreated,
		CreatedBy:              protoStringToInt64Ptr(proto.CreatedBy),
		CreatedByName:         protoStringPtr(proto.CreatedByName),
		CreatedByDesignation:  protoStringPtr(proto.CreatedByDesignation),
		CreatedByRole:         protoStringPtr(proto.CreatedByRole),
		TaxableAmount:         proto.TaxableAmount,
		GstAmount:             proto.GstAmount,
		OtherCharges:          proto.OtherCharges,
		GreenNoteHistory:      protoStringPtr(proto.GreenNoteHistory),
	}
	
	// Parse date if provided
	if proto.Date != "" {
		if t, err := time.Parse(time.RFC3339, proto.Date); err == nil {
			note.Date = &t
		}
	}
	
	// Convert add particulars
	for _, p := range proto.AddParticulars {
		note.AddParticulars = append(note.AddParticulars, domain.PaymentParticular{
			ParticularType: "ADD",
			Particular:     p.Particular,
			Amount:         p.Amount,
		})
	}
	
	// Convert less particulars
	for _, p := range proto.LessParticulars {
		note.LessParticulars = append(note.LessParticulars, domain.PaymentParticular{
			ParticularType: "LESS",
			Particular:     p.Particular,
			Amount:         p.Amount,
		})
	}
	
	// Note: Documents are handled separately via UploadDocument
	
	return note
}

// domainToProto converts domain model to proto response
func (s *paymentNoteService) DomainToProto(note *domain.PaymentNote) *paymentnotepb.PaymentNote {
	proto := &paymentnotepb.PaymentNote{
		Id:                     note.ID,
		UserId:                 fmt.Sprintf("%d", note.UserID),
		GreenNoteId:            protoStringVal(note.GreenNoteID),
		GreenNoteNo:            stringPtrToProto(note.GreenNoteNo),
		GreenNoteApprover:      stringPtrToProto(note.GreenNoteApprover),
		GreenNoteAppDate:       stringPtrToProto(note.GreenNoteAppDate),
		ReimbursementNoteId:    int64PtrToProto(note.ReimbursementNoteID),
		NoteNo:                 note.NoteNo,
		Subject:                stringPtrToProto(note.Subject),
		Date:                   timePtrToProto(note.Date),
		Department:             stringPtrToProto(note.Department),
		VendorCode:             stringPtrToProto(note.VendorCode),
		VendorName:             stringPtrToProto(note.VendorName),
		ProjectName:            stringPtrToProto(note.ProjectName),
		InvoiceNo:              stringPtrToProto(note.InvoiceNo),
		InvoiceDate:            stringPtrToProto(note.InvoiceDate),
		InvoiceAmount:          note.InvoiceAmount,
		InvoiceApprovedBy:      stringPtrToProto(note.InvoiceApprovedBy),
		LoaPoNo:                stringPtrToProto(note.LoaPoNo),
		LoaPoAmount:            note.LoaPoAmount,
		LoaPoDate:              stringPtrToProto(note.LoaPoDate),
		GrossAmount:            note.GrossAmount,
		TotalAdditions:         note.TotalAdditions,
		TotalDeductions:        note.TotalDeductions,
		NetPayableAmount:       note.NetPayableAmount,
		NetPayableRoundOff:     note.NetPayableRoundOff,
		NetPayableWords:        stringPtrToProto(note.NetPayableWords),
		TdsPercentage:          note.TdsPercentage,
		TdsSection:             stringPtrToProto(note.TdsSection),
		TdsAmount:              note.TdsAmount,
		AccountHolderName:      stringPtrToProto(note.AccountHolderName),
		BankName:               stringPtrToProto(note.BankName),
		AccountNumber:          stringPtrToProto(note.AccountNumber),
		IfscCode:               stringPtrToProto(note.IfscCode),
		RecommendationOfPayment: stringPtrToProto(note.RecommendationOfPayment),
		Status:                 note.Status,
		IsDraft:                note.IsDraft,
		AutoCreated:            note.AutoCreated,
		CreatedBy:              fmt.Sprintf("%d", int64PtrToProto(note.CreatedBy)),
		CreatedByName:         stringPtrToProto(note.CreatedByName),
		CreatedByDesignation:  stringPtrToProto(note.CreatedByDesignation),
		CreatedByRole:         stringPtrToProto(note.CreatedByRole),
		CreatedAt:              note.CreatedAt.Format(time.RFC3339),
		UpdatedAt:              note.UpdatedAt.Format(time.RFC3339),
		TaxableAmount:         note.TaxableAmount,
		GstAmount:             note.GstAmount,
		OtherCharges:          note.OtherCharges,
		GreenNoteHistory:      stringPtrToProto(note.GreenNoteHistory),
		GreenNoteUrl:          stringPtrToProto(note.GreenNoteURL),
	}
	
	// Convert add particulars
	for _, p := range note.AddParticulars {
		proto.AddParticulars = append(proto.AddParticulars, &paymentnotepb.PaymentParticular{
			Particular: p.Particular,
			Amount:     p.Amount,
		})
	}
	
	// Convert less particulars
	for _, p := range note.LessParticulars {
		proto.LessParticulars = append(proto.LessParticulars, &paymentnotepb.PaymentParticular{
			Particular: p.Particular,
			Amount:     p.Amount,
		})
	}
	
	// Convert approval logs
	for _, log := range note.ApprovalLogs {
		proto.ApprovalLogs = append(proto.ApprovalLogs, &paymentnotepb.PaymentApprovalLog{
			Id:       log.ID,
			Status:   log.Status,
			Comments: stringPtrToProto(log.Comments),
			Reviewer: &commonpb.RelatedUser{
				Id:    log.ReviewerID,
				Name:  stringPtrToProto(log.ReviewerName),
				Email: stringPtrToProto(log.ReviewerEmail),
				Designation: stringPtrToProto(log.ReviewerDesignation),
				Role:        stringPtrToProto(log.ReviewerRole),
			},
			CreatedAt: log.CreatedAt.Format(time.RFC3339),
		})
	}
	
	// Convert comments
	for _, c := range note.Comments {
		proto.Comments = append(proto.Comments, &paymentnotepb.PaymentNoteComment{
			Id:      c.ID,
			Comment: c.Comment,
			Status:  stringPtrToProto(c.Status),
			User: &commonpb.RelatedUser{
				Id:    c.UserID,
				Name:  stringPtrToProto(c.UserName),
				Email: stringPtrToProto(c.UserEmail),
				Designation: stringPtrToProto(c.UserDesignation),
				Role:        stringPtrToProto(c.UserRole),
			},
			CreatedAt: c.CreatedAt.Format(time.RFC3339),
		})
	}
	
	// Convert documents
	for _, d := range note.Documents {
		proto.Documents = append(proto.Documents, &paymentnotepb.PaymentNoteDocument{
			Id:               d.ID,
			FileName:         d.FileName,
			OriginalFilename: d.OriginalFilename,
			MimeType:         stringPtrToProto(d.MimeType),
			FileSize:         d.FileSize,
			ObjectKey:        d.ObjectKey,
			UploadedBy:       fmt.Sprintf("%d", d.UploadedBy),
			UploadedByName:   stringPtrToProto(d.UploadedByName),
			CreatedAt:        d.CreatedAt.Format(time.RFC3339),
		})
	}
	
	return proto
}

// DomainToSummaryProto converts domain model to summary proto response
func (s *paymentNoteService) DomainToSummaryProto(note *domain.PaymentNote) *paymentnotepb.PaymentNoteSummary {
	return &paymentnotepb.PaymentNoteSummary{
		Id:                 note.ID,
		NoteNo:             note.NoteNo,
		Status:             note.Status,
		NetPayableRoundOff: note.NetPayableRoundOff,
		IsDraft:            note.IsDraft,
		CreatedAt:          note.CreatedAt.Format(time.RFC3339),
		UpdatedAt:          note.UpdatedAt.Format(time.RFC3339),
		GreenNote: &commonpb.PaymentGreenNoteReference{
			Id:      protoStringVal(note.GreenNoteID),
			OrderNo: stringPtrToProto(note.GreenNoteNo),
		},
		Owner: &commonpb.RelatedUser{
			Id: note.UserID,
		},
	}
}

// Helper functions
func stringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func protoStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func protoInt64Ptr(i int64) *int64 {
	if i == 0 {
		return nil
	}
	return &i
}

func stringPtrToProto(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func int64PtrToProto(i *int64) int64 {
	if i == nil {
		return 0
	}
	return *i
}

func int32PtrToProto(i *int32) int32 {
	if i == nil {
		return 0
	}
	return *i
}

func timePtrToProto(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.Format(time.RFC3339)
}

func protoStringToInt64Val(s *string) int64 {
	if s == nil || *s == "" {
		return 0
	}
	var val int64
	fmt.Sscanf(*s, "%d", &val)
	return val
}

func protoStringToInt64Ptr(s string) *int64 {
	if s == "" || s == "0" {
		return nil
	}
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return &val
}

func protoToUserId(s string) int64 {
	var val int64
	fmt.Sscanf(s, "%d", &val)
	return val
}

func protoStringVal(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// extractApprover extracts the final approver name from the GreenNote approval history/remarks
// Assumes format might contain "Approved by [Name]" or similar logs
func extractApprover(history string) string {
	if history == "" {
		return ""
	}
	
	// Split by newlines to process each log entry
	lines := strings.Split(history, "\n")
	var approver string
	
	// Iterate through lines to find the last occurrence of "Approved" or "Approved by"
	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		if strings.Contains(lowerLine, "approved by") {
			// Try to extract name after "approved by"
			parts := strings.SplitAfterN(lowerLine, "approved by", 2)
			if len(parts) > 1 {
				approver = strings.TrimSpace(parts[1])
				// Clean up any trailing chars like parentheses if needed
				if idx := strings.Index(approver, "("); idx > 0 {
					approver = strings.TrimSpace(approver[:idx])
				}
			}
		}
	}
	
	return approver
}
