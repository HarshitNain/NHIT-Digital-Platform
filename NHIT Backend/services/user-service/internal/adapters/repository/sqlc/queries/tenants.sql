-- name: CreateTenant :one
INSERT INTO tenants (tenant_id, name, email, password)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetTenant :one
SELECT *
FROM tenants
WHERE tenant_id = $1;
