-- name: CreateUser :one
INSERT INTO users (
    tenant_id, org_id, name, email, password, department_id, designation_id,
    account_holder_name, bank_name, bank_account_number, ifsc_code, signature_url, is_active
)
VALUES (
    sqlc.arg('tenant_id')::uuid, 
    sqlc.narg('org_id')::uuid, 
    sqlc.arg('name')::text, 
    sqlc.arg('email')::text, 
    sqlc.arg('password')::text, 
    sqlc.arg('department_id')::uuid, 
    sqlc.arg('designation_id')::uuid, 
    sqlc.narg('account_holder_name')::text, 
    sqlc.narg('bank_name')::text, 
    sqlc.narg('bank_account_number')::text, 
    sqlc.narg('ifsc_code')::text, 
    sqlc.narg('signature_url')::text, 
    sqlc.arg('is_active')::boolean
)
RETURNING *;

-- name: CreateUserWithVerification :one
INSERT INTO users (
    tenant_id, org_id, name, email, password, email_verified_at, department_id, designation_id,
    account_holder_name, bank_name, bank_account_number, ifsc_code, signature_url, is_active
)
VALUES (
    sqlc.arg('tenant_id')::uuid, 
    sqlc.narg('org_id')::uuid, 
    sqlc.arg('name')::text, 
    sqlc.arg('email')::text, 
    sqlc.arg('password')::text, 
    sqlc.arg('email_verified_at')::timestamptz, 
    sqlc.arg('department_id')::uuid, 
    sqlc.arg('designation_id')::uuid, 
    sqlc.narg('account_holder_name')::text, 
    sqlc.narg('bank_name')::text, 
    sqlc.narg('bank_account_number')::text, 
    sqlc.narg('ifsc_code')::text, 
    sqlc.narg('signature_url')::text, 
    sqlc.arg('is_active')::boolean
)
RETURNING *;

-- name: GetUser :one
SELECT * FROM users
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL);

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE email = sqlc.arg('email')::text 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL);

-- name: GetUserByEmailAndTenant :one
SELECT * FROM users
WHERE tenant_id = sqlc.arg('tenant_id')::uuid 
  AND email = sqlc.arg('email')::text 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL);

-- name: UpdateUser :one
UPDATE users
SET 
    name = sqlc.arg('name')::text, 
    email = sqlc.arg('email')::text, 
    password = sqlc.arg('password')::text, 
    department_id = sqlc.arg('department_id')::uuid,
    designation_id = sqlc.arg('designation_id')::uuid,
    account_holder_name = sqlc.narg('account_holder_name')::text,
    bank_name = sqlc.narg('bank_name')::text,
    bank_account_number = sqlc.narg('bank_account_number')::text,
    ifsc_code = sqlc.narg('ifsc_code')::text,
    signature_url = sqlc.narg('signature_url')::text,
    is_active = sqlc.arg('is_active')::boolean,
    deactivated_at = sqlc.narg('deactivated_at')::timestamptz,
    deactivated_by = sqlc.narg('deactivated_by')::uuid,
    deactivated_by_name = sqlc.narg('deactivated_by_name')::text,
    updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserProfile :one
UPDATE users
SET name = sqlc.arg('name')::text, email = sqlc.arg('email')::text, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password = sqlc.arg('password')::text, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserEmailVerification :one
UPDATE users
SET email_verified_at = sqlc.arg('email_verified_at')::timestamptz, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserLastLogin :one
UPDATE users
SET last_login_at = sqlc.arg('last_login_at')::timestamptz, last_login_ip = sqlc.narg('last_login_ip')::text, user_agent = sqlc.narg('user_agent')::text, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserLastLogout :one
UPDATE users
SET last_logout_at = sqlc.arg('last_logout_at')::timestamptz, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserDepartment :one
UPDATE users
SET department_id = sqlc.arg('department_id')::uuid, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: UpdateUserDesignation :one
UPDATE users
SET designation_id = sqlc.arg('designation_id')::uuid, updated_at = NOW()
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
RETURNING *;

-- name: DeleteUser :exec
DELETE FROM users
WHERE user_id = sqlc.arg('user_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL);

-- name: ListUsersByTenant :many
SELECT * FROM users
WHERE tenant_id = sqlc.arg('tenant_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: ListAllUsers :many
SELECT * FROM users
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: CountUsersByTenant :one
SELECT COUNT(*) FROM users
WHERE tenant_id = sqlc.arg('tenant_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL);

-- name: CountAllUsers :one
SELECT COUNT(*) FROM users;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = sqlc.arg('email')::text) AS exists;

-- name: UserExistsByID :one
SELECT EXISTS(SELECT 1 FROM users WHERE user_id = sqlc.arg('user_id')::uuid) AS exists;

-- name: ListUsersByDepartment :many
SELECT * FROM users
WHERE department_id = sqlc.arg('department_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: ListUsersByDesignation :many
SELECT * FROM users
WHERE designation_id = sqlc.arg('designation_id')::uuid 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
ORDER BY created_at DESC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: SearchUsersByName :many
SELECT * FROM users
WHERE LOWER(name) LIKE LOWER('%' || sqlc.arg('name')::text || '%') 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
ORDER BY name ASC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;

-- name: SearchUsersByEmail :many
SELECT * FROM users
WHERE LOWER(email) LIKE LOWER('%' || sqlc.arg('email')::text || '%') 
  AND (sqlc.narg('org_id')::uuid IS NULL OR org_id = sqlc.narg('org_id') OR org_id IS NULL)
ORDER BY email ASC
LIMIT sqlc.arg('limit')::int OFFSET sqlc.arg('offset')::int;
