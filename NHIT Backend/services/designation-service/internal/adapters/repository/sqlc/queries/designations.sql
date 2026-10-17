-- name: CreateDesignation :one
INSERT INTO designations (id, name, description, created_at, updated_at, org_id)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetDesignationByID :one
SELECT * FROM designations WHERE id = $1;

-- name: UpdateDesignation :one
UPDATE designations
SET name = $2, description = $3, updated_at = $4
WHERE id = $1
RETURNING *;

-- name: DeleteDesignation :exec
DELETE FROM designations WHERE id = $1;

-- name: ListDesignations :many
SELECT *
FROM designations
WHERE ($1::uuid IS NULL OR org_id = $1)
ORDER BY name ASC
LIMIT $2 OFFSET $3;

-- name: CountDesignations :one
SELECT COUNT(*) FROM designations
WHERE ($1::uuid IS NULL OR org_id = $1);

-- name: DesignationExists :one
SELECT EXISTS(SELECT 1 FROM designations WHERE lower(name) = lower($1) AND org_id = $2) AS exists;

-- name: GetDesignationByName :one
SELECT * FROM designations WHERE lower(name) = lower($1) AND org_id = $2;
