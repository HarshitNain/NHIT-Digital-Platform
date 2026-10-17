package sqlc

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	greennotepb "github.com/ShristiRnr/Nhit-Note/api/pb/greennotepb"
	"encoding/json"
	sqlcgen "github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/adapters/repository/sqlc/generated"
	"github.com/ShristiRnr/Nhit-Note/services/greennote-service/internal/core/ports"
	"github.com/sqlc-dev/pqtype"

	timestamppb "google.golang.org/protobuf/types/known/timestamppb"

	"github.com/google/uuid"
)

// Repository is a minimal Postgres-backed implementation of ports.GreenNoteRepository.
type Repository struct {
	db   *sql.DB
	q    *sqlcgen.Queries
	docs ports.DocumentStorage
}

// NewPostgresGreenNoteRepository constructs a repository.
func NewPostgresGreenNoteRepository(db *sql.DB, docs ports.DocumentStorage) *Repository {
	return &Repository{db: db, q: sqlcgen.New(db), docs: docs}
}

func (r *Repository) List(ctx context.Context, req *greennotepb.ListGreenNotesRequest, orgID, tenantID string) (*greennotepb.ListGreenNotesResponse, error) {
	if r == nil || r.db == nil {
		return &greennotepb.ListGreenNotesResponse{}, nil
	}

	statusFilter := mapProtoEnumToDBStatus(req.GetStatus())
	includeAll := req.GetIncludeAll()
	page := req.GetPage()
	perPage := req.GetPerPage()
	if page <= 0 {
		page = 1
	}
	if perPage <= 0 {
		perPage = 20
	}
	offset := (page - 1) * perPage

	var total int64
	fmt.Println("\n========== ListGreenNotes DEBUG START ==========")
	fmt.Printf("Request Parameters:\n")
	fmt.Printf("  - Page: %d\n", page)
	fmt.Printf("  - PerPage: %d\n", perPage)
	fmt.Printf("  - Offset: %d\n", offset)
	fmt.Printf("  - OrgID: '%s'\n", orgID)
	fmt.Printf("  - TenantID: '%s'\n", tenantID)
	fmt.Printf("  - IncludeAll: %v\n", includeAll)
	fmt.Printf("  - Status Enum: %v\n", req.GetStatus())
	fmt.Printf("  - StatusFilter (mapped): '%s'\n", statusFilter)
	
	// Filter by org_id if provided. We assume multi-tenancy is required.
	countQuery := `SELECT COUNT(*) FROM green_notes WHERE org_id = $1`
	fmt.Printf("\n COUNT Query (Status Filter Removed):\n")
	fmt.Printf("  SQL: %s\n", countQuery)
	fmt.Printf("  Params: [$1='%s']\n", orgID)
	if err := r.db.QueryRowContext(ctx, countQuery, orgID).Scan(&total); err != nil {
		fmt.Printf("COUNT Query Error: %v\n", err)
		return nil, err
	}
	fmt.Printf("COUNT Query Result: %d total records\n", total)

	query := `
		SELECT id, project_name, supplier_name, total_amount, created_at, status, approval_levels, payment_note_id
		FROM green_notes
		WHERE org_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	fmt.Printf("\n SELECT Query (Status Filter Removed):\n")
	fmt.Printf("  SQL: %s\n", query)
	fmt.Printf("  Params: [$1='%s', $2=%d, $3=%d]\n", orgID, perPage, offset)
	rows, err := r.db.QueryContext(ctx, query, orgID, perPage, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := make([]*greennotepb.GreenNoteListItem, 0)
	for rows.Next() {
		var (
			id          string
			projectName sql.NullString
			vendorName  sql.NullString
			amount      float64
			created     time.Time
			status      string
			approvalLevels pqtype.NullRawMessage
			paymentNoteID  sql.NullString
		)
		if err := rows.Scan(&id, &projectName, &vendorName, &amount, &created, &status, &approvalLevels, &paymentNoteID); err != nil {
			return nil, err
		}
		item := &greennotepb.GreenNoteListItem{
			Id:          id,
			ProjectName: projectName.String,
			VendorName:  vendorName.String,
			Amount:      amount,
			Date:        created.Format(time.RFC3339),
			Status:      mapDBStatusToProtoEnum(status),
			ApprovalLevels: fromRawMessage(approvalLevels),
			PaymentNoteId:  paymentNoteID.String,
		}
		items = append(items, item)
	}
	if err := rows.Err(); err != nil {
		fmt.Printf(" Rows Error: %v\n", err)
		return nil, err
	}
	fmt.Printf("\n SELECT Query Result: Retrieved %d items\n", len(items))
	for i, item := range items {
		fmt.Printf("  [%d] ID=%s, Project=%s, Vendor=%s, Amount=%.2f, Status=%s\n", 
			i+1, item.Id, item.ProjectName, item.VendorName, item.Amount, item.Status)
	}
	fmt.Printf("========== ListGreenNotes DEBUG END ==========\n\n")

	totalPages := (int32(total) + perPage - 1) / perPage

	return &greennotepb.ListGreenNotesResponse{
		Notes:   items,
		Page:    page,
		PerPage: perPage,
		Total:   total,
		Pagination: &greennotepb.PaginationMetadata{
			CurrentPage: page,
			PageSize:    perPage,
			TotalItems:  total,
			TotalPages:  totalPages,
		},
	}, nil
}

func (r *Repository) Get(ctx context.Context, id string) (*greennotepb.GreenNotePayload, string, string, error) {
	if r == nil || r.db == nil || r.q == nil {
		return nil, "", "", ports.ErrNotFound
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return nil, "", "", ports.ErrNotFound
	}

	var orgID, tenantID string
	query := `SELECT org_id, tenant_id FROM green_notes WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, id).Scan(&orgID, &tenantID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, "", "", ports.ErrNotFound
		}
		return nil, "", "", err
	}

	noteRow, err := r.q.GetGreenNote(ctx, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, "", "", ports.ErrNotFound
	}
	if err != nil {
		return nil, "", "", err
	}

	// Use generated method for invoices
	invoicesRows, err := r.q.ListInvoicesByGreenNote(ctx, id)
	if err != nil {
		fmt.Printf(" Failed to list invoices for hydration: %v\n", err)
	}
	fmt.Printf(" DEBUG HYDRATION: Retrieved %d total invoice records for GN-%s\n", len(invoicesRows), id)

	docs, _ := r.q.ListSupportingDocuments(ctx, id)

	p := &greennotepb.GreenNotePayload{
		ProjectName:                       noteRow.ProjectName.String,
		SupplierName:                      noteRow.SupplierName.String,
		ExpenseCategory:                   noteRow.ExpenseCategory.String,
		ProtestNoteRaised:                 fromNullYesNoEnum(noteRow.ProtestNoteRaised),
		WhetherContract:                   fromNullYesNoEnum(noteRow.WhetherContract),
		ExtensionOfContractPeriodExecuted: fromNullYesNoEnum(noteRow.ExtensionOfContractPeriodExecuted),
		ExpenseAmountWithinContract:       fromNullYesNoEnum(noteRow.ExpenseAmountWithinContract),
		MilestoneAchieved:                 fromNullYesNoEnum(noteRow.MilestoneAchieved),
		PaymentApprovedWithDeviation:      fromNullYesNoEnum(noteRow.PaymentApprovedWithDeviation),
		RequiredDocumentsSubmitted:        fromNullYesNoEnum(noteRow.RequiredDocumentsSubmitted),
		DocumentsVerified:                 noteRow.DocumentsVerified.String,
		ContractStartDate:                 noteRow.ContractStartDate.String,
		ContractEndDate:                   noteRow.ContractEndDate.String,
		AppointedStartDate:                noteRow.AppointedStartDate.String,
		SupplyPeriodStart:                 noteRow.SupplyPeriodStart.String,
		SupplyPeriodEnd:                   noteRow.SupplyPeriodEnd.String,
		ContractPeriodCompleted:           fromNullYesNoEnum(noteRow.ContractPeriodCompleted),
		DepartmentName:                    noteRow.DepartmentName.String,
		WorkOrderNo:                       noteRow.WorkOrderNo.String,
		PoNumber:                          noteRow.PoNumber.String,
		MsmeClassification:                noteRow.MsmeClassification.String,
		ActivityType:                      noteRow.ActivityType.String,
		BriefOfGoodsServices:              noteRow.BriefOfGoodsServices.String,
		DetailedStatus:                    noteRow.Status.String,
		Status:                            mapDBStatusToProtoEnum(noteRow.Status.String),
		EnableMultipleInvoices:            noteRow.EnableMultipleInvoices,
		WorkOrderDate:                     noteRow.WorkOrderDate.String,
		MilestoneRemarks:                  noteRow.MilestoneRemarks.String,
		SpecifyDeviation:                  noteRow.SpecifyDeviation.String,
		DocumentsWorkdoneSupply:           noteRow.DocumentsWorkdoneSupply.String,
		DocumentsDiscrepancy:              noteRow.DocumentsDiscrepancy.String,
		Remarks:                           noteRow.Remarks.String,
		AuditorRemarks:                    noteRow.AuditorRemarks.String,
		ApprovalLevels:                    fromRawMessage(noteRow.ApprovalLevels),
		CreatedAt:                         timestamppb.New(noteRow.CreatedAt),
		UpdatedAt:                         timestamppb.New(noteRow.UpdatedAt),
		ApproverName:                      noteRow.ApproverName.String,
		ApproverDesignation:               noteRow.ApproverDesignation.String,
		ApproverRole:                      noteRow.ApproverRole.String,
		PaymentNoteId:                     noteRow.PaymentNoteID.String,
	}

	if noteRow.ApprovedAt.Valid {
		p.ApprovedAt = timestamppb.New(noteRow.ApprovedAt.Time)
	}

	p.BaseValue = parseDecimal(noteRow.BaseValue.String)
	p.Gst = parseDecimal(noteRow.Gst.String)
	p.OtherCharges = parseDecimal(noteRow.OtherCharges.String)
	p.TotalAmount = parseDecimal(noteRow.TotalAmount.String)
	p.BudgetExpenditure = parseDecimal(noteRow.BudgetExpenditure.String)
	p.ActualExpenditure = parseDecimal(noteRow.ActualExpenditure.String)
	p.ExpenditureOverBudget = parseDecimal(noteRow.ExpenditureOverBudget.String)
	p.AmountRetainedForNonSubmission = parseDecimal(noteRow.AmountRetainedForNonSubmission.String)

	// Enum mappings from text
	switch noteRow.ApprovalFor.String {
	case "APPROVAL_FOR_INVOICE":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_INVOICE
	case "APPROVAL_FOR_ADVANCE":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_ADVANCE
	case "APPROVAL_FOR_ADHOC":
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_ADHOC
	default:
		p.ApprovalFor = greennotepb.ApprovalFor_APPROVAL_FOR_UNSPECIFIED
	}

	switch noteRow.NatureOfExpenses.String {
	case "NATURE_OHC_001_MANPOWER":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_001_MANPOWER
	case "NATURE_OHC_002_STAFF_WELFARE":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_002_STAFF_WELFARE
	case "NATURE_OHC_003_OFFICE_RENT_UTILITIES":
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OHC_003_OFFICE_RENT_UTILITIES
	default:
		p.NatureOfExpenses = greennotepb.NatureOfExpenses_NATURE_OF_EXPENSES_UNSPECIFIED
	}

	if noteRow.ExpenseCategoryType.Valid {
		switch noteRow.ExpenseCategoryType.String {
		case "EXPENSE_CATEGORY_CAPITAL":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_CAPITAL
		case "EXPENSE_CATEGORY_REVENUE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_REVENUE
		case "EXPENSE_CATEGORY_OPERATIONAL":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_OPERATIONAL
		case "EXPENSE_CATEGORY_ADMINISTRATIVE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_ADMINISTRATIVE
		case "EXPENSE_CATEGORY_MAINTENANCE":
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_MAINTENANCE
		default:
			p.ExpenseCategoryType = greennotepb.ExpenseCategoryType_EXPENSE_CATEGORY_UNSPECIFIED
		}
	}

	for _, in := range invoicesRows {
		inv := &greennotepb.InvoiceInput{
			InvoiceNumber: in.InvoiceNumber,
			InvoiceDate:   in.InvoiceDate.String,
			TaxableValue:  parseDecimal(in.TaxableValue),
			Gst:           parseDecimal(in.Gst),
			OtherCharges:  parseDecimal(in.OtherCharges),
			InvoiceValue:  parseDecimal(in.InvoiceValue),
		}
		isPrimary := in.IsPrimary.Valid && in.IsPrimary.Bool
		fmt.Printf(" DEBUG HYDRATION: Invoice %s, RawIsPrimary(Valid=%v, Bool=%v) -> Evaluated=%v\n", in.InvoiceNumber, in.IsPrimary.Valid, in.IsPrimary.Bool, isPrimary)
		if isPrimary {
			if p.Invoice != nil {
				fmt.Printf("DEBUG HYDRATION: Overwriting existing primary invoice %s with %s\n", p.Invoice.InvoiceNumber, in.InvoiceNumber)
			}
			p.Invoice = inv
		} else {
			p.Invoices = append(p.Invoices, inv)
		}
	}
	
	if p.Invoice == nil && len(p.Invoices) > 0 {
		fmt.Printf("DEBUG HYDRATION: No primary invoice flagged for GN-%s. Promoting first invoice (%s) to primary.\n", id, p.Invoices[0].InvoiceNumber)
		p.Invoice = p.Invoices[0]
		p.Invoices = p.Invoices[1:]
	}
	
	if p.Invoice == nil {
		fmt.Printf(" DEBUG HYDRATION: Still no invoice data found for GN-%s after processing %d records\n", id, len(invoicesRows))
	}

	for _, d := range docs {
		doc := &greennotepb.SupportingDocument{
			Id:               d.ID,
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			FileSize:         d.FileSize,
			ObjectKey:        d.ObjectKey,
			CreatedAt:        timestamppb.New(d.CreatedAt),
			UpdatedAt:        timestamppb.New(d.UpdatedAt),
			DocumentType:     d.DocumentType.String,
		}
		p.ExistingDocuments = append(p.ExistingDocuments, doc)

		// Map to categorized fields as Upload metadata carriers
		catDoc := &greennotepb.SupportingDocumentUpload{
			Id:               d.ID,
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			DocumentType:     d.DocumentType.String,
			ObjectKey:        d.ObjectKey,
		}

		switch strings.ToLower(d.DocumentType.String) {
		case "additional":
			p.AdditionalInfoDocument = catDoc
		case "contract_extension":
			p.ContractExtensionDocument = catDoc
		case "hr":
			p.HrDepartmentDocument = catDoc
		}
	}

	return p, orgID, tenantID, nil
}

