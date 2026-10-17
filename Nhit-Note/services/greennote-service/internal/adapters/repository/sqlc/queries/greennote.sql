-- name: CreateGreenNote :one
INSERT INTO green_notes (
    id,
    project_name,
    supplier_name,
    expense_category,
    protest_note_raised,
    whether_contract,
    extension_of_contract_period_executed,
    expense_amount_within_contract,
    milestone_achieved,
    payment_approved_with_deviation,
    required_documents_submitted,
    documents_verified,
    contract_start_date,
    contract_end_date,
    appointed_start_date,
    supply_period_start,
    supply_period_end,
    base_value,
    other_charges,
    gst,
    total_amount,
    enable_multiple_invoices,
    status,
    approval_for,
    department_name,
    work_order_no,
    po_number,
    work_order_date,
    expense_category_type,
    msme_classification,
    activity_type,
    brief_of_goods_services,
    delayed_damages,
    nature_of_expenses,
    contract_period_completed,
    budget_expenditure,
    actual_expenditure,
    expenditure_over_budget,
    milestone_remarks,
    specify_deviation,
    documents_workdone_supply,
    documents_discrepancy,
    remarks,
    auditor_remarks,
    amount_retained_for_non_submission,
    approver_name,
    approver_designation,
    approver_role,
    approved_at,
    approval_levels,
    payment_note_id
) VALUES (
    $1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,
    $18,$19,$20,$21,$22,$23,$24,$25,$26,$27,$28,$29,$30,$31,
    $32,$33,$34,$35,$36,$37,$38,$39,$40,$41,$42,$43,$44,$45,
    $46, $47, $48, $49, $50, $51
)
RETURNING *;

-- name: InsertInvoice :one
INSERT INTO green_note_invoices (
    id,
    green_note_id,
    invoice_number,
    invoice_date,
    taxable_value,
    gst,
    other_charges,
    invoice_value,
    org_id,
    tenant_id,
    is_primary
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
RETURNING *;

-- name: InsertSupportingDocument :one
INSERT INTO green_note_documents (
    id,
    green_note_id,
    name,
    original_filename,
    mime_type,
    file_size,
    object_key,
    org_id,
    tenant_id,
    document_type
) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
RETURNING *;

-- name: GetGreenNote :one
SELECT *
FROM green_notes
WHERE id = $1
LIMIT 1;

-- name: ListInvoicesByGreenNote :many
SELECT *
FROM green_note_invoices
WHERE green_note_id = $1
ORDER BY invoice_date ASC;

-- name: ListSupportingDocuments :many
SELECT *
FROM green_note_documents
WHERE green_note_id = $1
ORDER BY created_at ASC;

-- name: ListGreenNotes :many
SELECT
    id,
    project_name,
    supplier_name AS vendor_name,
    total_amount AS amount,
    TO_CHAR(created_at, 'YYYY-MM-DD') AS date,
    status,
    approval_levels,
    payment_note_id
FROM green_notes
WHERE ($1::text IS NULL OR status = $1)
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountGreenNotes :one
SELECT COUNT(*) FROM green_notes
WHERE ($1::text IS NULL OR status = $1);

-- name: UpdateGreenNote :one
UPDATE green_notes
SET
    project_name = $2,
    supplier_name = $3,
    expense_category = $4,
    protest_note_raised = $5,
    whether_contract = $6,
    extension_of_contract_period_executed = $7,
    expense_amount_within_contract = $8,
    milestone_achieved = $9,
    payment_approved_with_deviation = $10,
    required_documents_submitted = $11,
    documents_verified = $12,
    contract_start_date = $13,
    contract_end_date = $14,
    appointed_start_date = $15,
    supply_period_start = $16,
    supply_period_end = $17,
    base_value = $18,
    other_charges = $19,
    gst = $20,
    total_amount = $21,
    enable_multiple_invoices = $22,
    status = $23,
    approval_for = $24,
    department_name = $25,
    work_order_no = $26,
    po_number = $27,
    work_order_date = $28,
    expense_category_type = $29,
    msme_classification = $30,
    activity_type = $31,
    brief_of_goods_services = $32,
    delayed_damages = $33,
    nature_of_expenses = $34,
    contract_period_completed = $35,
    budget_expenditure = $36,
    actual_expenditure = $37,
    expenditure_over_budget = $38,
    milestone_remarks = $39,
    specify_deviation = $40,
    documents_workdone_supply = $41,
    documents_discrepancy = $42,
    remarks = $43,
    auditor_remarks = $44,
    amount_retained_for_non_submission = $45,
    approver_name = $46,
    approver_designation = $47,
    approver_role = $48,
    approved_at = $49,
    approval_levels = $50,
    payment_note_id = $51,
    updated_at = NOW()
WHERE id = $1
RETURNING *;


-- name: CancelGreenNote :one
UPDATE green_notes
SET status = 'CANCELLED',
    remarks = $2,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteInvoicesByGreenNoteID :exec
DELETE FROM green_note_invoices 
WHERE green_note_id = $1;

-- name: LinkPaymentNote :exec
UPDATE green_notes
SET payment_note_id = $2,
    updated_at = CURRENT_TIMESTAMP
WHERE id = $1;