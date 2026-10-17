-- name: CreateVendorAccount :exec
INSERT INTO vendor_accounts (
    id, vendor_id, account_name, account_number, account_type,
    name_of_bank, branch_name, ifsc_code, swift_code,
    is_primary, is_active, remarks, created_by
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13
);

-- name: GetVendorAccountByID :one
SELECT * FROM vendor_accounts 
WHERE id = $1;

-- name: GetVendorAccountsByVendorID :many
SELECT * FROM vendor_accounts 
WHERE vendor_id = $1
ORDER BY is_primary DESC, created_at ASC;

-- name: GetPrimaryVendorAccount :one
SELECT * FROM vendor_accounts 
WHERE vendor_id = $1 AND is_primary = true AND is_active = true
LIMIT 1;

-- name: UpdateVendorAccount :exec
UPDATE vendor_accounts SET
    account_name = COALESCE($2, account_name),
    account_number = COALESCE($3, account_number),
    account_type = COALESCE($4, account_type),
    name_of_bank = COALESCE($5, name_of_bank),
    branch_name = COALESCE($6, branch_name),
    ifsc_code = COALESCE($7, ifsc_code),
    swift_code = COALESCE($8, swift_code),
    is_primary = COALESCE($9, is_primary),
    is_active = COALESCE($10, is_active),
    remarks = COALESCE($11, remarks),
    updated_at = NOW()
WHERE id = $1;

-- name: DeleteVendorAccount :exec
DELETE FROM vendor_accounts 
WHERE id = $1;

-- name: UnsetPrimaryVendorAccounts :exec
UPDATE vendor_accounts SET
    is_primary = false,
    updated_at = NOW()
WHERE vendor_id = $1 
    AND ($2::uuid IS NULL OR id != $2)
    AND is_primary = true;

-- name: ToggleVendorAccountStatus :exec
UPDATE vendor_accounts SET
    is_active = NOT is_active,
    updated_at = NOW()
WHERE id = $1;

-- name: SetAccountAsPrimary :exec
UPDATE vendor_accounts SET
    is_primary = true,
    updated_at = NOW()
WHERE id = $1;
