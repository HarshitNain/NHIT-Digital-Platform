-- name: CreateOrganization :one
INSERT INTO organizations (
    org_id, tenant_id, parent_org_id, name, code, database_name,
    description, logo,
    super_admin_name, super_admin_email, super_admin_password,
    initial_projects, status
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8,
    $9, $10, $11,
    $12, $13
)
RETURNING *;

-- name: GetOrganizationByID :one
SELECT * FROM organizations WHERE org_id = $1;

-- name: GetOrganizationByCode :one
SELECT * FROM organizations WHERE code = $1;

-- name: ListOrganizations :many
SELECT * FROM organizations ORDER BY created_at DESC OFFSET $1 LIMIT $2;

-- name: ListOrganizationsByTenant :many
SELECT * FROM organizations WHERE tenant_id = $1 ORDER BY created_at DESC OFFSET $2 LIMIT $3;

-- name: ListChildOrganizations :many
SELECT * FROM organizations WHERE parent_org_id = $1 ORDER BY created_at DESC OFFSET $2 LIMIT $3;

-- name: CountOrganizations :one
SELECT COUNT(*) FROM organizations;

-- name: CountOrganizationsByTenant :one
SELECT COUNT(*) FROM organizations WHERE tenant_id = $1;

-- name: CountChildOrganizations :one
SELECT COUNT(*) FROM organizations WHERE parent_org_id = $1;

-- name: UpdateOrganization :one
UPDATE organizations SET
    name = $2,
    code = $3,
    description = $4,
    logo = $5,
    status = $6,
    updated_at = NOW()
WHERE org_id = $1
RETURNING *;

-- name: DeleteOrganization :exec
DELETE FROM organizations WHERE org_id = $1;