func (r *Repository) Create(ctx context.Context, payload *greennotepb.GreenNotePayload, orgID, tenantID string) (string, error) {
	if r == nil || r.db == nil || r.q == nil || payload == nil {
		return "", nil
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return "", err
	}
	qtx := r.q.WithTx(tx)
	// Using manual SQL for Create.

	invBase := payload.GetBaseValue()
	invGst := payload.GetGst()
	invOther := payload.GetOtherCharges()
	invTotal := payload.GetTotalAmount()

	brief := payload.GetBriefOfGoodsServices()
	if brief == "" {
		brief = payload.GetProjectName()
	}

	status := strings.TrimSpace(payload.GetDetailedStatus())
	if status == "" {
		status = "pending"
	}

	// We'll use a manual query here instead of sqlc because we can't easily regenerate sqlc for org_id/tenant_id
	createNoteQuery := `
		INSERT INTO green_notes (
			id, project_name, supplier_name, expense_category, 
			protest_note_raised, whether_contract, extension_of_contract_period_executed,
			expense_amount_within_contract, milestone_achieved, payment_approved_with_deviation,
			required_documents_submitted, contract_period_completed, documents_verified,
			contract_start_date, contract_end_date, appointed_start_date,
			supply_period_start, supply_period_end, base_value, gst, other_charges,
			total_amount, enable_multiple_invoices, status, approval_for,
			department_name, work_order_no, po_number, work_order_date,
			expense_category_type, msme_classification, activity_type,
			brief_of_goods_services, delayed_damages, nature_of_expenses,
			budget_expenditure, actual_expenditure, expenditure_over_budget,
			milestone_remarks, specify_deviation, documents_workdone_supply,
			documents_discrepancy, remarks, auditor_remarks, 
			amount_retained_for_non_submission, approver_name, 
			approver_designation, approver_role, approved_at,
			approval_levels, org_id, tenant_id, payment_note_id
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, 
			$19, $20, $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, 
			$35, $36, $37, $38, $39, $40, $41, $42, $43, $44, $45, $46, $47, $48, $49, $50, $51, $52, $53
		) RETURNING id
	`
	id := uuid.NewString()
	var returnedID string
	err = tx.QueryRowContext(ctx, createNoteQuery,
		id, toNullString(payload.GetProjectName()), toNullString(payload.GetSupplierName()), toNullString(payload.GetExpenseCategory()),
		toNullYesNoEnum(payload.GetProtestNoteRaised()), toNullYesNoEnum(payload.GetWhetherContract()), toNullYesNoEnum(payload.GetExtensionOfContractPeriodExecuted()),
		toNullYesNoEnum(payload.GetExpenseAmountWithinContract()), toNullYesNoEnum(payload.GetMilestoneAchieved()), toNullYesNoEnum(payload.GetPaymentApprovedWithDeviation()),
		toNullYesNoEnum(payload.GetRequiredDocumentsSubmitted()), toNullYesNoEnum(payload.GetContractPeriodCompleted()), toNullString(payload.GetDocumentsVerified()),
		toNullString(payload.GetContractStartDate()), toNullString(payload.GetContractEndDate()), toNullString(payload.GetAppointedStartDate()),
		toNullString(payload.GetSupplyPeriodStart()), toNullString(payload.GetSupplyPeriodEnd()), formatDecimal(invBase), formatDecimal(invGst), formatDecimal(invOther),
		formatDecimal(invTotal), payload.GetEnableMultipleInvoices(), status, payload.GetApprovalFor().String(),
		toNullString(payload.GetDepartmentName()), toNullString(payload.GetWorkOrderNo()), toNullString(payload.GetPoNumber()), toNullString(payload.GetWorkOrderDate()),
		payload.GetExpenseCategoryType().String(), toNullString(payload.GetMsmeClassification()), toNullString(payload.GetActivityType()),
		brief, toNullString(payload.GetDelayedDamages()), payload.GetNatureOfExpenses().String(),
		formatDecimal(payload.GetBudgetExpenditure()), formatDecimal(payload.GetActualExpenditure()), formatDecimal(payload.GetExpenditureOverBudget()),
		toNullString(payload.GetMilestoneRemarks()), toNullString(payload.GetSpecifyDeviation()), toNullString(payload.GetDocumentsWorkdoneSupply()),
		toNullString(payload.GetDocumentsDiscrepancy()), toNullString(payload.GetRemarks()), toNullString(payload.GetAuditorRemarks()),
		formatDecimal(payload.GetAmountRetainedForNonSubmission()), 
		toNullString(payload.GetApproverName()), toNullString(payload.GetApproverDesignation()),
		toNullString(payload.GetApproverRole()), toNullTime(payload.GetApprovedAt()),
		toRawMessage(payload.GetApprovalLevels()), orgID, tenantID, toNullString(payload.GetPaymentNoteId()),
	).Scan(&returnedID)

	if err != nil {
		_ = tx.Rollback()
		return "", err
	}

	// Handle Combo Payload: Insert Primary Invoice and Multiple Invoices
	// 1. Primary Invoice
	primaryNumber := ""
	if inv := payload.GetInvoice(); inv != nil {
		primaryNumber = strings.TrimSpace(inv.GetInvoiceNumber())
		fmt.Printf("PERSISTENCE: Inserting primary invoice %s for GN-%s\n", primaryNumber, returnedID)
		_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
			ID:            uuid.NewString(),
			GreenNoteID:   returnedID,
			InvoiceNumber: primaryNumber,
			InvoiceDate:   toNullString(inv.GetInvoiceDate()),
			TaxableValue:  formatDecimal(inv.GetTaxableValue()),
			Gst:           formatDecimal(inv.GetGst()),
			OtherCharges:  formatDecimal(inv.GetOtherCharges()),
			InvoiceValue:  formatDecimal(inv.GetInvoiceValue()),
			OrgID:         toUUID(orgID),
			TenantID:      toUUID(tenantID),
			IsPrimary:     sql.NullBool{Valid: true, Bool: true},
		})
		fmt.Printf("PERSISTENCE: Saved primary invoice %s with IsPrimary=true\n", primaryNumber)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}

	// 2. Multiple Invoices (with de-duplication)
	insertedNumbers := make(map[string]bool)
	if primaryNumber != "" {
		insertedNumbers[primaryNumber] = true
	}

	for _, in := range payload.GetInvoices() {
		if in == nil {
			continue
		}
		invNum := strings.TrimSpace(in.GetInvoiceNumber())
		if active := insertedNumbers[invNum]; active {
			fmt.Printf("PERSISTENCE: Skipping duplicate invoice %s for GN-%s\n", invNum, returnedID)
			continue
		}
		insertedNumbers[invNum] = true

		fmt.Printf("PERSISTENCE: Inserting multiple invoice %s for GN-%s\n", invNum, returnedID)
		_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
			ID:            uuid.NewString(),
			GreenNoteID:   returnedID,
			InvoiceNumber: invNum,
			InvoiceDate:   toNullString(in.GetInvoiceDate()),
			TaxableValue:  formatDecimal(in.GetTaxableValue()),
			Gst:           formatDecimal(in.GetGst()),
			OtherCharges:  formatDecimal(in.GetOtherCharges()),
			InvoiceValue:  formatDecimal(in.GetInvoiceValue()),
			OrgID:         toUUID(orgID),
			TenantID:      toUUID(tenantID),
			IsPrimary:     sql.NullBool{Valid: true, Bool: false},
		})
		fmt.Printf("PERSISTENCE: Saved multiple invoice %s with IsPrimary=false\n", invNum)
		if err != nil {
			_ = tx.Rollback()
			return "", err
		}
	}

	if err := tx.Commit(); err != nil {
		return "", err
	}
	return returnedID, nil
}

