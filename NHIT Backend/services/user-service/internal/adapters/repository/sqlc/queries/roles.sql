-- name: CreateRole :one
INSERT INTO roles (tenant_id, parent_org_id, name, description, permissions, is_system_role, created_by)
VALUES (
    sqlc.arg('tenant_id')::uuid, 
    sqlc.narg('org_id')::uuid, 
    sqlc.arg('name')::text, 
    sqlc.arg('description')::text, 
    sqlc.arg('permissions')::text[], 
    sqlc.arg('is_system_role')::boolean, 
    sqlc.arg('created_by')::text
)
RETURNING *;

-- name: GetRole :one
SELECT * FROM roles
WHERE role_id = sqlc.arg('role_id')::uuid;

-- name: ListRolesByTenant :many
SELECT * FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountRolesByTenant :one
SELECT COUNT(*) FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid;

-- name: ListRolesByTenantAndOrg :many
SELECT * FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND parent_org_id = sqlc.narg('org_id')::uuid
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountRolesByTenantAndOrg :one
SELECT COUNT(*) FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND parent_org_id = sqlc.narg('org_id')::uuid;

-- name: ListRolesByOrganizationIncludingSystem :many
SELECT * FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (parent_org_id = sqlc.narg('org_id')::uuid OR (is_system_role = TRUE AND parent_org_id IS NULL))
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountRolesByOrganizationIncludingSystem :one
SELECT COUNT(*) FROM roles
WHERE tenant_id = sqlc.arg('tenant_id')::uuid
  AND (parent_org_id = sqlc.narg('org_id')::uuid OR (is_system_role = TRUE AND parent_org_id IS NULL));

-- name: UpdateRole :one
UPDATE roles
SET name = sqlc.arg('name')::text,
    description = sqlc.arg('description')::text,
    permissions = sqlc.arg('permissions')::text[],
    updated_at = NOW()
WHERE role_id = sqlc.arg('role_id')::uuid
RETURNING *;

-- name: DeleteRole :exec
DELETE FROM roles
WHERE role_id = sqlc.arg('role_id')::uuid;

-- name: ListRolesByIDs :many
SELECT * FROM roles
WHERE role_id = ANY($1::uuid[])
ORDER BY created_at DESC;

-- name: RoleExists :one
SELECT EXISTS(SELECT 1 FROM roles WHERE tenant_id = sqlc.arg('tenant_id')::uuid AND parent_org_id = sqlc.narg('org_id')::uuid AND lower(name) = lower(sqlc.arg('name')::text)) AS exists;

-- name: RoleExistsSystem :one
SELECT EXISTS(SELECT 1 FROM roles WHERE tenant_id = sqlc.arg('tenant_id')::uuid AND is_system_role = TRUE AND lower(name) = lower(sqlc.arg('name')::text)) AS exists;
