-- name: AssignRoleToUser :exec
INSERT INTO user_roles (user_id, role_id)
VALUES ($1, $2)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- name: RemoveRoleFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1 AND role_id = $2;

-- name: RemoveAllRolesFromUser :exec
DELETE FROM user_roles
WHERE user_id = $1;

-- name: ListRolesForUser :many
SELECT role_id FROM user_roles
WHERE user_id = $1;

-- name: ListDetailedRolesForUser :many
SELECT r.*
FROM user_roles ur
JOIN roles r ON r.role_id = ur.role_id
WHERE ur.user_id = $1
ORDER BY r.created_at DESC;

-- name: ListUsersForRole :many
SELECT user_id FROM user_roles
WHERE role_id = $1;

-- name: UserHasRole :one
SELECT EXISTS(
    SELECT 1 FROM user_roles
    WHERE user_id = $1 AND role_id = $2
) AS has_role;

-- name: CountRolesForUser :one
SELECT COUNT(*) FROM user_roles
WHERE user_id = $1;

-- name: CountUsersForRole :one
SELECT COUNT(*) FROM user_roles
WHERE role_id = $1;
