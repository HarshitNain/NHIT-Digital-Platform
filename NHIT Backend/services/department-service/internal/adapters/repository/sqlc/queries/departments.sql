-- name: CreateDepartment :one
INSERT INTO departments (name, description, org_id)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetDepartmentByID :one
SELECT * FROM departments WHERE id = $1;

-- name: GetDepartmentByName :one
SELECT * FROM departments WHERE lower(name) = lower($1) AND org_id = $2;

-- name: UpdateDepartment :one
UPDATE departments
SET name = $2, description = $3, updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteDepartment :exec
DELETE FROM departments WHERE id = $1;

-- name: ListDepartments :many
SELECT * FROM departments 
WHERE ($1::uuid IS NULL OR org_id = $1)
ORDER BY created_at DESC 
LIMIT $2 OFFSET $3;

-- name: CountDepartments :one
SELECT COUNT(*) FROM departments
WHERE ($1::uuid IS NULL OR org_id = $1);

-- name: DepartmentExists :one
SELECT EXISTS(SELECT 1 FROM departments WHERE lower(name) = lower($1) AND org_id = $2) AS exists;

-- name: DepartmentExistsByID :one
SELECT EXISTS(SELECT 1 FROM departments WHERE id = $1) AS exists;
