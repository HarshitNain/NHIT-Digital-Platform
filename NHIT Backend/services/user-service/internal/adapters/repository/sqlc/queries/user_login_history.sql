-- name: CreateLoginHistory :one
INSERT INTO user_login_history (user_id, ip_address, user_agent, login_time)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetLoginHistory :one
SELECT * FROM user_login_history
WHERE history_id = $1;

-- name: ListUserLoginHistories :many
SELECT * FROM user_login_history
WHERE user_id = $1
ORDER BY login_time DESC
LIMIT $2 OFFSET $3;

-- name: ListRecentLoginHistories :many
SELECT * FROM user_login_history
WHERE user_id = $1
ORDER BY login_time DESC
LIMIT $2;

-- name: CountUserLoginHistories :one
SELECT COUNT(*) FROM user_login_history
WHERE user_id = $1;

-- name: DeleteLoginHistory :exec
DELETE FROM user_login_history
WHERE history_id = $1;

-- name: DeleteUserLoginHistories :exec
DELETE FROM user_login_history
WHERE user_id = $1;

-- name: DeleteOldLoginHistories :exec
DELETE FROM user_login_history
WHERE login_time < $1;
