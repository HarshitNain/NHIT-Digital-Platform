-- name: CreatePayment :one
INSERT INTO payments (
    sl_no,
    template_type,
    project,
    account_full_name,
    from_account_type,
    full_account_number,
    to_account,
    to_account_type,
    name_of_beneficiary,
    account_number,
    name_of_bank,
    ifsc_code,
    amount,
    purpose,
    status,
    user_id,
    payment_note_id
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17
)
RETURNING *;

-- name: GetPaymentsBySlNo :many
SELECT *
FROM payments
WHERE sl_no = $1
ORDER BY id ASC;

-- name: GetPaymentByID :one
SELECT *
FROM payments
WHERE id = $1
LIMIT 1;

-- name: ListPayments :many
SELECT DISTINCT sl_no, MAX(created_at) as latest_created_at, MAX(status::text) as status
FROM payments
WHERE 
    ($1::text IS NULL OR status::text = $1) AND
    ($2::boolean IS NULL OR (user_id = $3))
GROUP BY sl_no
ORDER BY latest_created_at DESC
LIMIT $4 OFFSET $5;

-- name: CountPaymentGroups :one
SELECT COUNT(DISTINCT sl_no)
FROM payments
WHERE 
    ($1::text IS NULL OR status::text = $1) AND
    ($2::boolean IS NULL OR (user_id = $3));

-- name: UpdatePayment :one
UPDATE payments
SET
    template_type = $2,
    project = $3,
    account_full_name = $4,
    from_account_type = $5,
    full_account_number = $6,
    to_account = $7,
    to_account_type = $8,
    name_of_beneficiary = $9,
    account_number = $10,
    name_of_bank = $11,
    ifsc_code = $12,
    amount = $13,
    purpose = $14,
    status = $15,
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: UpdatePaymentStatus :exec
UPDATE payments
SET
    status = $2,
    updated_at = NOW()
WHERE sl_no = $1;

-- name: DeletePayment :exec
DELETE FROM payments
WHERE id = $1;

-- name: DeletePaymentsBySlNo :exec
DELETE FROM payments
WHERE sl_no = $1;

-- name: InsertPaymentVendorAccount :one
INSERT INTO payment_vendor_accounts (
    payment_id,
    vendor_id,
    vendor_account_id
) VALUES (
    $1, $2, $3
)
ON CONFLICT (payment_id, vendor_id) DO NOTHING
RETURNING *;

-- name: GetPaymentVendorAccounts :many
SELECT *
FROM payment_vendor_accounts
WHERE payment_id = $1;

-- name: CreatePaymentShortcut :one
INSERT INTO payment_shortcuts (
    sl_no,
    shortcut_name,
    request_data_json,
    user_id
) VALUES (
    $1, $2, $3, $4
)
RETURNING *;

-- name: GetPaymentShortcut :one
SELECT *
FROM payment_shortcuts
WHERE id = $1
LIMIT 1;

-- name: ListPaymentShortcuts :many
SELECT *
FROM payment_shortcuts
WHERE user_id = $1
ORDER BY created_at DESC;

-- name: GenerateSerialNumber :one
SELECT COALESCE(MAX(CAST(SUBSTRING(sl_no FROM '[0-9]+$') AS INTEGER)), 0) + 1 AS next_number
FROM payments
WHERE sl_no LIKE $1 || '%';

-- name: InsertBankLetterLog :one
INSERT INTO bank_letter_approval_logs (
    sl_no,
    status,
    comments,
    reviewer_id,
    reviewer_name,
    reviewer_email,
    approver_level
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
)
RETURNING *;

-- name: ListBankLetterLogs :many
SELECT *
FROM bank_letter_approval_logs
WHERE sl_no = $1
ORDER BY created_at DESC;
