-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    name,
    description,
    created_at
) VALUES (
    $1, $2, NOW()
) RETURNING *;

-- name: ListActivityLogs :many
SELECT * FROM activity_logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountActivityLogs :one
SELECT COUNT(*) FROM activity_logs;

-- name: DeleteActivityLog :exec
DELETE FROM activity_logs WHERE id = $1;
