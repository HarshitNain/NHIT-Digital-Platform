-- name: CreateVendor :exec
INSERT INTO vendors (
    id, tenant_id, org_id, vendor_code, vendor_name, vendor_email, vendor_mobile,
    account_type, vendor_nick_name, activity_type, gstin, pan, pin,
    country_name, state_name, city_name,
    msme_classification, msme, msme_registration_number, msme_start_date, msme_end_date,
    material_nature, gst_defaulted, section_206ab_verified, beneficiary_name,
    remarks_address, common_bank_details, income_tax_type, project_id, status,
    from_account_type, account_name, short_name, parent,
    code_auto_generated, created_by, account_number, name_of_bank,
    ifsc_code, address, document_url
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16, $17, $18, $19, $20,
    $21, $22, $23, $24, $25, $26, $27, $28, $29, $30, $31, $32, $33, $34, $35, $36, $37, $38, $39,
    $40, $41
);

-- name: GetVendorByID :one
SELECT * FROM vendors 
WHERE id = $1 AND org_id = $2;

-- name: GetVendorByCode :one
SELECT * FROM vendors 
WHERE vendor_code = $1 AND org_id = $2;

-- name: GetVendorByEmail :one
SELECT * FROM vendors 
WHERE vendor_email = $1 AND org_id = $2;

-- name: UpdateVendor :exec
UPDATE vendors SET
    vendor_name = COALESCE($3, vendor_name),
    vendor_email = COALESCE($4, vendor_email),
    vendor_mobile = COALESCE($5, vendor_mobile),
    pan = COALESCE($6, pan),
    beneficiary_name = COALESCE($7, beneficiary_name),
    msme_classification = COALESCE($8, msme_classification),
    msme = COALESCE($9, msme),
    msme_registration_number = COALESCE($10, msme_registration_number),
    material_nature = COALESCE($11, material_nature),
    gst_defaulted = COALESCE($12, gst_defaulted),
    section_206ab_verified = COALESCE($13, section_206ab_verified),
    income_tax_type = COALESCE($14, income_tax_type),
    project_id = COALESCE($15, project_id),
    account_name = COALESCE($16, account_name),
    status = COALESCE($17, status),
    remarks_address = COALESCE($18, remarks_address),
    address = COALESCE($19, address),
    document_url = COALESCE($20, document_url),
    pin = COALESCE($21, pin),
    activity_type = COALESCE($22, activity_type),
    vendor_nick_name = COALESCE($23, vendor_nick_name),
    country_name = COALESCE($24, country_name),
    state_name = COALESCE($25, state_name),
    city_name = COALESCE($26, city_name),
    from_account_type = COALESCE($27, from_account_type),
    short_name = COALESCE($28, short_name),
    parent = COALESCE($29, parent),
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND org_id = $30;

-- name: UpdateVendorCode :exec
UPDATE vendors SET
    vendor_code = $3,
    code_auto_generated = $4,
    updated_at = NOW()
WHERE id = $1 AND tenant_id = $2 AND org_id = $5;

-- name: DeleteVendor :exec
DELETE FROM vendors 
WHERE id = $1 AND tenant_id = $2 AND org_id = $3;

-- name: ListVendors :many
SELECT * FROM vendors 
WHERE tenant_id = $1 AND ($8::uuid IS NULL OR org_id = $8)
    AND ($2::text IS NULL OR status::text = $2)
    AND ($3::text IS NULL OR account_type::text = $3)
    AND ($4::text IS NULL OR project_id = $4)
    AND ($5::text IS NULL OR (
        vendor_name ILIKE '%' || $5 || '%' OR
        vendor_email ILIKE '%' || $5 || '%' OR
        vendor_code ILIKE '%' || $5 || '%'
    ))
ORDER BY created_at DESC
LIMIT $6 OFFSET $7;

-- name: CountVendors :one
SELECT COUNT(*) FROM vendors 
WHERE tenant_id = $1 AND ($6::uuid IS NULL OR org_id = $6)
    AND ($2::text IS NULL OR status::text = $2)
    AND ($3::text IS NULL OR account_type::text = $3)
    AND ($4::text IS NULL OR project_id = $4)
    AND ($5::text IS NULL OR (
        vendor_name ILIKE '%' || $5 || '%' OR
        vendor_email ILIKE '%' || $5 || '%' OR
        vendor_code ILIKE '%' || $5 || '%'
    ));

-- name: IsVendorCodeExists :one
SELECT EXISTS(
    SELECT 1 FROM vendors 
    WHERE vendor_code = $1 AND org_id = $2
    AND ($3::uuid IS NULL OR id != $3)
);

-- name: IsVendorEmailExists :one
SELECT EXISTS(
    SELECT 1 FROM vendors 
    WHERE vendor_email = $1 AND org_id = $2
    AND ($3::uuid IS NULL OR id != $3)
);
