-- name: CreatePaymentNote :one
INSERT INTO payment_notes (
    user_id,
    green_note_id,
    green_note_no,
    green_note_approver,
    green_note_app_date,
    reimbursement_note_id,
    note_no,
    subject,
    date,
    department,
    vendor_code,
    vendor_name,
    project_name,
    invoice_no,
    invoice_date,
    invoice_amount,
    invoice_approved_by,
    loa_po_no,
    loa_po_amount,
    loa_po_date,
    gross_amount,
    total_additions,
    total_deductions,
    net_payable_amount,
    net_payable_round_off,
    net_payable_words,
    tds_percentage,
    tds_section,
    tds_amount,
    account_holder_name,
    bank_name,
    account_number,
    ifsc_code,
    recommendation_of_payment,
    status,
    is_draft,
    auto_created,
    created_by,
    created_by_name,
    created_by_designation,
    created_by_role,
    created_by_signature,
    taxable_amount,
    gst_amount,
    other_charges,
    green_note_history
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10,
    $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30,
    $31, $32, $33, $34, $35, $36, $37, $38, $39, $40, $41,
    $42, $43, $44, $45, $46
)
RETURNING *;

-- name: GetPaymentNoteByID :one
SELECT * FROM payment_notes
WHERE id = $1
LIMIT 1;

-- name: GetPaymentNoteByGreenNoteID :one
SELECT * FROM payment_notes
WHERE green_note_id = $1
LIMIT 1;

-- name: GetPaymentNoteByNoteNo :one
SELECT * FROM payment_notes
WHERE note_no = $1
LIMIT 1;

-- name: ListPaymentNotes :many
SELECT * FROM payment_notes
WHERE
    (($1::text IS NULL OR $1::text = '') OR status = $1::payment_note_status) AND
    ($2::boolean IS NULL OR is_draft = $2) AND
    ($3::text IS NULL OR (
        note_no ILIKE '%' || $3 || '%' OR
        subject ILIKE '%' || $3 || '%' OR
        vendor_name ILIKE '%' || $3 || '%' OR
        project_name ILIKE '%' || $3 || '%'
    ))
ORDER BY created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountPaymentNotes :one
SELECT COUNT(*) FROM payment_notes
WHERE
    (($1::text IS NULL OR $1::text = '') OR status = $1::payment_note_status) AND
    ($2::boolean IS NULL OR is_draft = $2) AND
    ($3::text IS NULL OR (
        note_no ILIKE '%' || $3 || '%' OR
        subject ILIKE '%' || $3 || '%' OR
        vendor_name ILIKE '%' || $3 || '%' OR
        project_name ILIKE '%' || $3 || '%'
    ));

-- name: UpdatePaymentNote :one
UPDATE payment_notes
SET
    user_id = $2,
    green_note_id = $3,
    green_note_no = $4,
    green_note_approver = $5,
    green_note_app_date = $6,
    reimbursement_note_id = $7,
    subject = $8,
    date = $9,
    department = $10,
    vendor_code = $11,
    vendor_name = $12,
    project_name = $13,
    invoice_no = $14,
    invoice_date = $15,
    invoice_amount = $16,
    invoice_approved_by = $17,
    loa_po_no = $18,
    loa_po_amount = $19,
    loa_po_date = $20,
    gross_amount = $21,
    total_additions = $22,
    total_deductions = $23,
    net_payable_amount = $24,
    net_payable_round_off = $25,
    net_payable_words = $26,
    tds_percentage = $27,
    tds_section = $28,
    tds_amount = $29,
    account_holder_name = $30,
    bank_name = $31,
    account_number = $32,
    ifsc_code = $33,
    recommendation_of_payment = $34,
    status = $35,
    is_draft = $36,
    created_by_name = $37,
    created_by_designation = $38,
    created_by_role = $39,
    created_by_signature = $40,
    taxable_amount = $41,
    gst_amount = $42,
    other_charges = $43,
    green_note_history = $44,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentNoteStatus :one
UPDATE payment_notes
SET
    status = $2,
    is_draft = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeletePaymentNote :exec
DELETE FROM payment_notes
WHERE id = $1;

-- name: PutPaymentNoteOnHold :one
UPDATE payment_notes
SET
    status = 'H',
    hold_reason = $2,
    hold_date = NOW(),
    hold_by = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: RemovePaymentNoteFromHold :one
UPDATE payment_notes
SET
    status = $2,
    hold_reason = NULL,
    hold_date = NULL,
    hold_by = NULL,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentNoteUTR :one
UPDATE payment_notes
SET
    utr_no = $2,
    utr_date = $3,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: InsertPaymentParticular :one
INSERT INTO payment_note_particulars (
    payment_note_id,
    particular_type,
    particular,
    amount
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: ListPaymentParticulars :many
SELECT * FROM payment_note_particulars
WHERE payment_note_id = $1
ORDER BY id ASC;

-- name: DeletePaymentParticulars :exec
DELETE FROM payment_note_particulars
WHERE payment_note_id = $1;

-- name: InsertPaymentNoteComment :one
INSERT INTO payment_note_comments (
    payment_note_id,
    comment,
    status,
    user_id,
    user_name,
    user_email,
    user_designation,
    user_role,
    user_signature
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9
)
RETURNING *;

-- name: ListPaymentNoteComments :many
SELECT * FROM payment_note_comments
WHERE payment_note_id = $1
ORDER BY created_at DESC;

-- name: InsertApprovalLog :one
INSERT INTO payment_note_approval_logs (
    payment_note_id,
    status,
    comments,
    reviewer_id,
    reviewer_name,
    reviewer_email,
    approver_level,
    reviewer_designation,
    reviewer_role,
    reviewer_signature
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
)
RETURNING *;

-- name: ListApprovalLogs :many
SELECT * FROM payment_note_approval_logs
WHERE payment_note_id = $1
ORDER BY created_at DESC;

-- name: InsertPaymentNoteDocument :one
INSERT INTO payment_note_documents (
    payment_note_id,
    file_name,
    original_filename,
    mime_type,
    file_size,
    object_key,
    uploaded_by,
    uploaded_by_name
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8
)
RETURNING *;

-- name: ListPaymentNoteDocuments :many
SELECT * FROM payment_note_documents
WHERE payment_note_id = $1
ORDER BY created_at DESC;

-- name: DeletePaymentNoteDocument :exec
DELETE FROM payment_note_documents
WHERE id = $1;

-- name: GetNextPaymentNoteNumber :one
SELECT COALESCE(MAX(CAST(SUBSTRING(note_no FROM '[0-9]+$') AS INTEGER)), 0) + 1 AS next_number
FROM payment_notes
WHERE note_no LIKE $1 || '%';