func (r *Repository) Update(ctx context.Context, id string, payload *greennotepb.GreenNotePayload, orgID, tenantID string) error {
	if r == nil || r.db == nil || r.q == nil || payload == nil {
		return nil
	}

	id = strings.TrimSpace(id)
	if id == "" {
		return ports.ErrNotFound
	}

	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := r.q.WithTx(tx)

	invBase := payload.GetBaseValue()
	invGst := payload.GetGst()
	invOther := payload.GetOtherCharges()
	invTotal := payload.GetTotalAmount()

	brief := payload.GetBriefOfGoodsServices()
	if brief == "" {
		brief = payload.GetProjectName()
	}

	status := strings.TrimSpace(payload.GetDetailedStatus())
	if status == "" {
		status = "pending"
	}

	_, err = qtx.UpdateGreenNote(ctx, sqlcgen.UpdateGreenNoteParams{
		ID:                                id,
		ProjectName:                       toNullString(payload.GetProjectName()),
		SupplierName:                      toNullString(payload.GetSupplierName()),
		ExpenseCategory:                   toNullString(payload.GetExpenseCategory()),
		ProtestNoteRaised:                 toNullYesNoEnum(payload.GetProtestNoteRaised()),
		WhetherContract:                   toNullYesNoEnum(payload.GetWhetherContract()),
		ExtensionOfContractPeriodExecuted: toNullYesNoEnum(payload.GetExtensionOfContractPeriodExecuted()),
		ExpenseAmountWithinContract:       toNullYesNoEnum(payload.GetExpenseAmountWithinContract()),
		MilestoneAchieved:                 toNullYesNoEnum(payload.GetMilestoneAchieved()),
		PaymentApprovedWithDeviation:      toNullYesNoEnum(payload.GetPaymentApprovedWithDeviation()),
		RequiredDocumentsSubmitted:        toNullYesNoEnum(payload.GetRequiredDocumentsSubmitted()),
		ContractPeriodCompleted:           toNullYesNoEnum(payload.GetContractPeriodCompleted()),
		DocumentsVerified:                 toNullString(payload.GetDocumentsVerified()),
		ContractStartDate:                 toNullString(payload.GetContractStartDate()),
		ContractEndDate:                   toNullString(payload.GetContractEndDate()),
		AppointedStartDate:                toNullString(payload.GetAppointedStartDate()),
		SupplyPeriodStart:                 toNullString(payload.GetSupplyPeriodStart()),
		SupplyPeriodEnd:                   toNullString(payload.GetSupplyPeriodEnd()),
		BaseValue:                         toNullString(formatDecimal(invBase)),
		Gst:                               toNullString(formatDecimal(invGst)),
		OtherCharges:                      toNullString(formatDecimal(invOther)),
		TotalAmount:                       toNullString(formatDecimal(invTotal)),
		EnableMultipleInvoices:            payload.GetEnableMultipleInvoices(),
		Status:                            toNullString(status),
		ApprovalFor:                       toNullString(payload.GetApprovalFor().String()),
		DepartmentName:                    toNullString(payload.GetDepartmentName()),
		WorkOrderNo:                       toNullString(payload.GetWorkOrderNo()),
		PoNumber:                          toNullString(payload.GetPoNumber()),
		WorkOrderDate:                     toNullString(payload.GetWorkOrderDate()),
		ExpenseCategoryType:               toNullString(payload.GetExpenseCategoryType().String()),
		MsmeClassification:                toNullString(payload.GetMsmeClassification()),
		ActivityType:                      toNullString(payload.GetActivityType()),
		BriefOfGoodsServices:              toNullString(brief),
		DelayedDamages:                    toNullString(payload.GetDelayedDamages()),
		NatureOfExpenses:                  toNullString(payload.GetNatureOfExpenses().String()),
		BudgetExpenditure:                 toNullString(formatDecimal(payload.GetBudgetExpenditure())),
		ActualExpenditure:                 toNullString(formatDecimal(payload.GetActualExpenditure())),
		ExpenditureOverBudget:             toNullString(formatDecimal(payload.GetExpenditureOverBudget())),
		MilestoneRemarks:                  toNullString(payload.GetMilestoneRemarks()),
		SpecifyDeviation:                  toNullString(payload.GetSpecifyDeviation()),
		DocumentsWorkdoneSupply:           toNullString(payload.GetDocumentsWorkdoneSupply()),
		DocumentsDiscrepancy:              toNullString(payload.GetDocumentsDiscrepancy()),
		Remarks:                           toNullString(payload.GetRemarks()),
		AuditorRemarks:                    toNullString(payload.GetAuditorRemarks()),
		AmountRetainedForNonSubmission:    toNullString(formatDecimal(payload.GetAmountRetainedForNonSubmission())),
		ApproverName:                      toNullString(payload.GetApproverName()),
		ApproverDesignation:               toNullString(payload.GetApproverDesignation()),
		ApproverRole:                      toNullString(payload.GetApproverRole()),
		ApprovedAt:                        toNullTime(payload.GetApprovedAt()),
		ApprovalLevels:                    pqtype.NullRawMessage{RawMessage: toRawMessage(payload.GetApprovalLevels()), Valid: true},
		PaymentNoteID:                     toNullString(payload.GetPaymentNoteId()),
	})
	if err != nil {
		_ = tx.Rollback()
		return err
	}

	// Handle Combo Payload in Update
	_ = qtx.DeleteInvoicesByGreenNoteID(ctx, id)
	
	// 1. Primary Invoice
	primaryNumber := ""
	if inv := payload.GetInvoice(); inv != nil {
		primaryNumber = strings.TrimSpace(inv.GetInvoiceNumber())
		fmt.Printf("💾 PERSISTENCE: Updating/Inserting primary invoice %s for GN-%s\n", primaryNumber, id)
		_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
			ID:            uuid.NewString(),
			GreenNoteID:   id,
			InvoiceNumber: primaryNumber,
			InvoiceDate:   toNullString(inv.GetInvoiceDate()),
			TaxableValue:  formatDecimal(inv.GetTaxableValue()),
			Gst:           formatDecimal(inv.GetGst()),
			OtherCharges:  formatDecimal(inv.GetOtherCharges()),
			InvoiceValue:  formatDecimal(inv.GetInvoiceValue()),
			OrgID:         toUUID(orgID),
			TenantID:      toUUID(tenantID),
			IsPrimary:     sql.NullBool{Valid: true, Bool: true},
		})
		fmt.Printf("💾 PERSISTENCE (Update): Saved primary invoice %s with IsPrimary=true\n", primaryNumber)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	// 2. Multiple Invoices (with de-duplication)
	insertedNumbers := make(map[string]bool)
	if primaryNumber != "" {
		insertedNumbers[primaryNumber] = true
	}

	for _, in := range payload.GetInvoices() {
		if in == nil {
			continue
		}
		invNum := strings.TrimSpace(in.GetInvoiceNumber())
		if active := insertedNumbers[invNum]; active {
			fmt.Printf("💾 PERSISTENCE: Skipping duplicate invoice %s for GN-%s\n", invNum, id)
			continue
		}
		insertedNumbers[invNum] = true

		fmt.Printf("💾 PERSISTENCE: Updating/Inserting multiple invoice %s for GN-%s\n", invNum, id)
		_, err := qtx.InsertInvoice(ctx, sqlcgen.InsertInvoiceParams{
			ID:            uuid.NewString(),
			GreenNoteID:   id,
			InvoiceNumber: invNum,
			InvoiceDate:   toNullString(in.GetInvoiceDate()),
			TaxableValue:  formatDecimal(in.GetTaxableValue()),
			Gst:           formatDecimal(in.GetGst()),
			OtherCharges:  formatDecimal(in.GetOtherCharges()),
			InvoiceValue:  formatDecimal(in.GetInvoiceValue()),
			OrgID:         toUUID(orgID),
			TenantID:      toUUID(tenantID),
			IsPrimary:     sql.NullBool{Valid: true, Bool: false},
		})
		fmt.Printf("PERSISTENCE (Update): Saved multiple invoice %s with IsPrimary=false\n", invNum)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (r *Repository) AddSupportingDocuments(ctx context.Context, noteID string, docs []*greennotepb.SupportingDocument, orgID, tenantID string) error {
	if r == nil || r.db == nil || r.q == nil {
		return nil
	}
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	qtx := r.q.WithTx(tx)

	for _, d := range docs {
		_, err := qtx.InsertSupportingDocument(ctx, sqlcgen.InsertSupportingDocumentParams{
			ID:               uuid.NewString(),
			GreenNoteID:      noteID,
			Name:             d.Name,
			OriginalFilename: d.OriginalFilename,
			MimeType:         d.MimeType,
			FileSize:         d.FileSize,
			ObjectKey:        d.ObjectKey,
			OrgID:            toUUID(orgID),
			TenantID:         toUUID(tenantID),
			DocumentType:     toNullString(d.DocumentType),
		})
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func (r *Repository) Cancel(ctx context.Context, id string, reason string, orgID, tenantID string) error {
	if r == nil || r.db == nil {
		return nil
	}
	id = strings.TrimSpace(id)
	if id == "" {
		return ports.ErrNotFound
	}
	query := `
		UPDATE green_notes SET
			status = 'cancelled',
			remarks = CASE WHEN remarks IS NULL OR remarks = '' THEN $2 ELSE remarks || ' | cancel: ' || $2 END,
			updated_at = NOW()
		WHERE id = $1 AND org_id = $3
	`
	res, err := r.db.ExecContext(ctx, query, id, reason, orgID)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return ports.ErrNotFound
	}
	return nil
}

// --- helpers ---

func fromNullYesNoEnum(v sqlcgen.NullYesNoEnum) greennotepb.YesNo {
	if !v.Valid {
		return greennotepb.YesNo_YES_NO_UNSPECIFIED
	}
	switch v.YesNoEnum {
	case sqlcgen.YesNoEnumYES:
		return greennotepb.YesNo_YES
	case sqlcgen.YesNoEnumNO:
		return greennotepb.YesNo_NO
	default:
		return greennotepb.YesNo_YES_NO_UNSPECIFIED
	}
}

func (r *Repository) insertDocumentsTx(ctx context.Context, tx *sql.Tx, noteID string, uploads []*greennotepb.SupportingDocumentUpload, orgID, tenantID string) error {
	if r == nil || r.docs == nil {
		return nil
	}
	qtx := r.q.WithTx(tx)
	for _, u := range uploads {
		if u == nil || len(u.GetFileContent()) == 0 {
			continue
		}
		objectKey := fmt.Sprintf("note-%s/%d-%s", noteID, time.Now().UnixNano(), u.GetOriginalFilename())
		if err := r.docs.Save(ctx, objectKey, u.GetFileContent(), u.GetMimeType()); err != nil {
			return err
		}
		
		_, err := qtx.InsertSupportingDocument(ctx, sqlcgen.InsertSupportingDocumentParams{
			ID:               uuid.NewString(),
			GreenNoteID:      noteID,
			Name:             u.GetName(),
			OriginalFilename: u.GetOriginalFilename(),
			MimeType:         u.GetMimeType(),
			FileSize:         int64(len(u.GetFileContent())),
			ObjectKey:        objectKey,
			OrgID:            toUUID(orgID),
			TenantID:         toUUID(tenantID),
		})
		if err != nil {
			return err
		}
	}
	return nil
}

func toNullYesNoEnum(v greennotepb.YesNo) sqlcgen.NullYesNoEnum {
	switch v {
	case greennotepb.YesNo_YES:
		return sqlcgen.NullYesNoEnum{YesNoEnum: sqlcgen.YesNoEnumYES, Valid: true}
	case greennotepb.YesNo_NO:
		return sqlcgen.NullYesNoEnum{YesNoEnum: sqlcgen.YesNoEnumNO, Valid: true}
	default:
		return sqlcgen.NullYesNoEnum{}
	}
}

func toNullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}

func mapProtoEnumToDBStatus(e greennotepb.Status) string {
	switch e {
	case greennotepb.Status_STATUS_APPROVED:
		return "approved"
	case greennotepb.Status_STATUS_PENDING:
		return "pending"
	case greennotepb.Status_STATUS_REJECTED:
		return "rejected"
	case greennotepb.Status_STATUS_DRAFT:
		return "draft"
	case greennotepb.Status_STATUS_CANCELLED:
		return "cancelled"
	default:
		return ""
	}
}

func toNullTime(ts *timestamppb.Timestamp) sql.NullTime {
	if ts == nil {
		return sql.NullTime{Valid: false}
	}
	return sql.NullTime{Time: ts.AsTime(), Valid: true}
}

func mapDBStatusToProtoEnum(dbStatus string) greennotepb.Status {
	s := strings.ToLower(strings.TrimSpace(dbStatus))
	if strings.HasPrefix(s, "pending") || s == "0" {
		return greennotepb.Status_STATUS_PENDING
	}
	if strings.Contains(s, "approved") || s == "1" {
		return greennotepb.Status_STATUS_APPROVED
	}
	if strings.Contains(s, "rejected") || strings.Contains(s, "reject") || s == "2" {
		return greennotepb.Status_STATUS_REJECTED
	}
	if strings.Contains(s, "cancelled") || s == "4" {
		return greennotepb.Status_STATUS_CANCELLED
	}
	if s == "draft" || s == "3" {
		return greennotepb.Status_STATUS_DRAFT
	}
	return greennotepb.Status_STATUS_PENDING
}


func formatDecimal(v float64) string {
	return fmt.Sprintf("%.2f", v)
}

func parseDecimal(s string) float64 {
	if strings.TrimSpace(s) == "" {
		return 0
	}
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
func toUUID(s string) uuid.NullUUID {
	u, err := uuid.Parse(s)
	if err != nil {
		return uuid.NullUUID{Valid: false}
	}
	return uuid.NullUUID{UUID: u, Valid: true}
}
func toRawMessage(s string) json.RawMessage {
	if s == "" || s == "null" || s == "[]" {
		return json.RawMessage(`{"history": []}`)
	}
	return json.RawMessage(s)
}

func fromRawMessage(m pqtype.NullRawMessage) string {
	if !m.Valid || len(m.RawMessage) == 0 || string(m.RawMessage) == "null" {
		return `{"history": []}`
	}
	return string(m.RawMessage)
}

func (r *Repository) LinkPaymentNote(ctx context.Context, id string, paymentNoteID string) error {
	return r.q.LinkPaymentNote(ctx, sqlcgen.LinkPaymentNoteParams{
		ID:            id,
		PaymentNoteID: sql.NullString{String: paymentNoteID, Valid: paymentNoteID != ""},
	})
}
