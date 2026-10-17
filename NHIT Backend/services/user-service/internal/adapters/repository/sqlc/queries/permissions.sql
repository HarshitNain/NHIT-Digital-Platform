-- name: ListPermissions :many
SELECT * FROM permissions
ORDER BY module NULLS LAST, name;

-- name: ListPermissionsByModule :many
SELECT * FROM permissions
WHERE module = $1
ORDER BY name;
